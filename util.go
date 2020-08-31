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

import "strings"

// ExtToMIME returns the MIME media type of a given file extension.
// Example: ".jpg" returns "image/jpeg".
func ExtToMIME(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	}

	return "application/octet-stream"
}
