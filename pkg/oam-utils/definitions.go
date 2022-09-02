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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type EntityType uint

const (
	EntityType_UNKNOWN EntityType = iota
	EntityType_APP
	EntityType_METADATA
)

// applicationGVK with application GVK
var applicationGVK = []schema.GroupVersionKind{{
	Group:   "core.oam.dev",
	Version: "v1alpha2",
	Kind:    "ApplicationConfiguration",
}, {
	Group:   "core.oam.dev",
	Version: "v1beta1",
	Kind:    "Application",
}}

// metadataTypes with entities that contain information about the application itself.
var metadataGKV = []schema.GroupVersionKind{
	{
		Group:   "core.napptive.com",
		Version: "v1alpha1",
		Kind:    "ApplicationMetadata",
	}, {
		Group:   "core.oam.dev",
		Version: "v1alpha1",
		Kind:    "ApplicationMetadata",
	}}

// ApplicationFile with a struct that relates the name of a file to its content
type ApplicationFile struct {
	FileName string
	Content  []byte
}
