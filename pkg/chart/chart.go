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

package chart

// APIVersionv1 is the API version number for version 1.
const APIVersionv1 = "v1"

// Chart is a helm package that contains metadata, a default config, zero or more
// optionally parameterizable templates, and zero or more charts (dependencies).
type Chart struct {
	// Metadata is the contents of the Chartfile.
	Metadata *Metadata
	// LocK is the contents of Chart.lock.
	Lock *Lock
	// Templates for this chart.
	Templates []*File
	// TODO Delete RawValues after unit tests for `create` are refactored.
	RawValues []byte
	// Values are default config for this template.
	Values map[string]interface{}
	// Files are miscellaneous files in a chart archive,
	// e.g. README, LICENSE, etc.
	Files []*File

	parent       *Chart
	dependencies []*Chart
}

// SetDependencies replaces the chart dependencies.
func (ch *Chart) SetDependencies(charts ...*Chart) {
	ch.dependencies = nil
	ch.AddDependency(charts...)
}

// Name returns the name of the chart.
func (ch *Chart) Name() string {
	if ch.Metadata == nil {
		return ""
	}
	return ch.Metadata.Name
}

// AddDependency determines if the chart is a subchart.
func (ch *Chart) AddDependency(charts ...*Chart) {
	for i, x := range charts {
		charts[i].parent = ch
		ch.dependencies = append(ch.dependencies, x)
	}
}

// Root finds the root chart.
func (ch *Chart) Root() *Chart {
	if ch.IsRoot() {
		return ch
	}
	return ch.Parent().Root()
}

// Dependencies are the charts that this chart depends on.
func (ch *Chart) Dependencies() []*Chart { return ch.dependencies }

// IsRoot determines if the chart is the root chart.
func (ch *Chart) IsRoot() bool { return ch.parent == nil }

// Parent returns a subchart's parent chart.
func (ch *Chart) Parent() *Chart { return ch.parent }

// SetParent sets a subchart's parent chart.
func (ch *Chart) SetParent(chart *Chart) { ch.parent = chart }

// ChartPath returns the full path to this chart in dot notation.
func (ch *Chart) ChartPath() string {
	if !ch.IsRoot() {
		return ch.Parent().ChartPath() + "." + ch.Name()
	}
	return ch.Name()
}

// ChartFullPath returns the full path to this chart.
func (ch *Chart) ChartFullPath() string {
	if !ch.IsRoot() {
		return ch.Parent().ChartFullPath() + "/charts/" + ch.Name()
	}
	return ch.Name()
}
