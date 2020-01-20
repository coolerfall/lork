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
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var _ Writer = (*fileWriter)(nil)

const defaultLogFilename = "slago.log"

type fileWriter struct {
	opts *FileWriterOption

	locker sync.Mutex
	file   *os.File
	size   int64
}

// FileWriterOption represents available options for file writer.
type FileWriterOption struct {
	Filter        Filter
	Encoder       Encoder
	RollingPolicy RollingPolicy
	Filename      string
}

// NewFileWriter creates a new instance of file writer.
func NewFileWriter(options ...func(*FileWriterOption)) Writer {
	opts := &FileWriterOption{
		Filename: defaultLogFilename,
		Encoder:  NewJsonEncoder(),
	}

	for _, f := range options {
		f(opts)
	}

	fw := &fileWriter{
		opts: opts,
	}
	if opts.RollingPolicy == nil {
		opts.RollingPolicy = NewNoopRollingPolicy()
	}
	opts.RollingPolicy.Attach(fw)

	return fw
}

func (fw *fileWriter) Start() {
	if err := fw.openExistingOrNew(); err != nil {
		ReportfExit("file writer start error: %v", err)
	}

	if err := fw.opts.RollingPolicy.Prepare(); err != nil {
		ReportfExit("start rolling policy error: %v\n", err)
	}
}

func (fw *fileWriter) Stop() {
	_ = fw.Close()
}

func (fw *fileWriter) Write(p []byte) (n int, err error) {
	fw.locker.Lock()
	defer fw.locker.Unlock()

	writeLen := len(p)
	if fw.file == nil {
		if err = fw.openExistingOrNew(); err != nil {
			return
		}
	}

	if fw.opts.RollingPolicy.ShouldTrigger(fw.size + int64(writeLen)) {
		if err = fw.rotate(); err != nil {
			return
		} else {
			if err = fw.openNew(); err != nil {
				return
			}
		}
	}

	n, err = fw.file.Write(p)
	fw.size += int64(n)

	return n, err
}

func (fw *fileWriter) Encoder() Encoder {
	return fw.opts.Encoder
}

func (fw *fileWriter) Filter() Filter {
	return fw.opts.Filter
}

// Close implements io.Closer, and closes the current logfile.
func (fw *fileWriter) Close() error {
	fw.locker.Lock()
	defer fw.locker.Unlock()
	return fw.close()
}

// close closes the file if it is open.
func (fw *fileWriter) close() error {
	if fw.file == nil {
		return nil
	}
	err := fw.file.Close()
	fw.file = nil
	return err
}

// openNew opens a new log file for writing, moving any old log file out of the
// way. This methods assumes the file has already been closed.
func (fw *fileWriter) openNew() error {
	err := os.MkdirAll(fw.Dir(), 0755)
	if err != nil {
		return fmt.Errorf("can't make directories for new logfile: %s", err)
	}

	f, err := os.OpenFile(fw.opts.Filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("can't open new log file: %s", err)
	}
	fw.file = f
	fw.size = 0

	return nil
}

// openExistingOrNew opens the logfile if it exists and if the current write
// would not put it over MaxSize.  If there is no such file or the write would
// put it over the MaxSize, a new file is created.
func (fw *fileWriter) openExistingOrNew() error {
	filename := fw.opts.Filename
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return fw.openNew()
	}

	if err != nil {
		return fmt.Errorf("error getting log file info: %s", err)
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// if we fail to open the old log file for some reason, just ignore
		// it and open a new log file.
		return fw.openNew()
	}
	fw.file = file
	fw.size = info.Size()

	return nil
}

// Dir returns the directory for the current filename.
func (fw *fileWriter) Dir() string {
	return filepath.Dir(fw.opts.Filename)
}

// Filename returns the filename of current file.
func (fw *fileWriter) Filename() string {
	return fw.opts.Filename
}

func (fw *fileWriter) rotate() (err error) {
	err = fw.close()
	if err != nil {
		return err
	}

	err = fw.opts.RollingPolicy.Rotate()
	if err != nil {
		return
	}

	return fw.openNew()
}
