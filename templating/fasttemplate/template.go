// Package fasttemplate implements simple and fast template library.
//
// Fasttemplate is faster than text/template, strings.Replace
// and strings.Replacer.
//
// Fasttemplate ideally fits for fast and simple placeholders' substitutions.
package fasttemplate

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

// executeFunc calls f on each template tag (placeholder) occurrence.
//
// Returns the number of bytes written to w.
//
// This function is optimized for constantly changing templates.
// Use Template.ExecuteFunc for frozen templates.
func executeFunc(template, startTag, endTag string, w io.Writer, f TagFunc) (int64, error) {
	var nn int64
	var ni int
	var err error
	for {
		n := strings.Index(template, startTag)
		if n < 0 {
			break
		}
		ni, err = w.Write([]byte(template[:n]))
		nn += int64(ni)
		if err != nil {
			return nn, err
		}

		template = template[n+len(startTag):]
		n = strings.Index(template, endTag)
		if n < 0 {
			// cannot find end tag - just write it to the output.
			ni, err = w.Write([]byte(startTag))
			nn += int64(ni)
			if err != nil {
				return nn, err
			}
			break
		}
		tag := template[:n]
		ni, err = f(w, tag)
		nn += int64(ni)
		if err != nil {
			if err == missingTag {
				ni, err = w.Write([]byte(startTag + tag + endTag))
				nn += int64(ni)
				if err != nil {
					return nn, err
				}
			} else {
				return nn, err
			}
		}
		template = template[n+len(endTag):]
	}
	ni, err = w.Write([]byte(template))
	nn += int64(ni)

	return nn, err
}

// Execute substitutes template tags (placeholders) with the corresponding
// values from the map m and writes the result to the given writer w.
//
// Substitution map m may contain values with the following types:
//   * []byte - the fastest value type
//   * string - convenient value type
//   * TagFunc - flexible value type
//
// Returns the number of bytes written to w.
//
// This function is optimized for constantly changing templates.
// Use Template.Execute for frozen templates.
func Execute(template, startTag, endTag string, w io.Writer, m map[string]interface{}) (int64, error) {
	return executeFunc(template, startTag, endTag, w, func(w io.Writer, tag string) (int, error) { return stdTagFunc(w, tag, m) })
}

// executeFuncString calls f on each template tag (placeholder) occurrence
// and substitutes it with the data written to TagFunc's w.
//
// Returns the resulting string.
//
// This function is optimized for constantly changing templates.
// Use Template.ExecuteFuncString for frozen templates.
func executeFuncString(template, startTag, endTag string, f TagFunc) string {
	tagsCount := bytes.Count([]byte(template), []byte(startTag))
	if tagsCount == 0 {
		return template
	}

	bb := &bytes.Buffer{}
	if _, err := executeFunc(template, startTag, endTag, bb, f); err != nil {
		panic(fmt.Sprintf("unexpected error: %s", err))
	}
	return bb.String()
}

// ExecuteString substitutes template tags (placeholders) with the corresponding
// values from the map m and returns the result.
//
// Substitution map m may contain values with the following types:
//   * []byte - the fastest value type
//   * string - convenient value type
//   * TagFunc - flexible value type
//
// This function is optimized for constantly changing templates.
// Use Template.ExecuteString for frozen templates.
func ExecuteString(template, startTag, endTag string, m map[string]interface{}) string {
	return executeFuncString(template, startTag, endTag, func(w io.Writer, tag string) (int, error) { return stdTagFunc(w, tag, m) })
}

// TagFunc can be used as a substitution value in the map passed to Execute*.
// Execute* functions pass tag (placeholder) name in 'tag' argument.
//
// TagFunc must be safe to call from concurrently running goroutines.
//
// TagFunc must write contents to w and return the number of bytes written.
type TagFunc func(w io.Writer, tag string) (int, error)

var missingTag = errors.New("missing tag")

func stdTagFunc(w io.Writer, tag string, m map[string]interface{}) (int, error) {
	tag = strings.TrimSpace(tag)
	v, exists := m[tag]
	if !exists {
		return 0, missingTag
	}
	if v == nil {
		return 0, nil
	}
	switch value := v.(type) {
	case []byte:
		return w.Write(value)
	case string:
		return w.Write([]byte(value))
	case TagFunc:
		return value(w, tag)
	default:
		panic(fmt.Sprintf("tag=%q contains unexpected value type=%#v. Expected []byte, string or TagFunc", tag, v))
	}
}
