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
)

type Metadata struct {
	// Name of the resource
	Name string `json:"name"`
	// Annotations of the resource.
	Annotations map[string]string `json:"annotations,omitempty"`
	// Labels related to the resource.
	Labels map[string]string `json:"labels,omitempty"`
}

// AppPolicy with the application policy
type AppPolicy struct {
	// Name of the policy
	Name string `json:"name"`
	// Type of the policy
	Type string `json:"type"`
	// Properties of the policy
	Properties *runtime.RawExtension `json:"properties,omitempty"`
}

// ApplicationSpec with the application specification
type ApplicationSpec struct {
	// Components of the applcication
	Components *runtime.RawExtension `json:"components"`
	// Policies of the application
	Policies []AppPolicy `json:"policies,omitempty"`
	// Workflow with the workflowsteps of the application
	Workflow *runtime.RawExtension `json:"workflow,omitempty"`
}

// ApplicationDefinnition with the definition of an OAM application
type ApplicationDefinition struct {
	// ApiVersion
	ApiVersion string `json:"apiVersion"`
	// Kind
	Kind string `json:"kind"`
	// Metadata
	Metadata Metadata `json:"metadata"`
	// Spec
	Spec ApplicationSpec `json:"spec"`
}

// Application with an catalog application
type Application struct {
	// App with a map of OAM applications indexed by application the name
	apps map[string]*ApplicationDefinition
	// obj with map of the OAM applications stored as unstructured indexed by the name
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
				if err := convert(app, &appDefinition); err != nil {
					log.Error().Err(err).Str("File", file.FileName).Msg("error converting application")
					return nil, nerrors.NewInternalErrorFrom(err, "error creating application")
				}
				apps[appDefinition.Metadata.Name] = &appDefinition
				objs[appDefinition.Metadata.Name] = app
				// Metadata
			case EntityType_METADATA:
				log.Debug().Str("file", file.FileName).Msg("is metadata file")
				// Others
			default:
				entities = append(entities, entity)
			}

		}
	}
	// TODO: Creo q no deberíamos soltar este error ahora q tb guardamos las entidades (ComponentDefinition, p.e.)
	if len(apps) == 0 {
		log.Error().Msg("Error creating application, no application received")
		//return nil, nerrors.NewNotFoundError("error creating application, no application found")
	}

	return &Application{
		apps:     apps,
		objs:     objs,
		entities: entities,
	}, nil
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
func (a *Application) ToYAML() ([][]byte, [][]byte, error) {

	var appsFiles [][]byte
	for _, app := range a.apps {
		jsonStr, err := json.Marshal(app)
		if err != nil {
			log.Error().Err(err).Msg("error parsing to JSON")
			return nil, nil, nerrors.NewInternalError("error converting to JSON")
		}

		// Convert the JSON to an object.
		var jsonObj interface{}
		err = yamlv3.Unmarshal(jsonStr, &jsonObj)
		if err != nil {
			log.Error().Err(err).Msg("error in Unmarshal ")
			return nil, nil, nerrors.NewInternalError("error converting to YAML")
		}

		// Marshal this object into YAML.
		returned, err := yamlv3.Marshal(jsonObj)
		if err != nil {
			log.Error().Err(err).Msg("error in Marshal ")
			return nil, nil, nerrors.NewInternalError("error converting to YAML")
		}

		appsFiles = append(appsFiles, returned)

	}

	return appsFiles, a.entities, nil

}
