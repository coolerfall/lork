// Copyright (c) 2019-2022 Vincent Cheung (coolingfall@gmail.com).
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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

// RollingPolicy represents policy for log rolling.
type RollingPolicy interface {
	// Prepare prepares current rolling policy
	Prepare() error

	// Attach attaches file writer.
	Attach(w *fileWriter)

	// ShouldTrigger check if there's necessary to trigger rolling.
	ShouldTrigger(fileSize int64) bool

	// Rotate does the log rolling.
	Rotate() error
}

type noopRollingPolicy struct {
}

// NewNoopRollingPolicy creates a new instance of noop rolling policy which will do nothing.
func NewNoopRollingPolicy() RollingPolicy {
	return &noopRollingPolicy{}
}

func (rp *noopRollingPolicy) Prepare() error {
	return nil
}

func (rp *noopRollingPolicy) Attach(_ *fileWriter) {
}

func (rp *noopRollingPolicy) ShouldTrigger(_ int64) bool {
	return false
}

func (rp *noopRollingPolicy) Rotate() error {
	return nil
}

type timeBasedRollingPolicy struct {
	fileWriter          *fileWriter
	nextCheck           time.Time
	timeInCurrentPeriod time.Time

	archiver       Archiver
	archiveRemover ArchiveRemover
	maxHistory     int

	filenamePattern *filenamePattern
	rollingDate     *rollingDate
}

// TimeBasedRPOption represents available options for size and time
// based rolling policy.
type TimeBasedRPOption struct {
	FilenamePattern string
	MaxHistory      int
}

// NewTimeBasedRollingPolicy creates an instance of time based rolling policy
// for file writer.
func NewTimeBasedRollingPolicy(options ...func(*TimeBasedRPOption)) RollingPolicy {
	opt := &TimeBasedRPOption{
		FilenamePattern: "lork-archive.#date{2006-01-02}.log",
	}

	for _, f := range options {
		f(opt)
	}

	fp, err := newFilenamePattern(opt.FilenamePattern)
	if err != nil {
		ReportfExit("create rolling policy error: %v\n", err)
		return nil
	}

	return &timeBasedRollingPolicy{
		maxHistory:      opt.MaxHistory,
		filenamePattern: fp,
		archiver:        newArchiver(opt.FilenamePattern),
	}
}

func (rp *timeBasedRollingPolicy) Prepare() error {
	if rp.filenamePattern.hasIndexConverter() {
		return errors.New("invalid filename pattern, contains index pattern")
	}

	return rp.prepare(newTimeBasedArchiveRemover)
}

func (rp *timeBasedRollingPolicy) Attach(w *fileWriter) {
	rp.fileWriter = w
}

func (rp *timeBasedRollingPolicy) ShouldTrigger(_ int64) bool {
	if time.Now().After(rp.nextCheck) {
		rp.calcNextCheck()
		return true
	}

	return false
}

func (rp *timeBasedRollingPolicy) Rotate() error {
	rollingFilename := rp.filenamePattern.convert(rp.timeInCurrentPeriod, 0)
	if len(rollingFilename) == 0 {
		rollingFilename = fmt.Sprintf("lork-%s.log",
			rp.timeInCurrentPeriod.Format("2006-01-02"))
	}

	rp.timeInCurrentPeriod = time.Now()

	err := rp.archiver.Archive(rp.fileWriter.RawFilename(), rollingFilename)

	if rp.archiveRemover != nil {
		rp.archiveRemover.CleanAsync(time.Now())
	}

	return err
}

func (rp *timeBasedRollingPolicy) prepare(
	newArchiveRemover func(*filenamePattern, *rollingDate) ArchiveRemover) error {
	datePattern := rp.filenamePattern.datePattern()
	if len(datePattern) == 0 {
		return errors.New("invalid filename pattern, missing date pattern")
	}

	rp.rollingDate = newRollingDate(datePattern)
	rp.timeInCurrentPeriod = time.Now()
	// check latest modification time if file existed
	if info, err := os.Stat(rp.fileWriter.RawFilename()); err == nil {
		rp.timeInCurrentPeriod = info.ModTime()
	}
	rp.initNextCheck()

	if rp.maxHistory != 0 {
		rp.archiveRemover = newArchiveRemover(rp.filenamePattern, rp.rollingDate)
		rp.archiveRemover.MaxHistory(rp.maxHistory)
		rp.archiveRemover.CleanAsync(time.Now())
	}

	return nil
}

