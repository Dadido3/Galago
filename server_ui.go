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
	"archive/zip"
	"compress/flate"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/coreos/go-semver/semver"
)

var uiTemplates *template.Template

var isAlphanumeric = regexp.MustCompile(`^[0-9A-Za-z]+$`).MatchString

func init() {
	uiTemplates = template.New("").Funcs(template.FuncMap{
		"filterAlbums":     FilterAlbums,
		"filterImages":     FilterImages,
		"filterNonEmpty":   FilterNonEmpty,
		"filterContainers": FilterContainers,
		"imageToDataURI":   ImageToDataURI,
		"previousElement":  PreviousElement,
		"nextElement":      NextElement,
		"getPreviewImages": GetPreviewImages,
	})

	uiTemplates = template.Must(uiTemplates.ParseGlob(filepath.Join(".", "ui", "templates", "*.*html")))
	uiTemplates = template.Must(uiTemplates.ParseGlob(filepath.Join(".", "ui", "templates", "webcomponents", "*.html")))
}

type uiTemplate struct {
	Template *template.Template

	filename string
}

func newUITemplate(filename string) *uiTemplate {
	clone := template.Must(uiTemplates.Clone())
	template := template.Must(clone.ParseFiles(filepath.Join(".", "ui", "templates", "pages", filename)))
	return &uiTemplate{Template: template, filename: filename}
}

func (t *uiTemplate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//clone := template.Must(uiTemplates.Clone())
	//clone := template.Must(template.ParseGlob(filepath.Join(".", "ui", "templates", "*.*html")))
	//clone = template.Must(clone.ParseGlob(filepath.Join(".", "ui", "templates", "webcomponents", "*.html")))
	//t.Template = template.Must(clone.ParseFiles(filepath.Join(".", "ui", "templates", "pages", t.filename))) // TODO: Disable "debug" template parsing on each request

	//path := mux.Vars(r)["path"]

	log.Trace(r.URL.Path)

	d := struct {
		RootElement Album
		Version     *semver.Version
		Path        string
	}{
		RootElement: RootElement,
		Version:     version,
		Path:        r.URL.Path,
	}

	if err := t.Template.ExecuteTemplate(w, "base.gohtml", d); err != nil {
		err = fmt.Errorf("Error executing template %q: %w", "base.gohtml", err)
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type uiImage struct{}

func (t *uiImage) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Trace(r.URL.Path)

	element, err := RootElement.Traverse(r.URL.Path)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusNotFound) // Assume the error is because the element was not found
		return
	}

	image, ok := element.(Image)
	if !ok {
		err := fmt.Errorf("Element %v is not an image", element)
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	imageFile, size, mime, err := image.FileContent()
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer imageFile.Close()

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.Header().Set("Cache-Control", "public, max-age=86400") // 1 Day

	io.Copy(w, imageFile)

	log.Tracef("Sent original image %v", element.Name())
}

type uiCachedImage struct{}

func (t *uiCachedImage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Path

	// Make sure only alphanumeric hashes can be queried
	if !isAlphanumeric(hash) {
		log.Errorf("Invalid request. Tried to query cache element with hash %q", hash)
		http.Error(w, "The hash can only be alphanumeric", http.StatusBadRequest)
		return
	}

	ce, err := cache.QueryCacheEntryHash(hash)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	f, size, mime, err := ce.ReducedImage()
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.Header().Set("Cache-Control", "public, max-age=2419200") // 4 weeks

	io.Copy(w, f)

	log.Tracef("Sent cached reduced image %q", r.URL.Path)
}

type uiDownload struct{}

func (t *uiDownload) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	element, err := RootElement.Traverse(r.URL.Path)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Handle download of single image files
	if img, ok := element.(Image); ok {
		r, size, mime, err := img.FileContent()
		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Close()

		w.Header().Set("Content-Disposition", "attachment; filename="+element.URLName())
		w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
		w.Header().Set("Content-Type", mime)

		io.Copy(w, r)
		return
	}

	// Handle download of containers. This will pack all images contained in element into an archive and stream it to the browser.
	if element.Container() {
		w.Header().Set("Content-Disposition", "attachment; filename="+element.Name()+".zip")
		w.Header().Set("Content-Type", "application/zip")

		zw := zip.NewWriter(w)

		zw.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
			return flate.NewWriter(out, flate.NoCompression)
		})

		children, err := element.Children()
		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, child := range children {
			if cImg, ok := child.(Image); ok {
				cw, err := zw.Create(child.URLName())
				if err != nil {
					log.Error(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				cr, _, _, err := cImg.FileContent()
				if err != nil {
					log.Error(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer cr.Close()

				if _, err := io.Copy(cw, cr); err != nil {
					log.Error(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		if err := zw.Close(); err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}

	http.Error(w, "Requested file is neither an image nor any other supported element.", http.StatusBadRequest)
}

func serverUIInit() {
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(".", "ui", "static")))))

	//router.Handle("/login", auth.LoginHandler(storage.StorageSessions, storage.StorageUsers))
	//router.Handle("/logout", auth.LogoutHandler(storage.StorageSessions))

	router.PathPrefix("/image/").Handler(http.StripPrefix("/image/", &uiImage{}))
	router.PathPrefix("/cached/").Handler(http.StripPrefix("/cached/", &uiCachedImage{}))
	router.PathPrefix("/download/").Handler(http.StripPrefix("/download/", &uiDownload{}))

	router.Handle("/", http.StripPrefix("/", newUITemplate("gallery.gohtml")))
	router.PathPrefix("/gallery/").Handler(http.StripPrefix("/gallery/", newUITemplate("gallery.gohtml")))
	router.PathPrefix("/image-viewer/").Handler(http.StripPrefix("/image-viewer/", newUITemplate("image-viewer.gohtml")))
}
