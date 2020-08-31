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
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
)

var uiTemplates *template.Template

func init() {
	uiTemplates = template.New("").Funcs(template.FuncMap{
		"filterAlbums":     FilterAlbums,
		"filterImages":     FilterImages,
		"filterNonEmpty":   FilterNonEmpty,
		"filterContainers": FilterContainers,
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

	d := struct {
		RootElement Album
		Path        string
	}{
		RootElement: RootElement,
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

	log.Debug(r.URL.Path)

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

	imageFile, mime, err := image.FileContent(ImageSizeOriginal)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer imageFile.Close()

	w.Header().Set("Content-Type", mime)

	io.Copy(w, imageFile)
}

func serverUIInit() {
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(".", "ui", "static")))))

	//router.Handle("/login", auth.LoginHandler(storage.StorageSessions, storage.StorageUsers))
	//router.Handle("/logout", auth.LogoutHandler(storage.StorageSessions))

	router.PathPrefix("/image/").Handler(http.StripPrefix("/image/", &uiImage{}))

	router.Handle("/", http.StripPrefix("/", newUITemplate("gallery.gohtml")))
	router.PathPrefix("/gallery/").Handler(http.StripPrefix("/gallery/", newUITemplate("gallery.gohtml")))
}
