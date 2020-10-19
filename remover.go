// Copyright (c) 2019-2020 Anbillon Team (anbillonteam@gmail.com).
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

package slago

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ArchiveRemover represents a remover which removes archived logs.
type ArchiveRemover interface {
	// MaxHistory sets max history for logs.
	MaxHistory(max int)

	// CleanAsync cleans logs asynchronously with given time.
	CleanAsync(now time.Time)
}

type timeBasedArchiveRemover struct {
	fp          *filenamePattern
	rd          *rollingDate
	maxHistory  int
	parentClean bool

	lastClean int64
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
	r.maxHistory = max
}

func (r *timeBasedArchiveRemover) CleanAsync(now time.Time) {
	r.cleanAsync(now, r.listFilesInPeriod)
}

func (r *timeBasedArchiveRemover) cleanAsync(now time.Time,
	listFilesInPeriod func(t time.Time) []string) {
	go func() {
		periodsElapsed := r.calcElapsedPeriods(now)
		r.lastClean = time.Now().Unix()
		for i := 0; i < periodsElapsed; i++ {
			offset := -r.maxHistory - i
			timeOfPeriodToClean := r.rd.endOfNextNPeriod(now, offset)
			files := listFilesInPeriod(timeOfPeriodToClean)
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
	}()
}

func (r *timeBasedArchiveRemover) calcElapsedPeriods(now time.Time) int {
	var periodsElapsed = 0
	nowUnix := now.Unix()
	if r.lastClean == 0 {
		periodsElapsed = r.rd.periodCrossed(nowUnix, nowUnix+32*secondsInOneDay)
	} else {
		periodsElapsed = r.rd.periodCrossed(nowUnix, r.lastClean)
	}

	return periodsElapsed
}

func (r *timeBasedArchiveRemover) calcParentCleanFlag() bool {
	// if date pattern contains /, the parent should be cleaned
	if strings.Contains(r.fp.datePattern(), "/") {
		return true
	}

	// if literal string subsequent to dtc contains /, the parent should be removed
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
	r.timeBasedArchiveRemover.cleanAsync(now, r.listFilesInPeriod)
}

func (r *sizeAndTimeArchiveRemover) listFilesInPeriod(t time.Time) []string {
	matchingFiles := make([]string, 0)

	filename := r.fp.convert(t, 0)
	dir := filepath.Dir(filename)
	reg := r.fp.toFilenameRegexForFixed(t)
	reg = afterLastSlash(reg)
	filenameRegex, err := regexp.Compile(reg)
	if err != nil {
		return matchingFiles
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return matchingFiles
	}

	for _, file := range files {
		if !filenameRegex.MatchString(file.Name()) {
			continue
		}

		fn, err := filepath.Abs(file.Name())
		if err != nil {
			continue
		}
		matchingFiles = append(matchingFiles, fn)
	}

	return matchingFiles
}
