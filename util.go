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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"
)

// ExtToMIME returns the MIME media type of a given file extension.
// Example: ".jpg" returns "image/jpeg".
func ExtToMIME(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".bmp":
		return "image/bmp"
	}

	return "application/octet-stream"
}

// ImageToDataURI takes the result from FileContent and returns an data URI that can be embedded into HTML or CSS.
// This will close the stream f.
func ImageToDataURI(img Image) (string, error) {
	f, _, mime, err := img.FileContent(ImageSizeNano)
	if err != nil {
		return "", fmt.Errorf("Couldn't get file of image %v: %w", img, err)
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("Couldn't read file content of image %v: %w", img, err)
	}

	return fmt.Sprintf("data:%v;base64,%v", mime, base64.StdEncoding.EncodeToString(buf)), nil
}
