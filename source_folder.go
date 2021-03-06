// Copyright (C) 2020 David Vogel
//
// This file is part of Galago.
//
// Galago is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 2 of the License, or
// (at your option) any later version.
//
// Galago is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Galago.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Dadido3/configdb/tree"
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
	hidden        bool
	home          bool
	sourceTags    *SourceTags
}

// Compile time check if SourceFolder implements Element.
var _ Element = (*SourceFolder)(nil)

// CreateSourceFolder returns a new instance of a folder source.
func CreateSourceFolder(parent Element, index int, urlName string, c tree.Node) (Element, error) {
	var name string
	if err := c.Get(".Name", &name); err != nil {
		return nil, fmt.Errorf("Configuration of source %q errornous: %w", urlName, err)
	}

	hidden, _ := c["Hidden"].(bool)
	home, _ := c["Home"].(bool)

	var path string
	if err := c.Get(".Path", &path); err != nil {
		return nil, fmt.Errorf("Configuration of source %q errornous: %w", urlName, err)
	}

	s := &SourceFolder{
		parent:   parent,
		index:    index,
		name:     name,
		urlName:  urlName,
		filePath: path,
		hidden:   hidden,
		home:     home,
	}

	// Add tags source pointing towards the source folder itself
	tagsValue, tags := c["Tags"]
	if tags {
		name, hidden, enabled := "", false, false

		switch value := tagsValue.(type) {
		case string:
			name, hidden, enabled = value, false, true
		case bool:
			name, hidden, enabled = "Tags", true, value
		}

		if enabled {
			s.sourceTags = &SourceTags{
				parent:        s,
				index:         0, // Assume it's always placed at the first position
				name:          name,
				urlName:       "_tags_",
				internalPaths: []string{strings.TrimPrefix(s.Path(), "/")},
				hidden:        hidden,
			}
		}
	}

	return s, nil
}

// Clone returns a clone with the given parent and index set
func (s *SourceFolder) Clone(parent Element, index int) Element {
	clone := *s

	clone.parent = parent
	clone.index = index

	return &clone
}

// Parent returns the parent element, duh.
func (s *SourceFolder) Parent() Element {
	return s.parent
}

// Index returns the index of the element in its parent children list.
func (s *SourceFolder) Index() int {
	return s.index
}

// Children returns the folders and images of a source.
func (s *SourceFolder) Children() ([]Element, error) {
	elements := []Element{}

	files, err := ioutil.ReadDir(s.filePath)
	if err != nil {
		return nil, err
	}

	// Place tag list as the first child
	if s.sourceTags != nil {
		elements = append(elements, s.sourceTags)
	}

	// Sort descending by name
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() > files[j].Name()
	})

	// Add folders
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
		}
	}

	// Add images
	for _, file := range files {
		if !file.IsDir() {
			// Is file
			// Check if file extension is one of the supported formats
			ext := strings.ToLower(filepath.Ext(file.Name()))
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
				elements = append(elements, img)
			}
		}
	}

	return elements, nil
}

// Path returns the absolute path of the element, but not the filesystem path.
// For details see ElementPath.
func (s *SourceFolder) Path() string {
	return ElementPath(s)
}

// IsContainer returns whether an element can contain other elements or not.
func (s *SourceFolder) IsContainer() bool {
	return true
}

// IsHidden returns whether this element can be listed as child or not.
func (s *SourceFolder) IsHidden() bool {
	return s.hidden
}

// IsHome returns whether an element should be linked by the home button or not.
func (s *SourceFolder) IsHome() bool {
	return s.home
}

// Name returns the name that is shown to the user.
func (s *SourceFolder) Name() string {
	return s.name
}

// URLName returns the name/identifier that is used in URLs.
func (s *SourceFolder) URLName() string {
	return s.urlName
}

// Traverse the element's children with the given path.
func (s *SourceFolder) Traverse(path string) (Element, error) {
	return TraverseElements(s, path)
}

func (s *SourceFolder) String() string {
	return fmt.Sprintf("{SourceFolder %q: %q}", s.Path(), s.filePath)
}

