// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).
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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

// NewNoopRollingPolicy creates a new instance of rolling policy which will do nothing.
func NewNoopRollingPolicy() *noopRollingPolicy {
	return &noopRollingPolicy{}
}

func (rp *noopRollingPolicy) Prepare() error {
	return nil
}

func (rp *noopRollingPolicy) Attach(w *fileWriter) {
}

func (rp *noopRollingPolicy) ShouldTrigger(fileSize int64) bool {
	return false
}

func (rp *noopRollingPolicy) Rotate() error {
	return nil
}

type sizeBasedRollingPolicy struct {
}

type timeBasedRollingPolicy struct {
	fileWriter      *fileWriter
	filenamePattern string
	nextCheck       time.Time
}

func NewTimeBasedRollingPolicy(filenamePattern string) *timeBasedRollingPolicy {
	return &timeBasedRollingPolicy{
		filenamePattern: filenamePattern,
	}
}

func (rp *timeBasedRollingPolicy) Prepare() error {
	rp.filenamePattern = strings.TrimSpace(rp.filenamePattern)
	if len(rp.filenamePattern) == 0 {
		return errors.New("no valid filename pattern for this policy")
	}

	rp.calcNextCheck()
	return nil
}

func (rp *timeBasedRollingPolicy) Attach(w *fileWriter) {
	rp.fileWriter = w
}

func (rp *timeBasedRollingPolicy) ShouldTrigger(fileSize int64) bool {
	if time.Now().After(rp.nextCheck) {
		rp.calcNextCheck()
		return true
	}

	return false
}

func (rp *timeBasedRollingPolicy) Rotate() error {
	panic("implement me")
}

