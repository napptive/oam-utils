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
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"

	yamlv3 "gopkg.in/yaml.v3"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

// splitYAMLFile returns a list o YAMLs from a multi resource YAML file
func splitYAMLFile(file []byte) ([][]byte, error) {

	resources := make([][]byte, 0)
	decoder := yaml.NewDocumentDecoder(io.NopCloser(bytes.NewReader(file)))
	defer decoder.Close()

	for {
		b := make([]byte, len(file))

		n, err := decoder.Read(b)
		if err != nil && err != io.EOF {
			log.Error().Err(err).Msg("error reading yaml file")
			return nil, nerrors.NewInternalError("error reading yaml file")
		}
		if n == 0 || err == io.EOF {
			return resources, nil
		} else {
			resources = append(resources, b[0:n])
		}
	}

}

// convertUnstructured converts an *unstructured.Unstructured into the struct received
func convertFromUnstructured(unsObj *unstructured.Unstructured, converted interface{}) error {
	to, err := unsObj.MarshalJSON()
	if err != nil {
		log.Error().Err(err).Msg("error marshalling struct")
		return nerrors.NewInternalError("error converting struct")
	}
	if err = json.Unmarshal(to, &converted); err != nil {
		log.Error().Err(err).Msg("error unmarshalling struct")
		return nerrors.NewInternalError("error converting struct")
	}
	return nil
}

func isYAMLFile(fileName string) bool {
	return strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml")
}

// convertToYAML receives an interface (entity) and return its yaml representation
func convertToYAML(entry interface{}) ([]byte, error) {
	jsonStr, err := json.Marshal(entry)
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

	return returned, nil
}

// validateType returns true if the file has the same GroupVersionKind as one received
func validateType(inputType *schema.GroupVersionKind, types []schema.GroupVersionKind) bool {
	// - Decode YAML manifest into unstructured.Unstructured

	for _, appType := range types {
		if appType.Kind == inputType.Kind && appType.Version == inputType.Version && appType.Group == inputType.Group {
			return true
		}
	}
	return false
}

// getGVK returns the group version kind from a YAML file
func getGVK(entity []byte) (*schema.GroupVersionKind, *unstructured.Unstructured, error) {
	// - Decode YAML manifest into unstructured.Unstructured
	var decUnstructured = k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	unsObj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode(entity, nil, unsObj)
	if err != nil {
		log.Error().Err(err).Msg("error getting GVK from an entity")
		return nil, nil, nerrors.NewInternalError("%s", err.Error())
	}
	return gvk, unsObj, nil
}

// getGVKType converts a Group Version Kind to EntityType
func getGVKType(gvk *schema.GroupVersionKind) EntityType {
	if validateType(gvk, applicationGVK) {
		return EntityType_APP
	}
	if validateType(gvk, metadataGKV) {
		return EntityType_METADATA
	}
	return EntityType_UNKNOWN
}
