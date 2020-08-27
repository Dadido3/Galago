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
	"fmt"
	"image"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func init() {
	registerSourceType("folder", CreateSourceFolder)
}

// SourceFolder represents a source that can return the content of an local available folder.
type SourceFolder struct {
	parent        Element
	name, urlName string
	path          string
}

// CreateSourceFolder returns a new instance of a folder source.
func CreateSourceFolder(parent Element, urlName string, c map[string]interface{}) (Element, error) {
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
		parent:  parent,
		name:    name,
		urlName: urlName,
		path:    path,
	}, nil
}

// Parent returns the parent element, duh.
func (s SourceFolder) Parent() Element {
	return s.parent
}

// Children returns the folders and images of a source.
func (s SourceFolder) Children() ([]Element, error) {
	album, err := s.childrenRecursive(s, s.path)
	if err != nil {
		return []Element{}, err
	}

	// Return the children, not the album itself
	return album.Children()
}

// Path returns the absolute path of the element, but not the filesystem path.
// For details see ElementPath.
func (s SourceFolder) Path() string {
	return ElementPath(s)
}

func (s SourceFolder) childrenRecursive(parent Element, path string) (Album, error) {
	album := Album{
		parent:  parent,
		name:    filepath.Base(path),
		urlName: strings.ToLower(filepath.Base(path)),
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return album, err
	}

	for _, file := range files {
		if file.IsDir() {
			// Is directory
			childAlbum, err := s.childrenRecursive(album, filepath.Join(path, file.Name()))
			if err != nil {
				return album, err // Just return the error, otherwise the error message may get very long
			}
			album.children = append(album.children, childAlbum)
		} else {
			// Is file
			// Check if file extension is one of the supported formats
			ext := filepath.Ext(file.Name())
			if validExtensions[ext] {
				album.children = append(album.children, SourceFolderImage{
					parent:  parent,
					name:    file.Name(),
					urlName: strings.ToLower(file.Name()),
					s:       s,
					path:    filepath.Join(path, file.Name()),
				})
			}
		}
	}

	return album, nil
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
	return fmt.Sprintf("{SourceFolder: %q}", s.path)
}

// SourceFolderImage represents an image that is contained in a locally accessible folder.
type SourceFolderImage struct {
	parent        Element
	name, urlName string
	s             SourceFolder
	path          string
}

// Parent returns the parent element, duh.
func (si SourceFolderImage) Parent() Element {
	return si.parent
}

// Children returns nothing, as images don't contain more elements.
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

// LoadImage loads and returns the image from the source.
func (si SourceFolderImage) LoadImage() image.Image {
	return image.Rect(0, 0, 100, 100) // TODO: Implement image loading
}

// Name returns the name that is shown to the user.
func (si SourceFolderImage) Name() string {
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

func (si SourceFolderImage) String() string {
	return fmt.Sprintf("{SourceFolderImage: %q}", si.path)
}
