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
	"github.com/Dadido3/configdb"
	"github.com/Dadido3/configdb/tree"
)

// RootElement contains the root element of the tree.
var RootElement = &Album{}

func loadSources() {
	// Initialize sources and register callback for config changes
	conf.RegisterCallback([]string{".Sources"}, func(c *configdb.Config, modified, added, removed []string) {
		var sourcesConf tree.Node
		if err := c.Get(".Sources", &sourcesConf); err != nil {
			log.Errorf("Error while reading configuration file: %v", err)
			return
		}

		// Initialize and reset the sources
		RootElement = &Album{}

		for urlName := range sourcesConf {
			var sourceConf tree.Node
			if err := sourcesConf.Get("."+urlName, &sourceConf); err != nil {
				log.Warnf("Error while reading configuration file: %v", err)
				continue
			}

			var sourceTypeName string
			if err := sourceConf.Get(".Type", &sourceTypeName); err != nil {
				log.Warnf("Error while reading configuration file: %v", err)
				continue
			}

			var sourceType SourceType
			var ok bool
			if sourceType, ok = SourceTypes[sourceTypeName]; !ok {
				log.Warnf("Unknown source type %q of %q", sourceTypeName, urlName)
				continue
			}

			// Create new source of the given type and forward its configuration
			index := len(RootElement.children)
			if sourceInstance, err := sourceType.create(RootElement, index, urlName, sourceConf); err == nil {
				RootElement.children = append(RootElement.children, sourceInstance)
				log.Debugf("Created new instance %q of source %q", urlName, sourceTypeName)
			} else {
				log.Errorf("Couldn't create instance %q of source %q: %v", urlName, sourceTypeName, err)
			}
		}

		log.Info("Loaded sources from configuration")
	})
}