func (rp *timeBasedRollingPolicy) calcNextCheck() {
	now := time.Now()
	rp.nextCheck = time.Date(
		now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
}

type sizeAndTimeBasedRollingPolicy struct {
	fileWriter  *fileWriter
	triggerSize int64
	index       int
	nextCheck   time.Time

	patternParser *patternParser
	converter     Converter
	rollingDate   *rollingDate
}

// SizeAndTimeBasedRPOption represents available options for size and time
// based rolling policy.
type SizeAndTimeBasedRPOption struct {
	FilenamePattern string
	MaxFileSize     string
}

// NewSizeAndTimeBasedRollingPolicy creates a new instance of size and time
// based rolling policy for file writer.
func NewSizeAndTimeBasedRollingPolicy(options ...func(
	*SizeAndTimeBasedRPOption)) *sizeAndTimeBasedRollingPolicy {
	opt := &SizeAndTimeBasedRPOption{
		MaxFileSize:     "128MB",
		FilenamePattern: "slago-archive.#date{2006-01-02}.#index.log",
	}

	for _, f := range options {
		f(opt)
	}

	fileSize, err := parseFileSize(opt.MaxFileSize)
	if err != nil {
		ReportfExit("parse file size error: %v", err)
	}

	return &sizeAndTimeBasedRollingPolicy{
		triggerSize:   fileSize,
		patternParser: NewPatternParser(opt.FilenamePattern),
	}
}

func (rp *sizeAndTimeBasedRollingPolicy) Prepare() error {
	node, err := rp.patternParser.Parse()
	if err != nil {
		return err
	}

	converters := map[string]NewConverter{
		"index": newIndexConverter,
		"date":  newDateConverter,
	}
	converter, err := NewPatternCompiler(node, converters).Compile()
	if err != nil {
		return err
	}
	rp.converter = converter

	var gotIndex bool
	var datePattern string
	for c := rp.converter; c != nil; c = c.Next() {
		if dc, ok := c.(*dateConverter); ok {
			datePattern = dc.DatePattern()
		}
		if _, ok := c.(*indexConverter); ok {
			gotIndex = true
		}
	}

	if !gotIndex {
		return errors.New("invalid filename pattern, missing index pattern")
	}

	if len(datePattern) == 0 {
		return errors.New("invalid filename pattern, missing date pattern")
	}

	rp.rollingDate = newRollingDate(datePattern)
	rp.calcNextCheck()

	return rp.calcIndex()
}

func (rp *sizeAndTimeBasedRollingPolicy) Attach(w *fileWriter) {
	rp.fileWriter = w
}

func (rp *sizeAndTimeBasedRollingPolicy) ShouldTrigger(fileSize int64) bool {
	if time.Now().After(rp.nextCheck) {
		rp.calcNextCheck()
		rp.index = 1
		return true
	}

	if fileSize >= rp.triggerSize {
		return true
	}

	return false
}

func (rp *sizeAndTimeBasedRollingPolicy) Rotate() (err error) {
	var rollingFilename string
	buf := &bytes.Buffer{}

	for c := rp.converter; c != nil; c = c.Next() {
		switch c.(type) {
		case *literalConverter:
			c.Convert(nil, buf)

		case *dateConverter:
			ts := time.Now().Format(time.RFC3339)
			c.Convert([]byte(ts), buf)

		case *indexConverter:
			c.Convert([]byte(strconv.Itoa(rp.index)), buf)
		}
	}

	rollingFilename = buf.String()
	buf.Reset()
	if len(rollingFilename) == 0 {
		rollingFilename = fmt.Sprintf("slago-%s.%v.log",
			time.Now().Format("2006-01-02"), rp.index)
	}

	dir := filepath.Dir(rollingFilename)
	err = os.MkdirAll(dir, os.FileMode(0666))
	if err != nil {
		return
	}

	err = os.Rename(rp.fileWriter.Filename(), rollingFilename)
	if err != nil {
		return
	}

	rp.index++

	return
}

func (rp *sizeAndTimeBasedRollingPolicy) calcNextCheck() {
	rp.nextCheck = rp.rollingDate.next()
}

func (rp *sizeAndTimeBasedRollingPolicy) calcIndex() error {
	files, err := ioutil.ReadDir(rp.fileWriter.Dir())
	if err != nil {
		return err
	}

	reg := rp.toFilenameRegex()
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

func (rp *sizeAndTimeBasedRollingPolicy) toFilenameRegex() string {
	var buf = &bytes.Buffer{}
	for c := rp.converter; c != nil; c = c.Next() {
		switch c.(type) {
		case *literalConverter:
			c.Convert(nil, buf)

		case *dateConverter:
			ts := time.Now().Format(time.RFC3339)
			c.Convert([]byte(ts), buf)

		case *indexConverter:
			buf.WriteString("(\\d{1,3})")
		}
	}

	reg := buf.String()
	buf.Reset()

	return reg
}

const (
	topOfSecond periodType = iota + 1
	topOfMinute
	topOfHour
	topOfDay
	topOfMonth
)

var (
	periods = []periodType{
		topOfSecond, topOfMinute, topOfHour, topOfDay, topOfMonth,
	}
)

type periodType int8

type rollingDate struct {
	datePattern string
	_type       periodType
}

func newRollingDate(datePattern string) *rollingDate {
	rd := &rollingDate{
		datePattern: datePattern,
	}
	rd._type = rd.calcPeriodType()

	return rd
}

func (rd *rollingDate) calcPeriodType() periodType {
	now := time.Now()
	for _, t := range periods {
		tl := now.Format(rd.datePattern)
		next := rd.endOfPeriod(t, now)
		tr := next.Format(rd.datePattern)
		if tl != tr {
			return t
		}
	}

	return topOfSecond
}

func (rd *rollingDate) endOfPeriod(pt periodType, now time.Time) time.Time {
	switch pt {
	case topOfMinute:
		return time.Date(
			now.Year(), now.Month(), now.Day(), now.Hour(),
			now.Minute()+1, 0, 0, now.Location())
	case topOfHour:
		return time.Date(
			now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())
	case topOfDay:
		return time.Date(
			now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	case topOfMonth:
		return time.Date(
			now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location())
	case topOfSecond:
		fallthrough
	default:
		return time.Date(
			now.Year(), now.Month(), now.Day(), now.Hour(),
			now.Minute(), now.Second()+1, 0, now.Location())
	}
}

func (rd *rollingDate) next() time.Time {
	return rd.endOfPeriod(rd._type, time.Now())
}
