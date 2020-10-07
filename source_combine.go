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

	"github.com/Dadido3/configdb/tree"
)

func init() {
	registerSourceType("combine", CreateSourceCombine)
}

// SourceCombine represents a source that combines all given internal paths to a single source.
type SourceCombine struct {
	parent        Element
	index         int
	name, urlName string
	internalPaths []string
	hidden        bool
	home          bool
	sourceTags    *SourceTags
}

// Compile time check if SourceCombine implements Element.
var _ Element = (*SourceCombine)(nil)

// CreateSourceCombine returns a new instance of the source.
func CreateSourceCombine(parent Element, index int, urlName string, c tree.Node) (Element, error) {
	var name string
	if err := c.Get(".Name", &name); err != nil {
		return nil, fmt.Errorf("Configuration of source %q errornous: %w", urlName, err)
	}

	hidden, _ := c["Hidden"].(bool)
	home, _ := c["Home"].(bool)

	var paths []string
	if err := c.Get(".InternalPaths", &paths); err != nil {
		return nil, fmt.Errorf("Configuration of source %q errornous: %w", urlName, err)
	}

	s := &SourceCombine{
		parent:        parent,
		index:         index,
		name:          name,
		urlName:       urlName,
		internalPaths: paths,
		hidden:        hidden,
		home:          home,
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
func (s *SourceCombine) Clone(parent Element, index int) Element {
	clone := *s

	clone.parent = parent
	clone.index = index

	return &clone
}

// Parent returns the parent element, duh.
func (s *SourceCombine) Parent() Element {
	return s.parent
}

// Index returns the index of the element in its parent children list.
func (s *SourceCombine) Index() int {
	return s.index
}

// Children returns the folders and images of a source.
func (s *SourceCombine) Children() ([]Element, error) {
	elements := []Element{}

	// Place tag list as the first child
	if s.sourceTags != nil {
		elements = append(elements, s.sourceTags)
	}

	// Output clones of all elements defined in internalPaths
	for _, internalPath := range s.internalPaths {
		element, err := RootElement.Traverse(internalPath)
		if err != nil {
			log.Warnf("Internal path %q not found: %v", internalPath, err)
			continue
		}

		elements = append(elements, element.Clone(s, len(elements)))
	}

	return elements, nil
}

// Path returns the absolute path of the element, but not the filesystem path.
// For details see ElementPath.
func (s *SourceCombine) Path() string {
	return ElementPath(s)
}

// IsContainer returns whether an element can contain other elements or not.
func (s *SourceCombine) IsContainer() bool {
	return true
}

// IsHidden returns whether this element can be listed as child or not.
func (s *SourceCombine) IsHidden() bool {
	return s.hidden
}

// IsHome returns whether an element should be linked by the home button or not.
func (s *SourceCombine) IsHome() bool {
	return s.home
}

// Name returns the name that is shown to the user.
func (s *SourceCombine) Name() string {
	return s.name
}

// URLName returns the name/identifier that is used in URLs.
func (s *SourceCombine) URLName() string {
	return s.urlName
}

// Traverse the element's children with the given path.
func (s *SourceCombine) Traverse(path string) (Element, error) {
	return TraverseElements(s, path)
}

func (s *SourceCombine) String() string {
	return fmt.Sprintf("{SourceCombine %q: %v}", s.Path(), s.internalPaths)
}
