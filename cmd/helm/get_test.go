/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"testing"

	"k8s.io/helm/pkg/hapi/release"
	"k8s.io/helm/pkg/helm"
)

func TestGetCmd(t *testing.T) {
	tests := []cmdTestCase{{
		name:   "get with a release",
		cmd:    "get thomas-guide",
		golden: "output/get-release.txt",
		rels:   []*release.Release{helm.ReleaseMock(&helm.MockReleaseOptions{Name: "thomas-guide"})},
	}, {
		name:      "get requires release name arg",
		cmd:       "get",
		golden:    "output/get-no-args.txt",
		wantError: true,
	}}
	runTestCmd(t, tests)
}
