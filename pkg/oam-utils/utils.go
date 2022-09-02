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

// convert converts an *unstructured.Unstructured into the struct received
func convert(unsObj *unstructured.Unstructured, converted interface{}) error {
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

func getGVK(entity []byte) (*schema.GroupVersionKind, *unstructured.Unstructured, error) {
	// - Decode YAML manifest into unstructured.Unstructured
	var decUnstructured = k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	unsObj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode(entity, nil, unsObj)
	if err != nil {
		log.Error().Err(err).Msg("error getting GVK from an entity")
		return nil, nil, nerrors.NewInternalError("error getting GVK from an entity")
	}
	return gvk, unsObj, nil
}

func getGVKType(gvk *schema.GroupVersionKind) EntityType {
	if validateType(gvk, applicationGVK) {
		return EntityType_APP
	}
	if validateType(gvk, metadataGKV) {
		return EntityType_METADATA
	}
	return EntityType_UNKNOWN
}
