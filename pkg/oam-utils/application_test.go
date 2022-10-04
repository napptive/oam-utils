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
	"github.com/rs/zerolog/log"
)

const applicationFile = `
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
  name: application
  annotations:
    version: v1.0.0
    description: "Customized version of nginx"
spec:
  components: # comment
    - name: component1
      type: webservice
      properties:
        image: nginx:1.20.0 # Image
        ports:
        - port: 80 # Port
          expose: true
`
const cm = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-test
data:
  cpu: "0.50"
  memory: "250Mi"
`
const fileWithWorkflow = `apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: appWithWorkflow
  annotations: 
    version: "v0.0.1"
    description: "My app"
spec:
  components:
    - name: component1
      type: worker # Required worker
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

// completeApplication with a very complete application (two OAM applications and a cm in the same file)
const completeApplication = `
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
  name: app1
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
---
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: app2
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
---
`

const spec = `
components:
  - name: component1
    type: webservice # Webservice type
    properties:
      image: 'nginx:1.20.0'
      ports:
        - port: 82
          expose: true
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
		ginkgo.Context("Setting new name", func() {
			ginkgo.It("Should be able to apply parameters (name)", func() {
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
			ginkgo.It("Should not be able to apply parameters (name) in a non existing application", func() {
				files := []*ApplicationFile{{FileName: "metadata.yaml", Content: []byte(metadata)}}
				app, err := NewApplication(files)
				gomega.Expect(err).Should(gomega.Succeed())
				gomega.Expect(app).ShouldNot(gomega.BeNil())

				newName := "changed"
				err = app.ApplyParameters("appWithWorkflow", newName, "")
				gomega.Expect(err).ShouldNot(gomega.Succeed())
			})
			ginkgo.It("Should not be able to apply parameters (name) in a wrong existing application", func() {
				files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(fileWithWorkflow)}}
				app, err := NewApplication(files)
				gomega.Expect(err).Should(gomega.Succeed())
				gomega.Expect(app).ShouldNot(gomega.BeNil())

				newName := "changed"
				err = app.ApplyParameters("error", newName, "")
				gomega.Expect(err).ShouldNot(gomega.Succeed())
			})
		})
		ginkgo.Context("Setting component spec", func() {
			ginkgo.It("Show be able to apply parameters (Components spec)", func() {
				files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(applicationFile)}}
				app, err := NewApplication(files)
				gomega.Expect(err).Should(gomega.Succeed())
				gomega.Expect(app).ShouldNot(gomega.BeNil())

				err = app.ApplyParameters("application", "", spec)
				gomega.Expect(err).Should(gomega.Succeed())

			})
			ginkgo.It("Show be able to apply parameters (Components spec) in an application with workload specs", func() {
				files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(fileWithWorkflow)}}
				app, err := NewApplication(files)
				gomega.Expect(err).Should(gomega.Succeed())
				gomega.Expect(app).ShouldNot(gomega.BeNil())

				err = app.ApplyParameters("appWithWorkflow", "", spec)
				gomega.Expect(err).Should(gomega.Succeed())

				apps, _, err := app.ToYAML()
				gomega.Expect(err).Should(gomega.Succeed())
				gomega.Expect(apps).ShouldNot(gomega.BeNil())

				data := string(apps[0])
				gomega.Expect(data).ShouldNot(gomega.BeEmpty())

				parameters, err := app.GetParameters()
				gomega.Expect(err).Should(gomega.Succeed())
				gomega.Expect(parameters).ShouldNot(gomega.BeNil())

				log.Info().Str("application", string(apps[0])).Msg("application YAML")

			})
			ginkgo.It("Show not be able to apply parameters if the application does not exists", func() {
				files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(applicationFile)}}
				app, err := NewApplication(files)
				gomega.Expect(err).Should(gomega.Succeed())
				gomega.Expect(app).ShouldNot(gomega.BeNil())

				err = app.ApplyParameters("error", "", spec)
				gomega.Expect(err).ShouldNot(gomega.Succeed())

			})
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
			err = app.ApplyParameters("application", newName, "")
			gomega.Expect(err).Should(gomega.Succeed())

			name := app.GetNames()
			gomega.Expect(name["application"]).Should(gomega.Equal(newName))

			apps, entities, err := app.ToYAML()
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(apps).ShouldNot(gomega.BeEmpty())
			gomega.Expect(entities).ShouldNot(gomega.BeEmpty())
		})
	})

	ginkgo.Context("Getting parameters", func() {
		ginkgo.It("Should be able to return the parameters of a catalog application with two applications", func() {
			files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(completeApplication)}}
			app, err := NewApplication(files)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(app).ShouldNot(gomega.BeNil())

			parameters, err := app.GetParameters()
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(parameters).ShouldNot(gomega.BeEmpty())
			gomega.Expect(len(parameters)).Should(gomega.Equal(2))
		})
		ginkgo.It("Should be able to return an empty map of parameters it the catalog application has no oam applications", func() {
			files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(cm)}}
			app, err := NewApplication(files)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(app).ShouldNot(gomega.BeNil())

			parameters, err := app.GetParameters()
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(parameters).Should(gomega.BeEmpty())
		})

		ginkgo.It("FULL Test", func() {
			files := []*ApplicationFile{{FileName: "file1.yaml", Content: []byte(fileWithWorkflow)}}
			app, err := NewApplication(files)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(app).ShouldNot(gomega.BeNil())

			params, err := app.GetConfigurations()
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(params).ShouldNot(gomega.BeNil())

			err = app.ApplyParameters("appWithWorkflow", "", params["appWithWorkflow"].ComponentSpec)
			gomega.Expect(err).Should(gomega.Succeed())

			yaml, _, err := app.ToYAML()
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(yaml).ShouldNot(gomega.BeNil())

			log.Info().Str("application", string(yaml[0])).Msg("application YAML")

		})
	})

})
