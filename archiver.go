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
	"archive/zip"
	"compress/gzip"
	"errors"
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
	go compressArchive(origin, target, w, w)

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
		return fmt.Errorf("archiver create zip entry error; original retained at %s: %w",
			origin.Name(), errors.Join(err, zw.Close(), target.Close(), origin.Close()))
	}
	go compressArchive(origin, target, w, zw)

	return nil
}

func compressArchive(origin, target *os.File, writer io.Writer, compressor io.Closer) {
	_, err := io.Copy(writer, origin)
	err = errors.Join(err, compressor.Close(), target.Close(), origin.Close())
	if err != nil {
		Reportf("archiver compress %s error; original retained at %s: %v",
			target.Name(), origin.Name(), err)
		return
	}

	if err := os.Remove(origin.Name()); err != nil {
		Reportf("archiver remove file %s error: %v", origin.Name(), err)
	}
}

func archiveCompressionSuffix(filename string) string {
	for _, suffix := range []string{".gz", ".zip", "gz", "zip"} {
		if strings.HasSuffix(filename, suffix) {
			return suffix
		}
	}
	return ""
}

func uncompressedArchiveFilename(filename string) string {
	return strings.TrimSuffix(filename, archiveCompressionSuffix(filename))
}

func renameOpenFile(filename, archiveFilename string) (*os.File, *os.File, error) {
	originalFilename := uncompressedArchiveFilename(archiveFilename)
	if originalFilename == "" || originalFilename == archiveFilename {
		return nil, nil, fmt.Errorf("invalid compressed archive filename: %s", archiveFilename)
	}
	sourcePath, err := filepath.Abs(filename)
	if err != nil {
		return nil, nil, err
	}
	originalPath, err := filepath.Abs(originalFilename)
	if err != nil {
		return nil, nil, err
	}
	if sourcePath == originalPath {
		return nil, nil, fmt.Errorf("uncompressed archive conflicts with active log: %s", originalFilename)
	}
	if err := mkdirIfNotExist(filepath.Dir(archiveFilename)); err != nil {
		return nil, nil, err
	}
	target, err := os.OpenFile(archiveFilename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.ModePerm)
	if err != nil {
		return nil, nil, err
	}
	ready := false
	defer func() {
		if !ready {
			_ = target.Close()
			_ = os.Remove(archiveFilename)
		}
	}()

	// Reserve the uncompressed name so a retained original is never overwritten.
	reservation, err := os.OpenFile(originalFilename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.ModePerm)
	if err != nil {
		return nil, nil, err
	}
	if err := reservation.Close(); err != nil {
		_ = os.Remove(originalFilename)
		return nil, nil, err
	}
	if err := rename(filename, originalFilename); err != nil {
		_ = os.Remove(originalFilename)
		return nil, nil, err
	}

	origin, err := os.Open(originalFilename)
	if err != nil {
		return nil, nil, fmt.Errorf("archiver open original retained at %s: %w", originalFilename, err)
	}

	ready = true
	return origin, target, nil
}
