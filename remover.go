// Copyright (c) 2019-2023 Vincent Cheung (coolingfall@gmail.com).
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lork

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Bound the first scan even for second- and minute-based archive patterns.
const maxInitialCleanupPeriods = 14 * 24

// ArchiveRemover represents a remover which removes archived logs.
type ArchiveRemover interface {
	// MaxHistory sets max history for logs.
	MaxHistory(max int)

	// CleanAsync cleans logs asynchronously with given time.
	CleanAsync(now time.Time)
}

type archiveFileLister interface {
	listFilesInPeriod(time.Time) []string
}

type timeBasedArchiveRemover struct {
	fp          *filenamePattern
	rd          *rollingDate
	maxHistory  int
	parentClean bool

	locker    sync.Mutex
	lastClean time.Time
}

func newTimeBasedArchiveRemover(fp *filenamePattern, rd *rollingDate) ArchiveRemover {
	tbar := &timeBasedArchiveRemover{
		fp: fp,
		rd: rd,
	}

	tbar.parentClean = tbar.calcParentCleanFlag()

	return tbar
}

func (r *timeBasedArchiveRemover) MaxHistory(max int) {
	r.locker.Lock()
	defer r.locker.Unlock()
	r.maxHistory = max
}

func (r *timeBasedArchiveRemover) CleanAsync(now time.Time) {
	r.cleanAsync(now, r)
}

func (r *timeBasedArchiveRemover) cleanAsync(now time.Time,
	lister archiveFileLister) {
	r.locker.Lock()
	if r.maxHistory <= 0 || !r.rd.cleanupSupported {
		r.locker.Unlock()
		return
	}
	periodsElapsed := r.calcElapsedPeriods(now)
	if periodsElapsed <= 0 {
		r.locker.Unlock()
		return
	}
	// Reserve this range before launching a worker. Same-period calls neither
	// advance the cursor nor allocate another goroutine.
	r.lastClean = now
	maxHistory := r.maxHistory
	r.locker.Unlock()

	go r.cleanPeriods(now, periodsElapsed, maxHistory, lister)
}

func (r *timeBasedArchiveRemover) cleanPeriods(now time.Time, periodsElapsed, maxHistory int,
	lister archiveFileLister) {
	// Repeating or incomplete date labels cannot distinguish expired files
	// from retained files. Leave them untouched rather than risk data loss.
	if !r.rd.cleanupSupported {
		return
	}
	for i := 0; i < periodsElapsed; i++ {
		offset := -maxHistory - i
		timeOfPeriodToClean, exists := r.rd.periodForCleanup(now, offset)
		if !exists {
			continue
		}
		files := lister.listFilesInPeriod(timeOfPeriodToClean)
		if len(files) == 0 {
			continue
		}

		// remove all log files
		for _, fn := range files {
			_ = os.Remove(fn)
		}

		// remove parent directory
		if r.parentClean {
			dir := filepath.Dir(files[0])
			_ = os.Remove(dir)
		}
	}
}

func (r *timeBasedArchiveRemover) calcElapsedPeriods(now time.Time) int {
	if r.lastClean.IsZero() {
		return min(r.rd.periodCrossed(now, now.AddDate(0, 0, 32)), maxInitialCleanupPeriods)
	}
	return r.rd.periodCrossed(r.lastClean, now)
}

func (r *timeBasedArchiveRemover) calcParentCleanFlag() bool {
	// if date pattern contains /, the parent should be cleaned
	if strings.Contains(r.fp.datePattern(), "/") {
		return true
	}

	// if literal string after dtc contains /, the parent should be removed
	c := r.fp.headConverter()
	for ; c != nil; c = c.Next() {
		if _, ok := c.(*dateConverter); ok {
			c = c.Next()
			break
		}
	}

	var buf = new(bytes.Buffer)
	for ; c != nil; c = c.Next() {
		if _, ok := c.(*literalConverter); ok {
			c.Convert(nil, buf)
			s := buf.String()
			if strings.Contains(s, "/") {
				return true
			}
		}
	}

	return false
}

func (r *timeBasedArchiveRemover) listFilesInPeriod(t time.Time) []string {
	matchingFiles := make([]string, 0)

	filename := r.fp.convert(t, -1)
	matchingFiles = append(matchingFiles, filename)
	if originalFilename := uncompressedArchiveFilename(filename); originalFilename != filename {
		matchingFiles = append(matchingFiles, originalFilename)
	}

	return matchingFiles
}

type sizeAndTimeArchiveRemover struct {
	*timeBasedArchiveRemover
}

func newSizeAndTimeArchiveRemover(fp *filenamePattern, rd *rollingDate) ArchiveRemover {
	tbar := newTimeBasedArchiveRemover(fp, rd)
	return &sizeAndTimeArchiveRemover{
		timeBasedArchiveRemover: tbar.(*timeBasedArchiveRemover),
	}
}

func (r *sizeAndTimeArchiveRemover) CleanAsync(now time.Time) {
	r.timeBasedArchiveRemover.cleanAsync(now, r)
}

func (r *sizeAndTimeArchiveRemover) listFilesInPeriod(t time.Time) []string {
	matchingFiles := make([]string, 0)

	filename := r.fp.convert(t, 0)
	dir := filepath.Dir(filename)
	reg := r.fp.toArchiveFilenameRegexForFixed(t)
	reg = afterLastSlash(reg)
	filenameRegex, err := regexp.Compile("^" + reg + "$")
	if err != nil {
		return matchingFiles
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return matchingFiles
	}

	for _, file := range files {
		if file.IsDir() || !filenameRegex.MatchString(file.Name()) {
			continue
		}

		fn, err := filepath.Abs(filepath.Join(dir, file.Name()))
		if err != nil {
			continue
		}
		matchingFiles = append(matchingFiles, fn)
	}

	return matchingFiles
}
