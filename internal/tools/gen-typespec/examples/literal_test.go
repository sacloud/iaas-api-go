// Copyright 2022-2025 The sacloud/iaas-api-go Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package examples

import "testing"

func TestToTSPLiteral(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		indent int
		want   string
	}{
		{
			name:   "scalar int stays exact (no float truncation)",
			input:  `{"ID": 123456789012345}`,
			indent: 0,
			want:   `#{ID: 123456789012345}`,
		},
		{
			name:   "string / bool / null",
			input:  `{"Name": "example", "OK": true, "X": null}`,
			indent: 0,
			want:   `#{Name: "example", OK: true, X: null}`,
		},
		{
			name:   "nested object and array",
			input:  `{"a": [1, 2, {"b": "c"}], "z": true}`,
			indent: 0,
			want:   `#{a: #[1, 2, #{b: "c"}], z: true}`,
		},
		{
			name:   "empty containers",
			input:  `{"arr": [], "obj": {}}`,
			indent: 0,
			want:   `#{arr: #[], obj: #{}}`,
		},
		{
			name:   "string escaping",
			input:  `{"s": "a\"b\n"}`,
			indent: 0,
			want:   `#{s: "a\"b\n"}`,
		},
		{
			name:   "keys are sorted",
			input:  `{"b": 1, "a": 2}`,
			indent: 0,
			want:   `#{a: 2, b: 1}`,
		},
		{
			name:   "non-identifier keys get backticked",
			input:  `{"Provider.Class": "dns"}`,
			indent: 0,
			want:   "#{`Provider.Class`: \"dns\"}",
		},
		{
			name:   "indent mode",
			input:  `{"a": [1, {"b": 2}]}`,
			indent: 2,
			want: `#{
  a: #[
    1,
    #{
      b: 2
    }
  ]
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToTSPLiteral(tt.input, tt.indent)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("mismatch\n  got:  %s\n  want: %s", got, tt.want)
			}
		})
	}
}

func TestToTSPLiteralInvalid(t *testing.T) {
	if _, err := ToTSPLiteral(`{not json`, 0); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
