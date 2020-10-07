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
	"fmt"
	"strings"
)

// Element represents either a source, an album or an image.
type Element interface {
	Clone(parent Element, index int) Element // Returns a clone with the given parent and index set
	Parent() Element                         // Returns the parent element
	Index() int                              // Returns the index of the element in its parent children list
	Children() ([]Element, error)            // Elements can contain more elements (Like sources or albums)
	Path() string                            // Returns the absolute path of the element, but not the filesystem path
	IsContainer() bool                       // Returns whether an element can contain other elements or not
	IsHidden() bool                          // Returns whether an element is hidden. If it is, it can only be accessed when the url is known
	Name() string                            // The name that is shown to the user
	URLName() string                         // The name/identifier used in URLs
	Traverse(path string) (Element, error)   // Traverse the element's children with the given path
}

// FilterNonEmpty takes a list of elements, and returns only elements that contain something else.
// The content of the albums stays untouched.
func FilterNonEmpty(ee []Element) []Element {
	result := []Element{}
	for _, element := range ee {
		cc, err := element.Children()
		if err != nil {
			log.Errorf("Error while retrieving children of %v: %v", element, err)
			continue
		}
		if len(cc) > 0 {
			result = append(result, element)
		}
	}
	return result
}

// FilterContainers takes a list of elements, and returns only elements that can contain something else.
// The content of the albums stays untouched.
func FilterContainers(ee []Element) []Element {
	result := []Element{}
	for _, element := range ee {
		if element.IsContainer() {
			result = append(result, element)
		}
	}
	return result
}

// ErrorNotFound is returned when a path points to a non existing element.
type ErrorNotFound struct {
	path string
}

func (e *ErrorNotFound) Error() string {
	return fmt.Sprintf("Element %q doesn't exist", e.path)
}

// Prepend prepends an element name to the error's path.
func (e *ErrorNotFound) Prepend(p string) {
	e.path = strings.Join([]string{p, e.path}, "/")
}

// TraverseElements will traverse through the children of elements until the element at path is reached.
// As the path is relative to the given origin, the leading part of the path defines the first child element.
// As an edge case, an empty path points to the origin.
//
// An example for a path: animals/cats/img.jpg
func TraverseElements(origin Element, path string) (Element, error) {
	pathElements := strings.Split(path, "/")

	// Edge case: If the path is empty, return the current object
	if path == "" {
		return origin, nil
	}

	children, err := origin.Children()
	if err != nil {
		return nil, err
	}
	for _, child := range children {
		if child.URLName() == pathElements[0] {
			res, err := TraverseElements(child, strings.Join(pathElements[1:], "/"))
			if err, ok := err.(*ErrorNotFound); ok {
				err.Prepend(pathElements[0])
				return nil, err
			}
			return res, err
		}
	}

	return nil, &ErrorNotFound{pathElements[0]}
}

// PreviousElement returns the previous neighbor element if possible, or an error otherwise.
// Having no parent or no neighbor doesn't count as error.
func PreviousElement(e Element) (Element, error) {
	if parent, index := e.Parent(), e.Index(); parent != nil {
		children, err := parent.Children()
		if err != nil {
			return nil, err
		}

		if index > 0 && len(children) > 0 {
			return children[index-1], nil
		}
	}

	return nil, nil
}

// NextElement returns the next neighbor element if possible, or an error otherwise.
// Having no parent or no neighbor doesn't count as error.
func NextElement(e Element) (Element, error) {
	if parent, index := e.Parent(), e.Index(); parent != nil {
		children, err := parent.Children()
		if err != nil {
			return nil, err
		}

		if index >= 0 && len(children) > index+1 {
			return children[index+1], nil
		}
	}

	return nil, nil
}

// FilterNonHidden takes a list of elements, and returns only non hidden ones.
func FilterNonHidden(ee []Element) []Element {
	result := []Element{}
	for _, element := range ee {
		if !element.IsHidden() {
			result = append(result, element)
		}
	}
	return result
}

// GetPreviewImages tries to return up to n images that are contained inside the given element e.
// This will iterate over all children until the needed amount of images is found.
// Hidden children (images or containers) will be ignored, though.
func GetPreviewImages(e Element, n int) ([]Image, error) {
	result := []Image{}

	children, err := e.Children()
	if err != nil {
		return nil, err
	}
	children = FilterNonHidden(children)

	// Fill result with direct children image elements
	images := FilterImages(children)
	for _, image := range images {
		if len(result) >= n {
			return result, nil
		}
		result = append(result, image)
	}

	// Recursively go through child elements and get images from there
	for _, child := range children {
		if len(result) >= n {
			return result, nil
		}

		images, err := GetPreviewImages(child, n-len(result))
		if err != nil {
			return nil, err
		}

		result = append(result, images...)
	}

	return result, nil
}

// ElementPath returns the absolute path of the element.
// This will not return the file system path, but the path that an object can be addressed inside of Galago.
// This will include the root element, that has an empty name.
//
// Example output: /source123/cats/cat.jpg
func ElementPath(e Element) string {
	parent := e.Parent()

	if parent == nil {
		return e.URLName()
	}
	return strings.Join([]string{ElementPath(parent), e.URLName()}, "/")
}
