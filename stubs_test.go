package shimbad

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestStubAnalyzer(t *testing.T) {
	analysistest.Run(t, testdataDir(t), newStubAnalyzer(newTestConfiguration(t, nil)),
		"example.com/shimbad/stubs",
	)
}

func TestStubAnalyzerDeclarationOptions(t *testing.T) {
	config := newTestConfiguration(t, map[string]any{
		"include-methods":           true,
		"include-function-literals": true,
	})
	analysistest.Run(t, testdataDir(t), newStubAnalyzer(config),
		"example.com/shimbad/stubsdeclarations",
	)
}

func TestStubAnalyzerPlaceholderOverride(t *testing.T) {
	config := newTestConfiguration(t, map[string]any{
		"placeholder-patterns": []string{`(?i)pending implementation`},
	})
	analysistest.Run(t, testdataDir(t), newStubAnalyzer(config),
		"example.com/shimbad/stubscustomplaceholder",
	)
}

func TestStubAnalyzerDisabledRules(t *testing.T) {
	config := newTestConfiguration(t, map[string]any{
		"disabled-rules": []string{
			"empty-stub",
			"constant-stub",
			"panic-stub",
			"placeholder-result",
		},
	})
	analysistest.Run(t, testdataDir(t), newStubAnalyzer(config),
		"example.com/shimbad/stubsdisabled",
	)
}
