// Copyright (C) 2020 David Vogel
//
// This file is part of galago.
//
// galago is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 2 of the License, or
// (at your option) any later version.
//
// galago is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with galago.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	registerSourceType("folder", CreateSourceFolder)
}

// SourceFolder represents a source that can return the content of an local available folder.
type SourceFolder struct {
	parent        Element
	index         int
	name, urlName string
	filePath      string
}

// CreateSourceFolder returns a new instance of a folder source.
func CreateSourceFolder(parent Element, index int, urlName string, c map[string]interface{}) (Element, error) {
	var ok bool
	var pathInt interface{}
	if pathInt, ok = c["Path"]; !ok {
		return nil, fmt.Errorf("Missing %q field", "Path")
	}
	var path string
	if path, ok = pathInt.(string); !ok {
		return nil, fmt.Errorf("%q field is of the wrong type. Expected %T, got %T", "Path", path, pathInt)
	}

	var nameInt interface{}
	if nameInt, ok = c["Name"]; !ok {
		return nil, fmt.Errorf("Missing %q field", "Name")
	}
	var name string
	if name, ok = nameInt.(string); !ok {
		return nil, fmt.Errorf("%q field is of the wrong type. Expected %T, got %T", "Name", path, pathInt)
	}

	return SourceFolder{
		parent:   parent,
		index:    index,
		name:     name,
		urlName:  urlName,
		filePath: path,
	}, nil
}

// Parent returns the parent element, duh.
func (s SourceFolder) Parent() Element {
	return s.parent
}

// Index returns the index of the element in its parent children list.
func (s SourceFolder) Index() int {
	return s.index
}

// Children returns the folders and images of a source.
func (s SourceFolder) Children() ([]Element, error) {
	elements := []Element{}

	files, err := ioutil.ReadDir(s.filePath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			// Is directory
			// Return a new SourceFolder object of the subfolder
			album := &SourceFolder{
				parent:   s,
				index:    len(elements),
				name:     file.Name(),
				urlName:  strings.ToLower(file.Name()),
				filePath: filepath.Join(s.filePath, file.Name()),
			}
			elements = append(elements, album)
		} else {
			// Is file
			// Check if file extension is one of the supported formats
			ext := filepath.Ext(file.Name())
			if validExtensions[ext] {
				img := &SourceFolderImage{
					parent:   s,
					index:    len(elements),
					name:     file.Name(),
					urlName:  strings.ToLower(file.Name()),
					s:        s,
					filePath: filepath.Join(s.filePath, file.Name()),
					fileInfo: file,
				}
				cacheEntry, err := cache.QueryCacheEntryImage(img)
				if err != nil {
					log.Warnf("Couldn't generate cache entry for %v: %v", img, err)
				}
				if cacheEntry != nil {
					img.cacheEntry = *cacheEntry
				}
				elements = append(elements, img)
			}
		}
	}

	return elements, nil
}

// Path returns the absolute path of the element, but not the filesystem path.
// For details see ElementPath.
func (s SourceFolder) Path() string {
	return ElementPath(s)
}

// Container returns whether an element can contain other elements or not.
func (s SourceFolder) Container() bool {
	return true
}

// Name returns the name that is shown to the user.
func (s SourceFolder) Name() string {
	return s.name
}

// URLName returns the name/identifier that is used in URLs.
func (s SourceFolder) URLName() string {
	return s.urlName
}

// Traverse the element's children with the given path.
func (s SourceFolder) Traverse(path string) (Element, error) {
	return TraverseElements(s, path)
}

func (s SourceFolder) String() string {
	return fmt.Sprintf("{SourceFolder %q: %q}", s.Path(), s.filePath)
}

// SourceFolderImage represents an image that is contained in a locally accessible folder.
type SourceFolderImage struct {
	parent        Element
	index         int
	name, urlName string
	s             SourceFolder
	filePath      string // The path to the file in the filesystem
	fileInfo      os.FileInfo
	cacheEntry    CacheEntry
}

// Parent returns the parent element, duh.
func (si SourceFolderImage) Parent() Element {
	return si.parent
}

// Index returns the index of the element in its parent children list.
func (si SourceFolderImage) Index() int {
	return si.index
}

// Children returns nothing, as images don't contain other elements.
func (si SourceFolderImage) Children() ([]Element, error) {
	//return []Element{}, fmt.Errorf("Images don't contain children")
	return []Element{}, nil
}

// Path returns the absolute path of the element, but not the filesystem path.
// For details see ElementPath.
func (si SourceFolderImage) Path() string {
	return ElementPath(si)
}

// Container returns whether an element can contain other elements or not.
func (si SourceFolderImage) Container() bool {
	return false
}

// Name returns the name that is shown to the user.
func (si SourceFolderImage) Name() string {
	if si.cacheEntry.Title != "" {
		return si.cacheEntry.Title
	}

	return si.name
}

// URLName returns the name/identifier that is used in URLs.
func (si SourceFolderImage) URLName() string {
	return si.urlName
}

// Traverse the element's children with the given path.
func (si SourceFolderImage) Traverse(path string) (Element, error) {
	return TraverseElements(si, path)
}

// Hash returns a unique hash that stays the same as long as the file doesn't change.
func (si SourceFolderImage) Hash() string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("SourceFolderImage %q %v", si.filePath, si.fileInfo.ModTime()))) // This should be unique enough
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Width of the original image.
//
// This value is stored in the cache, so it is fast to get.
func (si SourceFolderImage) Width() int {
	return si.cacheEntry.Width
}

// Height of the original image.
//
// This value is stored in the cache, so it is fast to get.
func (si SourceFolderImage) Height() int {
	return si.cacheEntry.Height
}

// FileContent returns the compressed image file.
func (si SourceFolderImage) FileContent() (io.ReadCloser, int64, string, error) {
	f, err := os.Open(si.filePath)
	if err != nil {
		return nil, 0, "", err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, 0, "", err
	}
	return f, stat.Size(), ExtToMIME(filepath.Ext(f.Name())), err
}

func (si SourceFolderImage) String() string {
	return fmt.Sprintf("{SourceFolderImage %q: %q}", si.Path(), si.filePath)
}
