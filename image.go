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
	"io"
)

// Image references an image file stored in a source.
// Sources are supposed to return images that implement this type.
type Image interface {
	Hash() string // Unique hash that stays the same as long as the file doesn't change. Must be alphanumeric
	Width() int   // Width of the original image
	Height() int  // Height of the original image

	FileContent() (r io.ReadCloser, size int64, mime string, err error) // Returns the original image file
	CacheEntry() (*CacheEntry, error)                                   // Returns the cache entry of the image
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
