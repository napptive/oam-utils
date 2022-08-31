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
  name: nginx-app
  annotations:
    version: v1.0.0
    description: "Customized version of nginx"
spec:
  components:
    - name: nginx
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
  name: nginx-app1
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
  name: nginx-app2
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
  name: app
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

var _ = ginkgo.Describe("Handler test on log calls", func() {

	ginkgo.It("Should be able to return the application name", func() {
		files := [][]byte{[]byte(fileWithWorkflow)}
		app, err := NewApplication(files)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(app).ShouldNot(gomega.BeNil())

		newName := "changed"
		// Set Name
		app.SetName(newName)

		name := app.GetName()
		gomega.Expect(name).Should(gomega.Equal(name))

		conversion, err := app.ToYAML()
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(conversion).ShouldNot(gomega.BeEmpty())

	})

	ginkgo.It("Should be able to return the application name when the file contains several applications", func() {
		files := [][]byte{[]byte(ComposedFile)}
		app, err := NewApplication(files)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(app).ShouldNot(gomega.BeNil())

		name := app.GetName()
		gomega.Expect(name).ShouldNot(gomega.BeEmpty())
		gomega.Expect(name).Should(gomega.Equal("nginx-app1"))
	})

	ginkgo.It("Should be able to return the application name receiving two files", func() {
		files := [][]byte{[]byte(filewithoutApplication), []byte(fileWithWorkflow)}
		app, err := NewApplication(files)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(app).ShouldNot(gomega.BeNil())

		name := app.GetName()
		gomega.Expect(name).ShouldNot(gomega.BeEmpty())
		gomega.Expect(name).Should(gomega.Equal("app"))
	})

	ginkgo.It("Should be able to return the application name in a multiple YAML file", func() {
		files := [][]byte{[]byte(applicationFile)}
		app, err := NewApplication(files)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(app).ShouldNot(gomega.BeNil())

		newName := "changed"
		// Set Name
		app.SetName(newName)

		name := app.GetName()
		gomega.Expect(name).Should(gomega.Equal(name))

		conversion, err := app.ToYAML()
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(conversion).ShouldNot(gomega.BeEmpty())
	})

	ginkgo.It("should not be able to create an applicacion when it does not exist", func() {
		files := [][]byte{[]byte(filewithoutApplication)}
		_, err := NewApplication(files)
		gomega.Expect(err).ShouldNot(gomega.Succeed())
	})
})
