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

package slago

import (
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Archiver interface {
	// Archive makes an archiver with given filename and archive filename.
	Archive(filename, archiveFilename string) error
}

func newArchiver(filenamePattern string) Archiver {
	if strings.HasSuffix(filenamePattern, "gz") {
		return &gzipArchiver{}
	} else if strings.HasSuffix(filenamePattern, "zip") {
		return &zipArchiver{}
	} else {
		return &noneArchiver{}
	}
}

type noneArchiver struct {
}

func (a *noneArchiver) Archive(filename, archiveFilename string) error {
	return rename(filename, archiveFilename)
}

type gzipArchiver struct {
}

func (a *gzipArchiver) Archive(filename, archiveFilename string) error {
	origin, target, err := renameOpenFile(filename, archiveFilename)
	if err != nil {
		return err
	}

	w := gzip.NewWriter(target)
	w.Header = gzip.Header{
		ModTime: time.Now(),
		Name:    strings.TrimSuffix(archiveFilename, ".gz"),
	}
	go func() {
		_, _ = io.Copy(w, origin)
		_ = w.Close()
		_ = origin.Close()
		_ = os.Remove(filepath.Join(filepath.Dir(archiveFilename), origin.Name()))
	}()

	return nil
}

type zipArchiver struct {
}

func (a *zipArchiver) Archive(filename, archiveFilename string) error {
	origin, target, err := renameOpenFile(filename, archiveFilename)
	if err != nil {
		return err
	}

	entryName := strings.TrimSuffix(archiveFilename, ".zip")
	zw := zip.NewWriter(target)
	w, err := zw.CreateHeader(&zip.FileHeader{
		Name:     entryName,
		Modified: time.Now(),
		Method:   zip.Deflate,
	})
	if err != nil {
		return err
	}
	go func() {
		_, _ = io.Copy(w, origin)
		_ = zw.Close()
		_ = origin.Close()
		_ = os.Remove(filepath.Join(filepath.Dir(archiveFilename), origin.Name()))
	}()

	return nil
}

func renameOpenFile(filename, archiveFilename string) (*os.File, *os.File, error) {
	tmpFilename := fmt.Sprintf("%s-%v.%s", archiveFilename, time.Now().UnixNano(), "tmp")
	if err := rename(filename, tmpFilename); err != nil {
		return nil, nil, err
	}

	origin, err := os.Open(tmpFilename)
	if err != nil {
		return nil, nil, err
	}
	target, err := os.OpenFile(archiveFilename, os.O_RDWR|os.O_CREATE, os.ModePerm)

	return origin, target, err
}
