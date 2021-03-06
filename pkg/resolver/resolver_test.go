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

package resolver

import (
	"testing"

	"k8s.io/helm/pkg/chart"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name   string
		req    []*chart.Dependency
		expect *chart.Lock
		err    bool
	}{
		{
			name: "version failure",
			req: []*chart.Dependency{
				{Name: "oedipus-rex", Repository: "http://example.com", Version: ">a1"},
			},
			err: true,
		},
		{
			name: "cache index failure",
			req: []*chart.Dependency{
				{Name: "oedipus-rex", Repository: "http://example.com", Version: "1.0.0"},
			},
			err: true,
		},
		{
			name: "chart not found failure",
			req: []*chart.Dependency{
				{Name: "redis", Repository: "http://example.com", Version: "1.0.0"},
			},
			err: true,
		},
		{
			name: "constraint not satisfied failure",
			req: []*chart.Dependency{
				{Name: "alpine", Repository: "http://example.com", Version: ">=1.0.0"},
			},
			err: true,
		},
		{
			name: "valid lock",
			req: []*chart.Dependency{
				{Name: "alpine", Repository: "http://example.com", Version: ">=0.1.0"},
			},
			expect: &chart.Lock{
				Dependencies: []*chart.Dependency{
					{Name: "alpine", Repository: "http://example.com", Version: "0.2.0"},
				},
			},
		},
		{
			name: "repo from valid local path",
			req: []*chart.Dependency{
				{Name: "signtest", Repository: "file://../../../../cmd/helm/testdata/testcharts/signtest", Version: "0.1.0"},
			},
			expect: &chart.Lock{
				Dependencies: []*chart.Dependency{
					{Name: "signtest", Repository: "file://../../../../cmd/helm/testdata/testcharts/signtest", Version: "0.1.0"},
				},
			},
		},
		{
			name: "repo from invalid local path",
			req: []*chart.Dependency{
				{Name: "notexist", Repository: "file://../testdata/notexist", Version: "0.1.0"},
			},
			err: true,
		},
	}

	repoNames := map[string]string{"alpine": "kubernetes-charts", "redis": "kubernetes-charts"}
	r := New("testdata/chartpath", "testdata/helmhome")
	for _, tt := range tests {
		hash, err := HashReq(tt.req)
		if err != nil {
			t.Fatal(err)
		}

		l, err := r.Resolve(tt.req, repoNames, hash)
		if err != nil {
			if tt.err {
				continue
			}
			t.Fatal(err)
		}

		if tt.err {
			t.Fatalf("Expected error in test %q", tt.name)
		}

		if h, err := HashReq(tt.req); err != nil {
			t.Fatal(err)
		} else if h != l.Digest {
			t.Errorf("%q: hashes don't match.", tt.name)
		}

		// Check fields.
		if len(l.Dependencies) != len(tt.req) {
			t.Errorf("%s: wrong number of dependencies in lock", tt.name)
		}
		d0 := l.Dependencies[0]
		e0 := tt.expect.Dependencies[0]
		if d0.Name != e0.Name {
			t.Errorf("%s: expected name %s, got %s", tt.name, e0.Name, d0.Name)
		}
		if d0.Repository != e0.Repository {
			t.Errorf("%s: expected repo %s, got %s", tt.name, e0.Repository, d0.Repository)
		}
		if d0.Version != e0.Version {
			t.Errorf("%s: expected version %s, got %s", tt.name, e0.Version, d0.Version)
		}
	}
}

func TestHashReq(t *testing.T) {
	expect := "sha256:d661820b01ed7bcf26eed8f01cf16380e0a76326ba33058d3150f919d9b15bc0"
	req := []*chart.Dependency{
		{Name: "alpine", Version: "0.1.0", Repository: "http://localhost:8879/charts"},
	}
	h, err := HashReq(req)
	if err != nil {
		t.Fatal(err)
	}
	if expect != h {
		t.Errorf("Expected %q, got %q", expect, h)
	}

	req = []*chart.Dependency{}
	h, err = HashReq(req)
	if err != nil {
		t.Fatal(err)
	}
	if expect == h {
		t.Errorf("Expected %q !=  %q", expect, h)
	}
}
