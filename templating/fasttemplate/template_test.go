package fasttemplate

import (
	"bytes"
	"io"
	"testing"
)

func TestExecuteFunc(t *testing.T) {
	testExecuteFunc(t, "", "")
	testExecuteFunc(t, "a", "a")
	testExecuteFunc(t, "abc", "abc")
	testExecuteFunc(t, "{foo}", "xxxx")
	testExecuteFunc(t, "a{foo}", "axxxx")
	testExecuteFunc(t, "{foo}a", "xxxxa")
	testExecuteFunc(t, "a{foo}bc", "axxxxbc")
	testExecuteFunc(t, "{foo}{foo}", "xxxxxxxx")
	testExecuteFunc(t, "{foo}{foo}", "xxxxxxxx")
	testExecuteFunc(t, "{foo}bar{foo}", "xxxxbarxxxx")

	// unclosed tag
	testExecuteFunc(t, "{unclosed", "{unclosed")
	testExecuteFunc(t, "{{unclosed", "{{unclosed")
	testExecuteFunc(t, "{un{closed", "{un{closed")

	// test unknown tag
	testExecuteFunc(t, "{unknown}", "zz")
	testExecuteFunc(t, "{foo}q{unexpected}{missing}bar{foo}", "xxxxqzzzzbarxxxx")
}

func testExecuteFunc(t *testing.T, template, expectedOutput string) {
	var bb bytes.Buffer
	executeFunc(template, "{", "}", &bb, func(w io.Writer, tag string) (int, error) {
		if tag == "foo" {
			return w.Write([]byte("xxxx"))
		}
		return w.Write([]byte("zz"))
	})

	output := string(bb.Bytes())
	if output != expectedOutput {
		t.Fatalf("unexpected output for template=%q: %q. Expected %q", template, output, expectedOutput)
	}
}

func TestExecute(t *testing.T) {
	testExecute(t, "", "")
	testExecute(t, "a", "a")
	testExecute(t, "abc", "abc")
	testExecute(t, "{foo}", "xxxx")
	testExecute(t, "a{foo}", "axxxx")
	testExecute(t, "{foo}a", "xxxxa")
	testExecute(t, "a{foo}bc", "axxxxbc")
	testExecute(t, "{foo}{foo}", "xxxxxxxx")
	testExecute(t, "{foo}bar{foo}", "xxxxbarxxxx")

	// unclosed tag
	testExecute(t, "{unclosed", "{unclosed")
	testExecute(t, "{{unclosed", "{{unclosed")
	testExecute(t, "{un{closed", "{un{closed")

	// test unknown tag
	testExecute(t, "{unknown}", "{unknown}")
	testExecute(t, "{foo}q{unexpected}{missing}bar{foo}", "xxxxq{unexpected}{missing}barxxxx")
	testExecute(t, "{foo}q{ unexpected }{ missing }bar{foo}", "xxxxq{ unexpected }{ missing }barxxxx")
}

func testExecute(t *testing.T, template, expectedOutput string) {
	var bb bytes.Buffer
	Execute(template, "{", "}", &bb, map[string]interface{}{"foo": "xxxx"})
	output := string(bb.Bytes())
	if output != expectedOutput {
		t.Fatalf("unexpected output for template=%q: %q. Expected %q", template, output, expectedOutput)
	}
}

func TestExecuteString(t *testing.T) {
	testExecuteString(t, "", "")
	testExecuteString(t, "a", "a")
	testExecuteString(t, "abc", "abc")
	testExecuteString(t, "{foo}", "xxxx")
	testExecuteString(t, "a{foo}", "axxxx")
	testExecuteString(t, "{foo}a", "xxxxa")
	testExecuteString(t, "a{foo}bc", "axxxxbc")
	testExecuteString(t, "{foo}{foo}", "xxxxxxxx")
	testExecuteString(t, "{foo}bar{foo}", "xxxxbarxxxx")

	// unclosed tag
	testExecuteString(t, "{unclosed", "{unclosed")
	testExecuteString(t, "{{unclosed", "{{unclosed")
	testExecuteString(t, "{un{closed", "{un{closed")

	// test unknown tag
	testExecuteString(t, "{unknown}", "{unknown}")
	testExecuteString(t, "{foo}q{unexpected}{missing}bar{foo}", "xxxxq{unexpected}{missing}barxxxx")
	testExecuteString(t, "{foo}q{ unexpected }{ missing }bar{foo}", "xxxxq{ unexpected }{ missing }barxxxx")
}

func testExecuteString(t *testing.T, template, expectedOutput string) {
	output := ExecuteString(template, "{", "}", map[string]interface{}{"foo": "xxxx"})
	if output != expectedOutput {
		t.Fatalf("unexpected output for template=%q: %q. Expected %q", template, output, expectedOutput)
	}
}

func expectPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("missing panic")
		}
	}()
	f()
}