// SourceFolderImage represents an image that is contained in a locally accessible folder.
type SourceFolderImage struct {
	parent        Element
	index         int
	name, urlName string
	s             *SourceFolder
	filePath      string // The path to the file in the filesystem
	fileInfo      os.FileInfo
	cacheEntry    *CacheEntry
}

// Compile time check if SourceFolderImage implements Image and Element.
var _ Element = (*SourceFolderImage)(nil)
var _ Image = (*SourceFolderImage)(nil)

// Clone returns a clone with the given parent and index set
func (si *SourceFolderImage) Clone(parent Element, index int) Element {
	clone := *si

	clone.parent = parent
	clone.index = index

	return &clone
}

// Parent returns the parent element, duh.
func (si *SourceFolderImage) Parent() Element {
	return si.parent
}

// Index returns the index of the element in its parent children list.
func (si *SourceFolderImage) Index() int {
	return si.index
}

// Children returns nothing, as images don't contain other elements.
func (si *SourceFolderImage) Children() ([]Element, error) {
	//return []Element{}, fmt.Errorf("Images don't contain children")
	return []Element{}, nil
}

// Path returns the absolute path of the element, but not the filesystem path.
// For details see ElementPath.
func (si *SourceFolderImage) Path() string {
	return ElementPath(si)
}

// IsContainer returns whether an element can contain other elements or not.
func (si *SourceFolderImage) IsContainer() bool {
	return false
}

// IsHidden returns whether this element can be listed as child or not.
func (si *SourceFolderImage) IsHidden() bool {
	return false
}

// IsHome returns whether an element should be linked by the home button or not.
func (si *SourceFolderImage) IsHome() bool {
	return false
}

// Name returns the name that is shown to the user.
func (si *SourceFolderImage) Name() string {
	ce, err := si.CacheEntry()
	if err != nil {
		log.Errorf("Couldn't get or generate cache entry for %v: %v", si, err)
		return si.name
	}

	if ce.Title != "" {
		return ce.Title
	}

	return si.name
}

// URLName returns the name/identifier that is used in URLs.
func (si *SourceFolderImage) URLName() string {
	return si.urlName
}

// Traverse the element's children with the given path.
func (si *SourceFolderImage) Traverse(path string) (Element, error) {
	return TraverseElements(si, path)
}

// Hash returns a unique hash that stays the same as long as the file doesn't change.
func (si *SourceFolderImage) Hash() string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("SourceFolderImage %q %v", si.filePath, si.fileInfo.ModTime()))) // This should be unique enough
	return fmt.Sprintf("%x", h.Sum(nil))
}

// CacheEntry returns the cache entry of the image.
//
// This function is similar to calling QueryCacheEntryImage with the image's hash.
// But the cache entry is cached over the lifetime of this object, so it's more efficient than multiple cache queries.
//
// If no cache entry can be found, a new one will be generated.
// This function will block and then return a valid cache entry, if one could be generated.
// An error will be returned otherwise.
func (si *SourceFolderImage) CacheEntry() (*CacheEntry, error) {
	if si.cacheEntry != nil {
		return si.cacheEntry, nil
	}

	ce, err := cache.QueryCacheEntryImage(si)
	if err != nil {
		return nil, err
	}
	if ce != nil {
		si.cacheEntry = ce
	}

	return ce, nil
}

// Width of the original image.
//
// This value is stored in the cache, so it is fast to get.
func (si *SourceFolderImage) Width() int {
	ce, err := si.CacheEntry()
	if err != nil {
		log.Errorf("Couldn't get or generate cache entry for %v: %v", si, err)
		return 0
	}

	return ce.Width
}

// Height of the original image.
//
// This value is stored in the cache, so it is fast to get.
func (si *SourceFolderImage) Height() int {
	ce, err := si.CacheEntry()
	if err != nil {
		log.Errorf("Couldn't get or generate cache entry for %v: %v", si, err)
		return 0
	}

	return ce.Height
}

// FileContent returns the compressed image file.
func (si *SourceFolderImage) FileContent() (io.ReadCloser, int64, string, error) {
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

func (si *SourceFolderImage) String() string {
	return fmt.Sprintf("{SourceFolderImage %q: %q}", si.Path(), si.filePath)
}
