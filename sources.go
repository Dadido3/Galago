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

import "github.com/Dadido3/configdb"

// RootElement contains the root element of the tree.
var RootElement = Album{}

func init() {
	// Initialize sources and register callback for config changes
	conf.RegisterCallback([]string{".Sources"}, func(c *configdb.Config, modified, added, removed []string) {
		var sourcesConf map[string]map[string]interface{}
		if err := c.Get(".Sources", &sourcesConf); err != nil {
			log.Errorf("Error while reading configuration file: %v", err)
			return
		}

		// Initialize and reset the sources
		RootElement = Album{}

		for urlName, sourceConf := range sourcesConf {
			var s interface{}
			var ok bool
			if s, ok = sourceConf["Type"]; !ok {
				log.Errorf("Undefined source type of source entry %q", urlName)
				continue
			}

			var source string
			if source, ok = s.(string); !ok {
				log.Errorf("Unexpected type of the \"Type\" field of %q. Got %T, expected %T", urlName, s, source)
				continue
			}

			var sourceType SourceType
			if sourceType, ok = SourceTypes[source]; !ok {
				log.Warnf("Unknown source type %q of %q", source, urlName)
				continue
			}

			// Create new source of the given type and forward its configuration
			index := len(RootElement.children)
			if sourceInstance, err := sourceType.create(RootElement, index, urlName, sourceConf); err == nil {
				RootElement.children = append(RootElement.children, sourceInstance)
				log.Debugf("Created new instance %q of source %q", urlName, source)
			} else {
				log.Errorf("Couldn't create instance %q of source %q: %v", urlName, source, err)
			}
		}

		log.Info("Loaded sources from configuration")
	})
}
