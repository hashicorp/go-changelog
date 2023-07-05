package parser

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParser_Section(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		version string

		expectedSection *Section
		expectedErr     error
	}{
		{
			"empty log",
			"",
			"0.12.0",
			nil,
			&VersionNotFoundErr{"0.12.0"},
		},
		{
			"version not found",
			`## 0.11.0

something

## 0.10.0

testing
`,
			"0.12.0",
			nil,
			&VersionNotFoundErr{"0.12.0"},
		},
		{
			"matching unreleased version",
			`## 0.12.0 (Unreleased)

something

## 0.11.0

testing
`,
			"0.12.0",
			&Section{
				Header: []byte("## 0.12.0 (Unreleased)"),
				Body:   []byte("something"),
			},
			nil,
		},
		{
			"matching released version - top",
			`## 0.12.0
matching text
with newline

## 0.11.99

 - something
 - else
`,
			"0.12.0",
			&Section{
				Header: []byte("## 0.12.0"),
				Body: []byte(`matching text
with newline`),
			},
			nil,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			r := strings.NewReader(tc.content)
			p, err := NewSectionParser(r)
			if err != nil {
				t.Fatal(err)
			}
			s, err := p.Section(tc.version)
			if err == nil && tc.expectedErr != nil {
				t.Fatalf("expected error: %s", tc.expectedErr.Error())
			}

			if !errors.Is(err, tc.expectedErr) {
				diff := cmp.Diff(tc.expectedErr, err)
				t.Fatalf("error doesn't match: %s", diff)
			}

			if diff := cmp.Diff(tc.expectedSection, s); diff != "" {
				t.Fatalf("parsed section don't match: %s", diff)
			}
		})
	}
}
