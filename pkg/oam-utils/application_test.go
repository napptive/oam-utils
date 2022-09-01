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

const ComposedFile = (`
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

const filewithoutApplication = `
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
This is an example of a metadata file
`
const readme = `
# README file
`

var _ = ginkgo.Describe("Handler test on log calls", func() {

	ginkgo.Context("Creating application", func() {

	})

	ginkgo.Context("Getting names", func() {

	})

	ginkgo.Context("Applying parameters", func() {

	})

	ginkgo.Context("Generating YAML", func() {

	})

	ginkgo.It("Should be able to return the application name", func() {
		files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(fileWithWorkflow)}}
		app, err := NewApplication(files)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(app).ShouldNot(gomega.BeNil())

		newName := "changed"
		// Set Name
		err = app.ApplyParameters("appWithWorkflow", newName, "")
		gomega.Expect(err).Should(gomega.Succeed())

		name := app.GetNames()
		gomega.Expect(len(name)).ShouldNot(gomega.BeZero())
		gomega.Expect(name["appWithWorkflow"]).Should(gomega.Equal(newName))

		apps, entities, err := app.ToYAML()
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(apps).ShouldNot(gomega.BeEmpty())
		gomega.Expect(entities).Should(gomega.BeEmpty())

	})
	ginkgo.It("Should be able to return the application name of a full application", func() {
		files := []*ApplicationFile{{FileName: "file1.md", Content: []byte(readme)},
			{FileName: "file2.txt", Content: []byte(metadata)},
			{FileName: "file3.yaml", Content: []byte(fileWithWorkflow)}}

		app, err := NewApplication(files)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(app).ShouldNot(gomega.BeNil())

	})

	ginkgo.It("Should be able to return the application name when the file contains several applications", func() {
		files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(ComposedFile)}}
		app, err := NewApplication(files)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(app).ShouldNot(gomega.BeNil())

		name := app.GetNames()
		gomega.Expect(name).ShouldNot(gomega.BeEmpty())
		gomega.Expect(name["app1"]).Should(gomega.Equal("app1"))
		gomega.Expect(name["app2"]).Should(gomega.Equal("app2"))
	})

	ginkgo.It("Should be able to return the application name receiving two files", func() {
		files := []*ApplicationFile{
			{FileName: "file1.yaml", Content: []byte(filewithoutApplication)},
			{FileName: "file2.yaml", Content: []byte(fileWithWorkflow)}}

		app, err := NewApplication(files)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(app).ShouldNot(gomega.BeNil())

		names := app.GetNames()
		gomega.Expect(names).ShouldNot(gomega.BeEmpty())
		gomega.Expect(names["appWithWorkflow"]).Should(gomega.Equal("appWithWorkflow"))
	})

	ginkgo.It("Should be able to return the application name in a multiple YAML file", func() {
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

	ginkgo.It("should not be able to create an applicacion when it does not exist", func() {
		files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(filewithoutApplication)}}
		_, err := NewApplication(files)
		gomega.Expect(err).ShouldNot(gomega.Succeed())
	})

	ginkgo.It("Should be able to return the application name when a readme has yaml extension", func() {
		files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(readme)},
			{FileName: "file2.txt", Content: []byte(metadata)},
			{FileName: "file3.yaml", Content: []byte(fileWithWorkflow)}}

		app, err := NewApplication(files)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(app).ShouldNot(gomega.BeNil())

		names := app.GetNames()
		gomega.Expect(names["appWithWorkflow"]).Should(gomega.Equal("appWithWorkflow"))

	})
})
