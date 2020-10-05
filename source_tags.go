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

	"github.com/Dadido3/configdb/tree"
)

func init() {
	registerSourceType("tags", CreateSourceTags)
}

// SourceTags represents a source that gets all images from a list of elements, and shows them grouped by tags.
// The elements are referenced by their internal path.
//
// Hidden children will not be included.
// To include hidden containers, you need to specify their path explicitly.
type SourceTags struct {
	parent        Element
	index         int
	name, urlName string
	internalPaths []string
	hidden        bool
}

// Compile time check if SourceTags implements Element.
var _ Element = (*SourceTags)(nil)

// CreateSourceTags returns a new instance the source.
func CreateSourceTags(parent Element, index int, urlName string, c tree.Node) (Element, error) {
	var name string
	if err := c.Get(".Name", &name); err != nil {
		return nil, fmt.Errorf("Configuration of source %q errornous: %w", urlName, err)
	}

	hidden, _ := c["Hidden"].(bool)

	var paths []string
	if err := c.Get(".InternalPaths", &paths); err != nil {
		return nil, fmt.Errorf("Configuration of source %q errornous: %w", urlName, err)
	}

	return &SourceTags{
		parent:        parent,
		index:         index,
		name:          name,
		urlName:       urlName,
		internalPaths: paths,
		hidden:        hidden,
	}, nil
}

// Parent returns the parent element, duh.
func (s *SourceTags) Parent() Element {
	return s.parent
}

// Index returns the index of the element in its parent children list.
func (s *SourceTags) Index() int {
	return s.index
}

// Children returns the folders and images of a source.
func (s *SourceTags) Children() ([]Element, error) {
	tags := map[string][]Element{}    // List of tags with their elements
	visited := map[Element]struct{}{} // To prevent duplicates and recursion

	var recursive func(e Element) error
	recursive = func(e Element) error {
		// Ignore self
		if self, ok := e.(*SourceTags); ok && self == s {
			return nil
		}

		// Check for duplicates and prevent recursion
		if _, ok := visited[e]; ok {
			return nil
		}
		visited[e] = struct{}{}

		// If the element is an image, add itself to the correct tag entries
		if img, ok := e.(Image); ok {
			ce, err := img.CacheEntry()
			if err != nil {
				return err
			}
			for _, tag := range ce.Tags {
				if _, ok := tags[tag]; !ok {
					tags[tag] = []Element{}
				}
				tags[tag] = append(tags[tag], e)
			}
		}

		// Check children
		children, err := e.Children()
		if err != nil {
			return err
		}
		for _, child := range FilterNonHidden(children) {
			if err := recursive(child); err != nil {
				return err
			}
		}

		return nil
	}

	// Check all given internal paths recursively
	for _, internalPath := range s.internalPaths {
		element, err := RootElement.Traverse(internalPath)
		if err != nil {
			log.Warnf("Internal path %q not found: %v", internalPath, err)
		}
		if err := recursive(element); err != nil {
			return nil, err
		}
	}

	// Create albums by tag list
	elements := []Element{}
	for tagName, tagElements := range tags {
		elements = append(elements, &Album{
			parent:   s,
			name:     tagName,
			urlName:  tagName,
			index:    len(elements),
			children: tagElements,
		})
	}

	return elements, nil
}

// Path returns the absolute path of the element, but not the filesystem path.
// For details see ElementPath.
func (s *SourceTags) Path() string {
	return ElementPath(s)
}

// IsContainer returns whether an element can contain other elements or not.
func (s *SourceTags) IsContainer() bool {
	return true
}

// IsHidden returns whether this element can be listed as child or not.
func (s *SourceTags) IsHidden() bool {
	return s.hidden
}

// Name returns the name that is shown to the user.
func (s *SourceTags) Name() string {
	return s.name
}

// URLName returns the name/identifier that is used in URLs.
func (s *SourceTags) URLName() string {
	return s.urlName
}

// Traverse the element's children with the given path.
func (s *SourceTags) Traverse(path string) (Element, error) {
	return TraverseElements(s, path)
}

func (s *SourceTags) String() string {
	return fmt.Sprintf("{SourceTags %q: %v}", s.Path(), s.internalPaths)
}
