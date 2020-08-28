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

import "fmt"

// Album can contain other elements
type Album struct {
	name, urlName string
	parent        Element
	children      []Element
}

// Parent returns the parent element, duh.
func (a Album) Parent() Element {
	return a.parent
}

// Children returns the content of the album.
func (a Album) Children() ([]Element, error) {
	return a.children, nil
}

// Path returns the absolute path of the element, but not the filesystem path.
// For details see ElementPath.
func (a Album) Path() string {
	return ElementPath(a)
}

// Container returns whether an element can contain other elements or not.
func (a Album) Container() bool {
	return true
}

// Name returns the name that is shown to the user.
func (a Album) Name() string {
	return a.name
}

// URLName returns the name/identifier that is used in URLs.
func (a Album) URLName() string {
	return a.urlName
}

// Traverse the element's children with the given path.
func (a Album) Traverse(path string) (Element, error) {
	return TraverseElements(a, path)
}

// FilterAlbums takes a list of elements, and returns only the albums.
// The content of the albums stays untouched.
func FilterAlbums(ee []Element) []*Album {
	result := []*Album{}
	for _, element := range ee {
		if album, ok := element.(*Album); ok {
			result = append(result, album)
		}
	}
	return result
}

func (a Album) String() string {
	return fmt.Sprintf("{Album %q: %v children}", a.Path(), len(a.children))
}
