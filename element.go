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
	"strings"
)

// Element represents either a source, an album or an image.
type Element interface {
	Parent() Element                       // Returns the parent element
	Index() int                            // Returns the index of the element in its parent children list
	Children() ([]Element, error)          // Elements can contain more elements (Like sources or albums)
	Path() string                          // Returns the absolute path of the element, but not the filesystem path
	Container() bool                       // Returns whether an element can contain other elements or not
	Name() string                          // The name that is shown to the user
	URLName() string                       // The name/identifier used in URLs
	Traverse(path string) (Element, error) // Traverse the element's children with the given path
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
		if element.Container() {
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

// ElementPath returns the absolute path of the element.
// This will not return the file system path, but the path that an object can be addressed inside of galago.
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
