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
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

const applicationFile = (`
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-test
data:
  cpu: "0.50"
  memory: "250Mi"
---
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: appapplication
  annotations:
    version: v1.0.0
    description: "Customized version of nginx"
spec:
  components:
    - name: component1
      type: webservice
      properties:
        image: nginx:1.20.0
        ports:
        - port: 80
          expose: true
`)
const twoApplicationFile = (`
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: app1
  annotations:
    version: v1.0.0
    description: "Customized version of nginx"
spec:
  components:
    - name: nginx1
      type: webservice
      properties:
        image: nginx:1.20.0
        ports:
        - port: 80
          expose: true
---
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: app2
  annotations:
    version: v1.0.0
    description: "Customized version of nginx"
spec:
  components:
    - name: nginx2
      type: webservice
      properties:
        image: nginx:1.20.0
        ports:
        - port: 80
          expose: true
`)
const cm = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-test
data:
  cpu: "0.50"
  memory: "250Mi"
`
const fileWithWorkflow = `
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: appWithWorkflow
  annotations: 
    version: "v0.0.1"
    description: "My app"
spec:
  components:
    - name: component1
      type: worker
      properties:
        image: busybox
        cmd: ["sleep", "86400"]
    - name: component2
      type: worker
      properties:
        image: busybox
        cmd: ["sleep", "86400"]
      traits:
        - type: scaler
          properties:
            replicas: 1
  workflow:
    steps:
    - name: apply-app
      type: apply-application-in-parallel
`
const metadata = `
apiVersion: core.napptive.com/v1alpha1
kind: ApplicationMetadata
`
const readme = `
# README file
`

var _ = ginkgo.Describe("Handler test on log calls", func() {

	ginkgo.Context("Creating application", func() {
		ginkgo.It("Should be able to create an application without other entities", func() {
			files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(fileWithWorkflow)}}
			app, err := NewApplication(files)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(app).ShouldNot(gomega.BeNil())
		})
		ginkgo.It("Should be able to create an application with entities", func() {
			files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(fileWithWorkflow)},
				{FileName: "cm.yaml", Content: []byte(cm)}}
			app, err := NewApplication(files)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(app).ShouldNot(gomega.BeNil())
		})
		ginkgo.PIt("Should be able to create an application only with other entities???", func() {

		})
		ginkgo.It("Should be able to create an application with all the required files", func() {
			files := []*ApplicationFile{
				{FileName: "file1.yaml", Content: []byte(fileWithWorkflow)},
				{FileName: "metadata.yaml", Content: []byte(metadata)},
				{FileName: "readme.md", Content: []byte(readme)}}
			app, err := NewApplication(files)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(app).ShouldNot(gomega.BeNil())
		})

	})

	ginkgo.Context("Getting names", func() {
		ginkgo.It("Should be able to get application names", func() {
			files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(fileWithWorkflow)}}
			app, err := NewApplication(files)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(app).ShouldNot(gomega.BeNil())

			names := app.GetNames()
			gomega.Expect(names).ShouldNot(gomega.BeNil())
			gomega.Expect(names["appWithWorkflow"]).Should(gomega.Equal("appWithWorkflow"))

		})
		ginkgo.It("Should be able to receive an empty map of names if there is no application", func() {
			files := []*ApplicationFile{{FileName: "metadata.yaml", Content: []byte(metadata)}}
			app, err := NewApplication(files)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(app).ShouldNot(gomega.BeNil())

			names := app.GetNames()
			gomega.Expect(names).Should(gomega.BeEmpty())
		})

	})

	ginkgo.Context("Applying parameters", func() {
		ginkgo.It("Should be able to apply parameters", func() {
			files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(fileWithWorkflow)}}
			app, err := NewApplication(files)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(app).ShouldNot(gomega.BeNil())

			newName := "changed"
			err = app.ApplyParameters("appWithWorkflow", newName, "")
			gomega.Expect(err).Should(gomega.Succeed())

			names := app.GetNames()
			gomega.Expect(names).ShouldNot(gomega.BeNil())
			gomega.Expect(names["appWithWorkflow"]).Should(gomega.Equal(newName))

		})
		ginkgo.It("Should not be able to apply parameters in a non existing application", func() {
			files := []*ApplicationFile{{FileName: "metadata.yaml", Content: []byte(metadata)}}
			app, err := NewApplication(files)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(app).ShouldNot(gomega.BeNil())

			newName := "changed"
			err = app.ApplyParameters("appWithWorkflow", newName, "")
			gomega.Expect(err).ShouldNot(gomega.Succeed())
		})
		ginkgo.It("Should not be able to apply components in a wrong existing application", func() {
			files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(fileWithWorkflow)}}
			app, err := NewApplication(files)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(app).ShouldNot(gomega.BeNil())

			newName := "changed"
			err = app.ApplyParameters("error", newName, "")
			gomega.Expect(err).ShouldNot(gomega.Succeed())
		})
	})

	ginkgo.Context("Generating YAML", func() {
		ginkgo.It("Should be able to get application yaml", func() {
			files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(applicationFile)}}
			app, err := NewApplication(files)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(app).ShouldNot(gomega.BeNil())

			newName := "changed"
			// Set Name
			err = app.ApplyParameters("appapplication", newName, "")
			gomega.Expect(err).Should(gomega.Succeed())

			name := app.GetNames()
			gomega.Expect(name["appapplication"]).Should(gomega.Equal(newName))

			apps, entities, err := app.ToYAML()
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(apps).ShouldNot(gomega.BeEmpty())
			gomega.Expect(entities).ShouldNot(gomega.BeEmpty())
		})
	})

})
