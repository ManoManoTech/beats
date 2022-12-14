// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// +build mage

package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"
	// register kubernetes runner
	// mage:import
	// mage:import
	// mage:import
	// mage:import
	// mage:import
	// mage:import
	// mage:import
	// mage:import
	// mage:import

	"github.com/magefile/mage/mg"

	devtools "github.com/elastic/beats/v7/dev-tools/mage"
	_ "github.com/elastic/beats/v7/dev-tools/mage/kubernetes"
	"github.com/elastic/beats/v7/dev-tools/mage/target/build"
	"github.com/elastic/beats/v7/dev-tools/mage/target/common"
	_ "github.com/elastic/beats/v7/dev-tools/mage/target/compose"
	_ "github.com/elastic/beats/v7/dev-tools/mage/target/dashboards"
	_ "github.com/elastic/beats/v7/dev-tools/mage/target/docs"
	_ "github.com/elastic/beats/v7/dev-tools/mage/target/pkg"
	"github.com/elastic/beats/v7/dev-tools/mage/target/test"
	"github.com/elastic/beats/v7/dev-tools/mage/target/unittest"
	metricbeat "github.com/elastic/beats/v7/metricbeat/scripts/mage"
	_ "github.com/elastic/beats/v7/metricbeat/scripts/mage/target/metricset"
)

func init() {
	common.RegisterCheckDeps(Update)
	test.RegisterDeps(IntegTest)
	unittest.RegisterGoTestDeps(Fields)
	unittest.RegisterPythonTestDeps(Fields)

	devtools.BeatDescription = "Metricbeat is a lightweight shipper for metrics."
}

// CollectAll generates the docs and the fields.
func CollectAll() {
	mg.Deps(CollectDocs, FieldsDocs)
}

// Package packages the Beat for distribution.
// Use SNAPSHOT=true to build snapshots.
// Use PLATFORMS to control the target platforms.
// Use VERSION_QUALIFIER to control the version qualifier.
func Package() {
	start := time.Now()
	defer func() { fmt.Println("package ran for", time.Since(start)) }()

	devtools.UseElasticBeatOSSPackaging()
	metricbeat.CustomizePackaging()
	devtools.PackageKibanaDashboardsFromBuildDir()

	mg.Deps(Update)
	mg.Deps(build.CrossBuild, build.CrossBuildGoDaemon)
	mg.SerialDeps(devtools.Package, TestPackages)
}

// TestPackages tests the generated packages (i.e. file modes, owners, groups).
func TestPackages() error {
	return devtools.TestPackages(
		devtools.WithModulesD(),
		devtools.WithModules(),

		// To be increased or removed when more light modules are added
		devtools.MinModules(1),
	)
}

// Dashboards collects all the dashboards and generates index patterns.
func Dashboards() error {
	return devtools.KibanaDashboards("module")
}

// Config generates both the short and reference configs.
func Config() {
	mg.Deps(configYML, metricbeat.GenerateDirModulesD)
}

func configYML() error {
	return devtools.Config(devtools.AllConfigTypes, metricbeat.OSSConfigFileParams(), ".")
}

// MockedTests runs the HTTP tests using the mocked data inside each {module}/{metricset}/testdata folder.
// Use MODULE={module_name} to run only mocked tests with a single module.
// Use GENERATE=true or GENERATE=1 to regenerate JSON files.
func MockedTests(ctx context.Context) error {
	params := devtools.DefaultGoTestUnitArgs()

	params.ExtraFlags = []string{"github.com/elastic/beats/v7/metricbeat/mb/testing/data/."}

	if module := os.Getenv("MODULE"); module != "" {
		params.ExtraFlags = append(params.ExtraFlags, "-module="+module)
	}

	if generate, _ := strconv.ParseBool(os.Getenv("GENERATE")); generate {
		params.ExtraFlags = append(params.ExtraFlags, "-data")
	}

	params.Packages = nil

	return devtools.GoTest(ctx, params)
}

