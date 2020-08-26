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
	Children() ([]Element, error)          // Elements can contain more elements (Like sources or albums)
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

// TraverseElements will traverse through the children of elements until the element at path is reached.
// An example for a path: root/album1/album2/img.jpg
func TraverseElements(origin Element, path string) (Element, error) {
	pathElements := strings.Split(path, "/")

	// If the path is empty, return the current object
	if path == "" {
		return origin, nil
	}

	children, err := origin.Children()
	if err != nil {
		return nil, err
	}
	for _, child := range children {
		if child.URLName() == pathElements[0] {
			return TraverseElements(child, strings.Join(pathElements[1:], "/"))
		}
	}

	return nil, fmt.Errorf("No matching element found for given path")
}
