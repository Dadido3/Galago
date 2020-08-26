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

import "image"

// Image represents an image object.
// Sources are supposed to implement this type.
type Image interface {
	LoadImage() image.Image
}

// FilterImages takes a list of elements, and returns only the images.
func FilterImages(ee []Element) []Image {
	result := []Image{}
	for _, element := range ee {
		if image, ok := element.(Image); ok {
			result = append(result, image)
		}
	}
	return result
}
