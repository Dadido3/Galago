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

// SourceType represents a type of a source.
type SourceType struct {
	create func(parent Element, index int, rlName string, c map[string]interface{}) (Element, error) // Create an instance of a source element.
}

// SourceTypes contains all possible source types.
var SourceTypes = map[string]SourceType{}

func registerSourceType(name string, create func(parent Element, index int, urlName string, c map[string]interface{}) (Element, error)) {
	SourceTypes[name] = SourceType{
		create: create,
	}
}
