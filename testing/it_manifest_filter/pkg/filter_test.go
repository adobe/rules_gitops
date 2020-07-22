package filter_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	filter "github.com/adobe/rules_gitops/testing/it_manifest_filter/pkg"
	"github.com/google/go-cmp/cmp"
)

func TestHappyPath(t *testing.T) {
	testcases := []string{"happypath", "statefulset", "statefulset2", "certificate"}
	for _, testcase := range testcases {
		t.Run(testcase, func(t *testing.T) {
			infn := fmt.Sprintf("testdata/%s.yaml", testcase)
			expectedfn := fmt.Sprintf("testdata/%s.expected.yaml", testcase)
			inf, err := os.Open(infn)
			if err != nil {
				t.Errorf("Unable to open file %s", infn)
				return
			}
			defer inf.Close()
			expectedb, err := ioutil.ReadFile(expectedfn)
			if err != nil {
				t.Errorf("Unable to read file %s", expectedfn)
				return
			}
			expected := strings.TrimSpace(string(expectedb))
			var outbuf bytes.Buffer
			err = filter.ReplacePDWithEmptyDirs(inf, &outbuf)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
				return
			}
			if diff := cmp.Diff(expected, strings.TrimSpace(outbuf.String())); diff != "" {
				t.Errorf("Unexpected output (-want +got):\n%s", diff)
			}
		})
	}
}
