/*
Copyright 2022 Napptive

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package oam_utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"

	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
	yamlv3 "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

type Metadata struct {
	// Name of the resource
	Name string `json:"name"`
	// Annotations of the resource.
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels related to the resource.
	Labels map[string]string `json:"labels,omitempty"`
}

type ApplicationComponent struct {
	Name       string                `json:"name"`
	Type       string                `json:"type"`
	Properties *runtime.RawExtension `json:"properties,omitempty"`
}

type AppPolicy struct {
	Name       string                `json:"name"`
	Type       string                `json:"type"`
	Properties *runtime.RawExtension `json:"properties,omitempty"`
}

type ApplicationSpec struct {
	Components *runtime.RawExtension `json:"components"`
	Policies   []AppPolicy           `json:"policies,omitempty"`
	Workflow   *runtime.RawExtension `json:"workflow,omitempty"`
}

type ApplicationDefinition struct {
	ApiVersion string          `json:"apiVersion"`
	Kind       string          `json:"kind"`
	Metadata   Metadata        `json:"metadata"`
	Spec       ApplicationSpec `json:"spec"`
}

// Application with an oam application
type Application struct {
	// App with a map of applications definition indexed by application the name
	apps map[string]*ApplicationDefinition
	// obj with map of the application stored as unstructured indexed by the name
	objs map[string]*unstructured.Unstructured
	// entities with an array of other entities
	entities [][]byte
}

// NewApplicationFromTGZ receives a tgz file and returns convert the content into an application
func NewApplicationFromTGZ(rawApplication []byte) (*Application, error) {
	files := make([]*ApplicationFile, 0)

	br := bytes.NewReader(rawApplication)
	uncompressedStream, err := gzip.NewReader(br)
	if err != nil {
		log.Error().Err(err).Msg("error creating applciation from tgz")
		return nil, nerrors.NewInternalErrorFrom(err, "error creating application")
	}
	tarReader := tar.NewReader(uncompressedStream)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error().Err(err).Msg("error creating applciation from tgz")
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
	objs := make(map[string]*unstructured.Unstructured, 0)
	entities := [][]byte{}

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
			// check if the YAML contains an application
			isApp, app, err := isApplication(entity)
			if err != nil {
				log.Error().Err(err).Str("File", file.FileName).Msg("error getting the application file")
				return nil, nerrors.NewInternalErrorFrom(err, "error creating application, error in file: %s", file.FileName)
			}
			if *isApp {
				var appDefinition ApplicationDefinition
				if err := convert(app, &appDefinition); err != nil {
					log.Error().Err(err).Str("File", file.FileName).Msg("error converting application")
					return nil, nerrors.NewInternalErrorFrom(err, "error creating application")
				}
				apps[appDefinition.Metadata.Name] = &appDefinition
				objs[appDefinition.Metadata.Name] = app
			} else {
				entities = append(entities, entity)
			}
		}
	}
	if len(apps) == 0 {

		log.Error().Msg("Error creating application, no application received")
		return nil, nerrors.NewNotFoundError("error creating application, no application found")
	}

	return &Application{
		apps:     apps,
		objs:     objs,
		entities: entities,
	}, nil
}

// isApplication returns a boolean indicating if the entity received is an application
// if it is, return it in an unstructured
func isApplication(entity []byte) (*bool, *unstructured.Unstructured, error) {
	isApp := false
	// - Decode YAML manifest into unstructured.Unstructured
	var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	unsObj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode(entity, nil, unsObj)
	if err != nil {
		log.Warn().Err(err).Msg("error checking if the file contains an application, might not contain be an entity")
		return &isApp, nil, nil
	}

	isApp = gvk.Group == applicationGVK.Group && gvk.Kind == applicationGVK.Kind && gvk.Version == applicationGVK.Version

	return &isApp, unsObj, nil
}

// GetName returns the application name
func (a *Application) GetNames() map[string]string {

	names := make(map[string]string, 0)
	for name, application := range a.apps {
		names[name] = application.Metadata.Name
	}
	return names
}

// ApplyParameters overwrite the application name and the components spec in application named `applicationName`
// TODO: implement ComponentSpec management
func (a *Application) ApplyParameters(applicationName string, newName string, componentsSpec string) error {

	// check if the applicacion exists
	app, exists := a.apps[applicationName]
	if !exists {
		return nerrors.NewNotFoundError("application %s not found", applicationName)
	}
	if newName != "" {
		app.Metadata.Name = newName
	}

	return nil
}

// ToYAML converts the application in YAML
func (a *Application) ToYAML() ([]byte, error) {

	var yamlFile []byte
	separator := ("\n---\n")
	for _, app := range a.apps {
		jsonStr, err := json.Marshal(app)
		if err != nil {
			log.Error().Err(err).Msg("error parsing to JSON")
			return nil, nerrors.NewInternalError("error converting to JSON")
		}

		// Convert the JSON to an object.
		var jsonObj interface{}
		err = yamlv3.Unmarshal(jsonStr, &jsonObj)
		if err != nil {
			log.Error().Err(err).Msg("error in Unmarshal ")
			return nil, nerrors.NewInternalError("error converting to YAML")
		}

		// Marshal this object into YAML.
		returned, err := yamlv3.Marshal(jsonObj)
		if err != nil {
			log.Error().Err(err).Msg("error in Marshal ")
			return nil, nerrors.NewInternalError("error converting to YAML")
		}

		yamlFile = append(yamlFile, returned...)
		yamlFile = append(yamlFile, []byte(separator)...)

	}

	// TODO: add entities

	return yamlFile, nil

}
