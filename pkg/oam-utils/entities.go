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
	"github.com/napptive/nerrors/pkg/nerrors"
	"github.com/rs/zerolog/log"
	yamlV3 "gopkg.in/yaml.v3"
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

// ApplicationDefinition with the definition of an OAM application
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

// copyComponents returns an ApplicationSpec without any field except Components
func (as *ApplicationSpec) copyComponents() *ApplicationSpec {
	return &ApplicationSpec{
		Components: as.Components,
	}
}

// ComponentsNode with the components Spec in YAML (with comments)
type ComponentsNode struct {
	Spec ComponentsYAML
}

func (cn *ComponentsNode) toYAML() (string, error) {
	data, err := yamlV3.Marshal(&ComponentsYAML{Components: cn.Spec.Components})
	if err != nil {
		log.Error().Err(err).Msg("error converting to YAML")
		return "", nerrors.NewInternalError("error converting to YAML")
	}
	return string(data), nil
}

// ComponentsYAML with the components in YAML (the array of components)
type ComponentsYAML struct {
	Components yamlV3.Node
}

// getComponentsNode returns a ComponentNode from a YAML file
func getComponentsNode(app []byte) (*ComponentsNode, error) {
	var node ComponentsNode

	if err := yamlV3.Unmarshal(app, &node); err != nil {
		log.Error().Err(err).Msg("Error creating components node")
		return nil, nerrors.NewInternalError("Error creating componentsNode")
	}
	return &node, nil
}

func getComponentsYAML(app []byte) (*ComponentsYAML, error) {
	var node ComponentsYAML

	if err := yamlV3.Unmarshal(app, &node); err != nil {
		log.Error().Err(err).Msg("Error creating components yaml")
		return nil, nerrors.NewInternalError("Error creating ComponentsYAML")
	}
	return &node, nil
}
