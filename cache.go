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
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/nfnt/resize"
	"golang.org/x/image/bmp"
	"gopkg.in/yaml.v2"
)

// Cache manages the on disk cache for metadata and image files.
type Cache struct {
	dirPath string
}

var cache *Cache

// NewCache creates a new cache for image data.
func NewCache(path string) *Cache {
	return &Cache{
		dirPath: path,
	}
}

// QueryCacheEntry returns a cache entry for a given hash, or an error if there is no cache element.
func (c *Cache) QueryCacheEntry(hash string) (*CacheEntry, error) {
	data, err := ioutil.ReadFile(filepath.Join(c.dirPath, fmt.Sprintf("%v.yaml", hash)))
	if err != nil {
		return nil, err
	}

	var ce *CacheEntry
	err = yaml.Unmarshal([]byte(data), &ce)
	if err != nil {
		return nil, err
	}

	return ce, nil
}

// StoreCacheEntry saves a cache entry to disk.
func (c *Cache) StoreCacheEntry(hash string, ce *CacheEntry) error {
	data, err := yaml.Marshal(ce)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(c.dirPath, fmt.Sprintf("%v.yaml", hash)), data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// QueryImage returns the image file for a given hash, or an error if there is no file.
func (c *Cache) QueryImage(hash string) (*os.File, error) {
	f, err := os.Open(filepath.Join(c.dirPath, fmt.Sprintf("%v.jpg", hash)))
	if err != nil {
		return nil, err
	}

	return f, nil
}

// StoreImage saves the given image to disk.
func (c *Cache) StoreImage(hash string, img image.Image) error {
	f, err := os.Create(filepath.Join(c.dirPath, fmt.Sprintf("%v.jpg", hash)))
	if err != nil {
		return err
	}
	defer f.Close()

	err = jpeg.Encode(f, img, nil)
	if err != nil {
		return err
	}

	return nil
}

// PrepareAndStoreImage takes an image file, prepares it for caching and writes it into the cache.
func (c Cache) PrepareAndStoreImage(imgElement Image) (*CacheEntry, error) {
	hash := imgElement.Hash()

	// Rely on the fact that ImageSizeOriginal should not be cached.
	file, _, err := imgElement.FileContent(ImageSizeOriginal)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get original image from %v: %w", imgElement, err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("Couldn't decode image %v: %w", imgElement, err)
	}

	imgReduced := resize.Resize(0, 1080, img, resize.Lanczos3)
	imgNano := resize.Resize(0, 8, img, resize.Lanczos3)

	if err := cache.StoreImage(hash, imgReduced); err != nil {
		return nil, fmt.Errorf("Couldn't store image %v to cache: %w", imgElement, err)
	}

	imgNanoBuf := new(bytes.Buffer)

	// Encode uses a Writer, use a Buffer if you need the raw []byte
	if err := bmp.Encode(imgNanoBuf, imgNano); err != nil {
		return nil, fmt.Errorf("Couldn't encode image %v as nano BMP: %w", imgElement, err)
	}

	ce := &CacheEntry{
		NanoBitmap: imgNanoBuf.String(),
		Width:      img.Bounds().Dx(),
		Height:     img.Bounds().Dy(),
	}

	if err := c.StoreCacheEntry(hash, ce); err != nil {
		log.Warnf("Couldn't store cache entry for image %v: %v", imgElement, err)
	}

	return ce, nil
}

// CacheEntry contains the metadata of an image.
type CacheEntry struct {
	Width, Height int

	NanoBitmap string // Byteslice of a BMP file containing a really small version of the image
}