// AssembleDarwinUniversal merges the darwin/amd64 and darwin/arm64 into a single
// universal binary using `lipo`. It assumes the darwin/amd64 and darwin/arm64
// were built and only performs the merge.
func AssembleDarwinUniversal() error {
	return build.AssembleDarwinUniversal()
}

// Fields generates a fields.yml and fields.go for each module.
func Fields() {
	mg.Deps(fieldsYML, moduleFieldsGo)
}

func fieldsYML() error {
	return devtools.GenerateFieldsYAML("module")
}

func moduleFieldsGo() error {
	return devtools.GenerateModuleFieldsGo("module")
}

// Update is an alias for running fields, dashboards, config.
func Update() {
	mg.SerialDeps(
		Fields, Dashboards, Config, CollectAll,
		metricbeat.PrepareModulePackagingOSS,
		metricbeat.GenerateOSSMetricbeatModuleIncludeListGo)
}

// FieldsDocs generates docs/fields.asciidoc containing all fields
// (including x-pack).
func FieldsDocs() error {
	inputs := []string{
		devtools.OSSBeatDir("module"),
		devtools.XPackBeatDir("module"),
	}
	output := devtools.CreateDir("build/fields/fields.all.yml")
	if err := devtools.GenerateFieldsYAMLTo(output, inputs...); err != nil {
		return err
	}
	return devtools.Docs.FieldDocs(output)
}

// CollectDocs creates the documentation under docs/
func CollectDocs() error {
	return metricbeat.CollectDocs()
}

// IntegTest executes integration tests (it uses Docker to run the tests).
func IntegTest() {
	mg.SerialDeps(GoIntegTest, PythonIntegTest)
}

// GoIntegTest executes the Go integration tests.
// Use TEST_COVERAGE=true to enable code coverage profiling.
// Use RACE_DETECTOR=true to enable the race detector.
// Use TEST_TAGS=tag1,tag2 to add additional build tags.
// Use MODULE=module to run only tests for `module`.
func GoIntegTest(ctx context.Context) error {
	if !devtools.IsInIntegTestEnv() {
		mg.SerialDeps(Fields, Dashboards)
	}
	return devtools.GoTestIntegrationForModule(ctx)
}

// PythonIntegTest executes the python system tests in the integration
// environment (Docker).
// Use MODULE=module to run only tests for `module`.
// Use PYTEST_ADDOPTS="-k pattern" to only run tests matching the specified pattern.
// Use any other PYTEST_* environment variable to influence the behavior of pytest.
func PythonIntegTest(ctx context.Context) error {
	if !devtools.IsInIntegTestEnv() {
		mg.SerialDeps(Fields, Dashboards)
	}
	runner, err := devtools.NewDockerIntegrationRunner(devtools.ListMatchingEnvVars("PYTEST_")...)
	if err != nil {
		return err
	}
	return runner.Test("pythonIntegTest", func() error {
		mg.Deps(devtools.BuildSystemTestBinary)
		return devtools.PythonTestForModule(devtools.DefaultPythonTestIntegrationArgs())
	})
}
// Build builds the Beat binary.
func Build() error {
	return devtools.Build(devtools.DefaultBuildArgs())
}

// GolangCrossBuild build the Beat binary inside of the golang-builder.
// Do not use directly, use crossBuild instead.
func GolangCrossBuild() error {
	return devtools.GolangCrossBuild(devtools.DefaultGolangCrossBuildArgs())
}

// BuildGoDaemon builds the go-daemon binary (use crossBuildGoDaemon).
func BuildGoDaemon() error {
	return devtools.BuildGoDaemon()
}

// CrossBuild cross-builds the beat for all target platforms.
func CrossBuild() error {
	return devtools.CrossBuild()
}

// CrossBuildGoDaemon cross-builds the go-daemon binary using Docker.
func CrossBuildGoDaemon() error {
	return devtools.CrossBuildGoDaemon()
}