func (rp *timeBasedRollingPolicy) calcNextCheck() {
	rp.nextCheck = rp.rollingDate.next(time.Now())
}

func (rp *timeBasedRollingPolicy) initNextCheck() {
	rp.nextCheck = rp.rollingDate.next(rp.timeInCurrentPeriod)
}

type sizeAndTimeBasedRollingPolicy struct {
	*timeBasedRollingPolicy

	triggerSize int64
	index       int
}

// SizeAndTimeBasedRPOption represents available options for size and time
// based rolling policy.
type SizeAndTimeBasedRPOption struct {
	FilenamePattern string
	MaxFileSize     string
	MaxHistory      int
}

// NewSizeAndTimeBasedRollingPolicy creates a new instance of size and time
// based rolling policy for file writer.
func NewSizeAndTimeBasedRollingPolicy(options ...func(
	*SizeAndTimeBasedRPOption)) RollingPolicy {
	opt := &SizeAndTimeBasedRPOption{
		MaxFileSize:     "128MB",
		FilenamePattern: "lork-archive.#date{2006-01-02}.#index.log",
	}

	for _, f := range options {
		f(opt)
	}

	fileSize, err := parseFileSize(opt.MaxFileSize)
	if err != nil {
		ReportfExit("parse file size error: %v", err)
	}

	tbrp := NewTimeBasedRollingPolicy(func(o *TimeBasedRPOption) {
		o.FilenamePattern = opt.FilenamePattern
		o.MaxHistory = opt.MaxHistory
	}).(*timeBasedRollingPolicy)
	return &sizeAndTimeBasedRollingPolicy{
		timeBasedRollingPolicy: tbrp,
		triggerSize:            fileSize,
	}
}

func (rp *sizeAndTimeBasedRollingPolicy) Prepare() error {
	if rp.fileWriter == nil {
		return errors.New("rolling policy is not attached to a file writer")
	}

	if err := rp.prepare(newSizeAndTimeArchiveRemover); err != nil {
		return err
	}

	return rp.calcIndex()
}

func (rp *sizeAndTimeBasedRollingPolicy) ShouldTrigger(fileSize int64) bool {
	if rp.timeBasedRollingPolicy.ShouldTrigger(fileSize) {
		rp.index = 0
		return true
	}

	if fileSize >= rp.triggerSize {
		return true
	}

	return false
}

func (rp *sizeAndTimeBasedRollingPolicy) Rotate() (err error) {
	rollingFilename := rp.filenamePattern.convert(rp.timeInCurrentPeriod, rp.index)
	if len(rollingFilename) == 0 {
		rollingFilename = fmt.Sprintf("lork-%s.%v.log",
			rp.timeInCurrentPeriod.Format("2006-01-02"), rp.index)
	}

	archiver := rp.timeBasedRollingPolicy.archiver
	err = archiver.Archive(rp.fileWriter.RawFilename(), rollingFilename)
	if err != nil {
		return
	}

	rp.timeInCurrentPeriod = time.Now()
	rp.index++

	if rp.archiveRemover != nil {
		rp.archiveRemover.CleanAsync(time.Now())
	}

	return
}

func (rp *sizeAndTimeBasedRollingPolicy) calcIndex() error {
	rollingFilename := rp.filenamePattern.convert(rp.timeInCurrentPeriod, 0)
	parentDir := filepath.Dir(rollingFilename)
	files, err := os.ReadDir(parentDir)
	if err != nil {
		if os.IsNotExist(err) {
			rp.index = 0
			return nil
		} else {
			return err
		}
	}

	reg := rp.filenamePattern.toFilenameRegex()
	reg = afterLastSlash(reg)
	filenameRegex, err := regexp.Compile(reg)
	if err != nil {
		return err
	}

	var latestIndex = 0
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		result := filenameRegex.FindStringSubmatch(f.Name())
		if len(result) != 2 {
			continue
		}
		index, err := strconv.Atoi(result[1])
		if err != nil {
			continue
		}

		if index > latestIndex {
			latestIndex = index
		}
	}

	rp.index = latestIndex + 1

	return nil
}
