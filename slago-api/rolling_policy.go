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
	"text/template"
	"time"
)

// RollingPolicy represents policy for log rolling.
type RollingPolicy interface {
	// Start starts rolling config
	Start() error

	// Attach attaches file writer.
	Attach(w *fileWriter)

	// ShouldTrigger check if there's necessary to trigger rolling.
	ShouldTrigger(fileSize int64) bool

	// Rotate starts log rolling.
	Rotate(filename string) error
}

type noopRollingPolicy struct {
}

func (rp *noopRollingPolicy) Start() error {
	return nil
}

func (rp *noopRollingPolicy) Attach(w *fileWriter) {
}

func (rp *noopRollingPolicy) ShouldTrigger(fileSize int64) bool {
	return false
}

func (rp *noopRollingPolicy) Rotate(filename string) error {
	return nil
}

type timeBasedRollingPolicy struct {
}

type sizeBasedRollingPolicy struct {
}

type sizeAndTimeBasedRollingPolicy struct {
	fileWriter      *fileWriter
	triggerSize     int64
	filenamePattern string
	tpl             *template.Template
	index           int
	nextCheck       time.Time
}

// NewSizeAndTimeBasedRollingPolicy creates a new instance of size and time
// based rolling policy for file writer.
func NewSizeAndTimeBasedRollingPolicy(filenamePattern string,
	maxSize string) *sizeAndTimeBasedRollingPolicy {
	fileSize, err := parseFileSize(maxSize)
	if err != nil {
		Reportf("parse file size error: %v", err)
		os.Exit(0)
	}

	return &sizeAndTimeBasedRollingPolicy{
		triggerSize:     fileSize,
		filenamePattern: filenamePattern,
	}
}

func (rp *sizeAndTimeBasedRollingPolicy) Start() error {
	rp.filenamePattern = strings.TrimSpace(rp.filenamePattern)
	if len(rp.filenamePattern) == 0 {
		return errors.New("no valid filename pattern for this policy")
	}

	funcs := make(template.FuncMap)
	funcs[dateKey] = func(format string) string {
		return time.Now().Format(format)
	}
	tpl, err := template.New("").Funcs(funcs).Parse(rp.filenamePattern)
	if err != nil {
		return err
	}
	rp.tpl = tpl

	rp.calcNextCheck()
	return rp.calcIndex()
}

func (rp *sizeAndTimeBasedRollingPolicy) Attach(w *fileWriter) {
	rp.fileWriter = w
}

func (rp *sizeAndTimeBasedRollingPolicy) ShouldTrigger(fileSize int64) bool {
	if time.Now().After(rp.nextCheck) {
		rp.calcNextCheck()
		rp.index = 0
		return true
	}

	if fileSize >= rp.triggerSize {
		return true
	}

	return false
}

func (rp *sizeAndTimeBasedRollingPolicy) Rotate(filename string) (err error) {
	var rollingFilename string
	buf := &bytes.Buffer{}

	if err = rp.tpl.Execute(buf, map[string]interface{}{
		indexKey: rp.index,
	}); err != nil {
		rollingFilename = fmt.Sprintf("slago.%s.log", time.Now().Format("2006-01-02"))
	}

	rollingFilename = buf.String()
	buf.Reset()

	dir := filepath.Dir(rollingFilename)
	err = os.MkdirAll(dir, os.FileMode(0666))
	if err != nil {
		return
	}

	err = os.Rename(filename, rollingFilename)
	if err != nil {
		return
	}

	rp.index++

	return
}

func (rp *sizeAndTimeBasedRollingPolicy) calcNextCheck() {
	now := time.Now()
	rp.nextCheck = time.Date(
		now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
}

func (rp *sizeAndTimeBasedRollingPolicy) calcIndex() error {
	files, err := ioutil.ReadDir(rp.fileWriter.Dir())
	if err != nil {
		return err
	}

	var lastModTime time.Time
	fi, err := os.Stat(rp.fileWriter.Filename())
	if err != nil {
		lastModTime = time.Now()
	} else {
		lastModTime = fi.ModTime()
	}

	var buf = &bytes.Buffer{}
	funcs := make(template.FuncMap)
	funcs[dateKey] = func(format string) string {
		return lastModTime.Format(format)
	}
	tpl, err := template.New("").Funcs(funcs).Parse(rp.filenamePattern)
	if err != nil {
		return err
	}
	if err := tpl.Execute(buf, map[string]interface{}{
		indexKey: "(\\d{1,3})",
	}); err != nil {
		return err
	}

	filenameRegex, err := regexp.Compile(buf.String())
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
