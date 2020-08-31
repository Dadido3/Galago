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
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"

	"github.com/Dadido3/configdb"
	"github.com/coreos/go-semver/semver"
	"github.com/gorilla/mux"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
)

var log = logrus.New()
var version = semver.Must(semver.NewVersion("0.1.0"))
var conf = configdb.NewOrPanic([]configdb.Storage{
	configdb.UseYAMLFile(filepath.Join(".", "config", "config.yaml")),
})
var router = mux.NewRouter()
var validExtensions = map[string]bool{".jpg": true, ".jpeg": true}

func main() {

	// Logging
	os.MkdirAll(filepath.Join(".", "log"), os.ModePerm)
	rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   filepath.Join(".", "log", "newest.log"),
		MaxSize:    50, // megabytes
		MaxBackups: 3,
		MaxAge:     365, //days
		Level:      logrus.TraceLevel,
		Formatter: &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				//return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)
				return fmt.Sprintf("%s():%d", f.Function, f.Line), ""
			},
		},
	})

	if err != nil {
		log.Fatalf("Failed to initialize file rotate hook: %v", err)
	}

	log.SetReportCaller(true)
	log.AddHook(rotateFileHook)
	log.SetLevel(logrus.TraceLevel)
	log.SetOutput(colorable.NewColorableStdout())
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			//return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)
			return fmt.Sprintf("%s():%d", f.Function, f.Line), ""
		},
	})

	var cachePath string
	if err := conf.Get(".Cache.Path", &cachePath); err != nil {
		log.Fatalf("Can't load cache path from config files: %v", err)
	}
	cache = NewCache(cachePath)

	// Add routes to the webserver
	serverUIInit()

	log.Infof("galago %v started", version)

	var addr string
	if err := conf.Get(".Server.ListenAddress", &addr); err != nil {
		log.Fatalf("Can't load server listen address from config files: %v", err)
	}
	log.Infof("Listening at %v", addr)
	log.Fatalf("Webserver returned error: %v", http.ListenAndServe(addr, router))
}
