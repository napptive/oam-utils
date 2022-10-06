/*
 * Copyright 2022 Napptive
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * https://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package oam_utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"strings"

	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Application with a catalog application
type Application struct {
	// App with a map of OAM applications indexed by application the name
	apps map[string]*ApplicationDefinition
	// obj with map of the OAM applications stored as unstructured indexed by the name
	// this struct always have the initial values, it is no been updated when setting parameters
	// objs map[string]*unstructured.Unstructured
	// entities with an array of other entities
	entities [][]byte
	// componentsYAML with the components YAML spec (with comments) indexed by applicationName
	componentsYAML map[string]*ComponentsNode
}

type InstanceConf struct {
	// Name with the application name
	Name string
	// ComponentSpec with the component specification
	ComponentSpec string
}

// NewApplicationFromTGZ receives a tgz file and returns convert the content into an application
func NewApplicationFromTGZ(rawApplication []byte) (*Application, error) {
	files := make([]*ApplicationFile, 0)

	br := bytes.NewReader(rawApplication)
	uncompressedStream, err := gzip.NewReader(br)
	if err != nil {
		log.Error().Err(err).Msg("error creating application from tgz")
		return nil, nerrors.NewInternalErrorFrom(err, "error creating application")
	}
	tarReader := tar.NewReader(uncompressedStream)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error().Err(err).Msg("error creating application from tgz")
			return nil, nerrors.NewInternalErrorFrom(err, "error creating application")
		}

		switch header.Typeflag {
		case tar.TypeDir:
			log.Debug().Str("name", header.Name).Msg("is a directory")
		case tar.TypeReg:
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, nerrors.NewInternalErrorFrom(err, "error creating application, error reading %s file", header.Name)
			}
			files = append(files, &ApplicationFile{
				FileName: header.Name,
				Content:  data,
			})
		default:
			log.Warn().Str("type", string(header.Typeflag)).Msg("ignoring compressed type")
		}
	}
	return NewApplication(files)
}

// NewApplication converts an oam application from an array of yaml files into an Application
func NewApplication(files []*ApplicationFile) (*Application, error) {

	apps := make(map[string]*ApplicationDefinition, 0)
	//objs := make(map[string]*unstructured.Unstructured, 0)
	nodes := make(map[string]*ComponentsNode, 0)
	var entities [][]byte

	for _, file := range files {

		// check if the file is a yaml File
		if !isYAMLFile(file.FileName) {
			log.Info().Str("file", file.FileName).Msg("skipping the file")
			continue
		}

		resources, err := splitYAMLFile([]byte(file.Content))
		if err != nil {
			log.Error().Err(err).Str("File", file.FileName).Msg("error split application file")
			return nil, nerrors.NewInternalErrorFrom(err, "cannot create application, error in file: %s", file.FileName)
		}

		for _, entity := range resources {

			gvk, app, err := getGVK(entity)
			if err != nil {
				// YAML file without GVK is not a oam or kubernetes entity, not stored.
				log.Warn().Str("File", file.FileName).Msg("yaml file without GVK")
				continue
			}
			switch getGVKType(gvk) {
			// Application
			case EntityType_APP:
				var appDefinition ApplicationDefinition
				if err := convertFromUnstructured(app, &appDefinition); err != nil {
					log.Error().Err(err).Str("File", file.FileName).Msg("error converting application")
					return nil, nerrors.NewInternalErrorFrom(err, "error creating application")
				}
				apps[appDefinition.Metadata.Name] = &appDefinition

				node, err := getComponentsNodeFromYAML(entity)
				if err != nil {
					log.Error().Err(err).Str("File", file.FileName).Msg("error creating application")
					return nil, nerrors.NewInternalErrorFrom(err, "error creating application")
				}
				nodes[appDefinition.Metadata.Name] = node

				// Metadata
			case EntityType_METADATA:
				log.Debug().Str("file", file.FileName).Msg("is metadata file")
				// Others
			default:
				entities = append(entities, entity)
			}

		}
	}
	// a catalog application might not contain oam application.
	// For example, if a user wants to store their component definitions
	if len(apps) == 0 {
		log.Warn().Msg("Error creating application, no application received")
	}

	log.Debug().Int("apps", len(apps)).Int("entities", len(entities)).Msg("Apps configuration")

	return &Application{
		apps:           apps,
		entities:       entities,
		componentsYAML: nodes,
	}, nil
}

// GetNames returns the application name
func (a *Application) GetNames() map[string]string {

	names := make(map[string]string, 0)
	for name, application := range a.apps {
		names[name] = application.Metadata.Name
	}
	return names
}

// GetParameters returns the components spec of an application indexed by application name
func (a *Application) GetParameters() (map[string]string, error) {

	parameters := make(map[string]string, 0)
	for appName, components := range a.componentsYAML {
		appParameters, err := components.toYAML()
		if err != nil {
			log.Error().Err(err).Str("appName", appName).Msg("error converting to YAML")
			return nil, nerrors.NewInternalError("error getting the parameters of %s application", appName)
		}
		parameters[appName] = appParameters
	}

	return parameters, nil
}

// GetConfigurations return the name and the componentSpec by application
func (a *Application) GetConfigurations() (map[string]*InstanceConf, error) {
	configurations := make(map[string]*InstanceConf, 0)
	for appName, components := range a.componentsYAML {
		appParameters, err := components.toYAML()
		if err != nil {
			log.Error().Err(err).Str("appName", appName).Msg("error converting to YAML")
			return nil, nerrors.NewInternalError("error getting the parameters of %s application", appName)
		}
		configurations[appName] = &InstanceConf{
			Name:          appName,
			ComponentSpec: appParameters,
		}
	}
	return configurations, nil
}

// ApplyParameters overwrite the application name and the components spec in application named `applicationName`
func (a *Application) ApplyParameters(applicationName string, newName string, newAppSpec string) error {

	if len(a.apps) == 0 {
		return nerrors.NewNotFoundError("there is no applications to apply parameters")
	}

	// check if the application exists
	app, exists := a.apps[applicationName]
	if !exists {
		return nerrors.NewNotFoundError("application %s not found", applicationName)
	}
	if newName != "" {
		app.Metadata.Name = newName
	}
	if newAppSpec != "" {
		spec, err := a.toApplicationSpec(newAppSpec)
		if err != nil {
			return nerrors.NewInternalError("Unable to apply parameters: %s", err.Error())
		}
		app.Spec.Components = spec.Components
	}

	// update nodes
	node, err := getComponentsFromYAML([]byte(newAppSpec))
	if err != nil {
		return nerrors.NewInternalErrorFrom(err, "error creating application")
	}
	if node != nil {
		a.componentsYAML[applicationName] = &ComponentsNode{
			Spec: *node,
		}
	}

	return nil
}

// ToYAML converts the application in YAML
func (a *Application) ToYAML() ([][]byte, [][]byte, error) {

	var appsFiles [][]byte
	for _, app := range a.apps {
		// Marshal this object into YAML.
		returned, err := convertToYAML(app)
		if err != nil {
			log.Error().Err(err).Msg("error in Marshal ")
			return nil, nil, nerrors.NewInternalError("error converting to YAML")
		}

		appsFiles = append(appsFiles, returned)
	}

	return appsFiles, a.entities, nil
}

// toApplicationSpec convert a YAML to ApplicationSpec
func (a *Application) toApplicationSpec(spec string) (*ApplicationSpec, error) {

	reader := strings.NewReader(spec)
	d := yaml.NewYAMLOrJSONDecoder(reader, 4096)

	ext := ApplicationSpec{}
	if err := d.Decode(&ext); err != nil {
		log.Error().Err(err).Str("spec", spec).Msg("error in toRawExtension")
		return nil, nerrors.NewInternalError("Error processing %s", err.Error())
	}
	return &ext, nil
}
