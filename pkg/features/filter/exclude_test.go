package filter

import (
	"github.com/timo-reymann/deterministic-zip/pkg/cli"
	"reflect"
	"testing"
)

func TestExclude_IsEnabled(t *testing.T) {
	config := cli.Configuration{Exclude: []string{
		"foo.*",
	}}
	exclude := Exclude{}
	if !exclude.IsEnabled(&config) {
		t.Fatalf("Should execute for non empty exclude")
	}
}

func TestExclude_Execute(t *testing.T) {
	testCases := []struct {
		sourceFiles []string
		targetFiles []string
		patterns    []string
	}{
		{
			sourceFiles: []string{
				"foo.bar",
			},
			targetFiles: []string{},
			patterns: []string{
				"*.bar",
			},
		},
		{
			sourceFiles: []string{
				".git/HEAD",
				".git/abc",
				".git/refs/bla",
			},
			targetFiles: []string{},
			patterns: []string{
				".git/*",
			},
		},
		{
			sourceFiles: []string{
				".git/HEAD",
				".git/abc",
				".git/refs/bla",
			},
			targetFiles: []string{},
			patterns: []string{
				".git/*",
			},
		},
		{
			sourceFiles: []string{
				"foo.bar",
			},
			targetFiles: []string{},
			patterns: []string{
				"*.zip",
				"*.bar",
			},
		},
	}

	for _, tc := range testCases {
		config := cli.Configuration{
			SourceFiles: tc.sourceFiles,
			Exclude:     tc.patterns,
		}
		exclude := Exclude{}
		if err := exclude.Execute(&config); err != nil {
			t.Fatal(err)
		}

		// DeepEquals doesnt like empty arrays
		if len(tc.targetFiles) == 0 && len(config.SourceFiles) == 0 {
			continue
		}

		if !reflect.DeepEqual(tc.targetFiles, config.SourceFiles) {
			t.Fatalf("Expected %v, but got %v for patterns %v", tc.targetFiles, config.SourceFiles, tc.patterns)
		}
	}

}
