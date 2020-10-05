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
	"github.com/Dadido3/configdb/tree"
)

// SourceType represents a type of a source.
type SourceType struct {
	create func(parent Element, index int, rlName string, c tree.Node) (Element, error) // Create an instance of a source element.
}

// SourceTypes contains all possible source types.
var SourceTypes = map[string]SourceType{}

func registerSourceType(name string, create func(parent Element, index int, urlName string, c tree.Node) (Element, error)) {
	SourceTypes[name] = SourceType{
		create: create,
	}
}
