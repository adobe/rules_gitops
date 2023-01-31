package k8ssort

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSort(t *testing.T) {
	cases := []struct {
		Name  string
		YAML  string
		Want  string
		Error string
	}{{
		Name: "one item",
		YAML: `apiVersion: v1
kind: Namespace
metadata:
  name: mynamespace`,
		Want: `apiVersion: v1
kind: Namespace
metadata:
  name: mynamespace
`,
		Error: "",
	}, {
		Name: "two items, sorted",
		YAML: `apiVersion: v1
kind: Namespace
metadata:
  name: mynamespace
---
apiVersion: v1
kind: Pod
metadata:
  name: mypod
`,
		Want: `apiVersion: v1
kind: Namespace
metadata:
  name: mynamespace
---
apiVersion: v1
kind: Pod
metadata:
  name: mypod
`,
		Error: "",
	}, {
		Name: "two items, unsorted",
		YAML: `apiVersion: v1
kind: Pod
metadata:
  name: mypod
---
apiVersion: v1
kind: Namespace
metadata:
  name: mynamespace`,
		Want: `apiVersion: v1
kind: Namespace
metadata:
  name: mynamespace
---
apiVersion: v1
kind: Pod
metadata:
  name: mypod
`,
		Error: "",
	}}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			r := strings.NewReader(c.YAML)
			got, err := Sort(r, "", false)
			if err != nil && c.Error == "" {
				t.Fatalf("unexpected error in yaml %s: %v", c.YAML, err)
			} else if err == nil && c.Error != "" {
				t.Fatalf("unexpected success in yaml %s: want error substr: %s", c.YAML, c.Error)
			} else if err != nil && !strings.Contains(err.Error(), c.Error) {
				t.Fatalf("error mismatch in yaml %s: got: %v want substr: %s", c.YAML, err, c.Error)
			}
			if diff := cmp.Diff(string(got), c.Want); diff != "" {
				t.Fatalf("yaml mismatch: (-got, +want): %s", diff)
			}
		})
	}
}
