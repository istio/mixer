// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package interfacegen

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

type logFn func(string, ...interface{})

// TestBazelGeneration uses the outputs files generated via the bazel rule
// for testdata:generated_files and compares them against the golden files.
func TestBazelGeneration(t *testing.T) {
	tests := []struct {
		name, got, want string
	}{
		{"Metrics", "testdata/metric_template_library_processor.gen.go", "testdata/MetricTemplateProcessorInterface.golden.go"},
		{"Quota", "testdata/quota_template_library_processor.gen.go", "testdata/QuotaTemplateProcessorInterface.golden.go"},
		{"Logs", "testdata/log_template_library_processor.gen.go", "testdata/LogTemplateProcessorInterface.golden.go"},
		{"Lists", "testdata/list_template_library_processor.gen.go", "testdata/ListTemplateProcessorInterface.golden.go"},
		{"Nested Message", "testdata/nested_message_library_processor.gen.go", "testdata/NestedMessageProcessorInterface.golden.go"},
	}
	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			if same := deepCompare(v.got, v.want, t.Errorf); !same {
				t.Error("Files were not the same.")
			}
		})
	}
}

func TestGenerate_Errors(t *testing.T) {
	g := Generator{OutFilePath: "."}
	err := g.Generate("testdata/error_template.descriptor_set")
	if err == nil {
		t.Fatalf("Generate(%s) should have produced an error", "testdata/error_template.descriptor_set")
	}
	b, fileErr := ioutil.ReadFile("testdata/ErrorTemplate.baseline")
	if fileErr != nil {
		t.Fatalf("Could not read baseline file: %v", err)
	}
	want := fmt.Sprintf("%s", b)
	got := err.Error()
	if got != want {
		t.Fatalf("Generate(%s) => '%s'\nwanted: '%s'", "testdata/error_template.descriptor_set", got, want)
	}
}

const chunkSize = 64000

func deepCompare(file1, file2 string, logf logFn) bool {
	f1, err := os.Open(file1)
	if err != nil {
		logf("could not open file: %v", err)
		return false
	}

	f2, err := os.Open(file2)
	if err != nil {
		logf("could not open file: %v", err)
		return false
	}

	for {
		b1 := make([]byte, chunkSize)
		s1, err1 := f1.Read(b1)

		b2 := make([]byte, chunkSize)
		s2, err2 := f2.Read(b2)

		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				return true
			} else if err1 == io.EOF || err2 == io.EOF {
				return false
			} else {
				return false
			}
		}

		if !bytes.Equal(b1, b2) {
			logf("bytes don't match (sizes: %d, %d):\n%s\n%s", s1, s2, string(b1), string(b2))
			return false
		}
	}
}
