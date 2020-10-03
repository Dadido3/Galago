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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/nfnt/resize"
	"golang.org/x/image/bmp"
	"gopkg.in/yaml.v2"

	"trimmer.io/go-xmp/models/dc"
	xmpbase "trimmer.io/go-xmp/models/xmp_base"
	"trimmer.io/go-xmp/xmp"
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

// QueryCacheEntryHash returns a cache entry for a given hash, or an error if there is no cache element.
func (c *Cache) QueryCacheEntryHash(hash string) (*CacheEntry, error) {
	data, err := ioutil.ReadFile(filepath.Join(c.dirPath, fmt.Sprintf("%v.yaml", hash)))
	if err != nil {
		return nil, err
	}

	var ce *CacheEntry
	err = yaml.Unmarshal([]byte(data), &ce)
	if err != nil {
		return nil, err
	}

	ce.cache = c
	ce.hash = hash

	return ce, nil
}

// QueryCacheEntryImage returns a cache entry for a given image element, or an error if there is no cache element.
//
// When there is no cache entry, a new one based on the image will be generated.
// This function will block and then return a valid cache entry, if one could be generated.
// An error will be returned otherwise.
func (c *Cache) QueryCacheEntryImage(img Image) (*CacheEntry, error) {
	hash := img.Hash()

	// Return an already existing cache entry if possible
	if ce, err := c.QueryCacheEntryHash(hash); err == nil {
		return ce, nil
	}

	// Generate new cache entry, and return it if possible
	return cache.PrepareAndStoreImage(img)
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

// PrepareAndStoreImage takes an image file, prepares it for caching and writes it into the cache.
func (c *Cache) PrepareAndStoreImage(imgElement Image) (*CacheEntry, error) {
	hash := imgElement.Hash()

	// Rely on the fact that ImageSizeOriginal should not be cached
	file, _, _, err := imgElement.FileContent()
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

	// Encode uses a Writer, use a Buffer if you need the raw []byte
	imgNanoBuf := new(bytes.Buffer)
	if err := bmp.Encode(imgNanoBuf, imgNano); err != nil {
		return nil, fmt.Errorf("Couldn't encode image %v as nano BMP: %w", imgElement, err)
	}

	ce := &CacheEntry{
		cache:      c,
		hash:       hash,
		NanoBitmap: imgNanoBuf.String(),
		Width:      img.Bounds().Dx(),
		Height:     img.Bounds().Dy(),
	}

	if err := ce.SetReducedImage(hash, imgReduced); err != nil {
		return nil, fmt.Errorf("Couldn't store image %v to cache: %w", imgElement, err)
	}

	// Get metadata
	file, _, _, err = imgElement.FileContent()
	if err != nil {
		return nil, fmt.Errorf("Couldn't get original image from %v: %w", imgElement, err)
	}
	defer file.Close()

	if d, err := xmp.Scan(file); err == nil {
		// Retrieve some values from the XMP namespace
		xmpNS := d.FindNs("xmp", "http://ns.adobe.com/xap/1.0/")
		if xmpModel, ok := d.FindModel(xmpNS).(*xmpbase.XmpBase); ok {
			// Rating
			if rating, err := xmpModel.GetTag("Rating"); err == nil {
				if ratingInt, err := strconv.ParseInt(rating, 10, 0); err == nil {
					ce.Rating = int(ratingInt)
				}
			}
		}

		// Retrieve some values from the DC namespace
		dcNS := d.FindNs("dc", "http://purl.org/dc/elements/1.1/")
		if dcModel, ok := d.FindModel(dcNS).(*dc.DublinCore); ok {
			// Title
			ce.Title = dcModel.Title.Default()
		}

	} else {
		log.Warnf("Couldn't read and parse metadata from %v: %v", imgElement, err)
	}

	// Store cache entry
	if err := c.StoreCacheEntry(hash, ce); err != nil {
		log.Warnf("Couldn't store cache entry for image %v: %v", imgElement, err)
	}

	return ce, nil
}

// CacheEntry contains the metadata of an image.
type CacheEntry struct {
	cache *Cache
	hash  string // The hash of the cache entry

	Title         string // Title based on metadata
	Rating        int    // -1: Rejected, 0: Unrated, 1-5: Rated
	Width, Height int

	NanoBitmap string // Byteslice of a BMP file containing a really small version of the image
}

// ReducedImagePath returns the filepath to the reduced version of the image.
func (ce *CacheEntry) ReducedImagePath() string {
	if ce.cache == nil {
		return ""
	}
	return filepath.Join(ce.cache.dirPath, fmt.Sprintf("%v.jpg", ce.hash))
}

// SetReducedImage saves the image as reduced version to the disk.
func (ce *CacheEntry) SetReducedImage(hash string, img image.Image) error {
	if ce.cache == nil {
		return fmt.Errorf("Cache entry doesn't contain valid pointer to cache")
	}

	f, err := os.Create(ce.ReducedImagePath())
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

// ReducedImage returns the reduced version of the cached image.
func (ce *CacheEntry) ReducedImage() (r io.ReadCloser, size int64, mime string, err error) {
	if ce.cache == nil {
		return nil, 0, "", fmt.Errorf("Cache entry doesn't contain valid pointer to cache")
	}

	f, err := os.Open(ce.ReducedImagePath())
	if err != nil {
		return nil, 0, "", err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, 0, "", err
	}
	return f, stat.Size(), "image/jpeg", err
}

// NanoImage returns a really small version of the cached image.
//
// It's only a few pixels in legnth and width to be suitable for embedding.
func (ce *CacheEntry) NanoImage() (r io.ReadCloser, size int64, mime string, err error) {
	r = ioutil.NopCloser(bytes.NewReader([]byte(ce.NanoBitmap)))
	return r, int64(len(ce.NanoBitmap)), "image/bmp", nil
}
