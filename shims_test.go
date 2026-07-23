package shimbad

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestShimAnalyzer(t *testing.T) {
	analysistest.Run(t, testdataDir(t), newShimAnalyzer(newTestConfiguration(t, nil)),
		"example.com/shimbad/shims",
		"example.com/shimbad/shimsgenerateddefault",
		"example.com/shimbad/shimstestsdefault",
	)
}

func TestShimAnalyzerDeclarationOptions(t *testing.T) {
	config := newTestConfiguration(t, map[string]any{
		"include-methods":           true,
		"include-function-literals": true,
	})
	analysistest.Run(t, testdataDir(t), newShimAnalyzer(config),
		"example.com/shimbad/shimsdeclarations",
	)
}

func TestShimAnalyzerFileOptions(t *testing.T) {
	config := newTestConfiguration(t, map[string]any{
		"include-tests":     true,
		"include-generated": true,
	})
	analysistest.Run(t, testdataDir(t), newShimAnalyzer(config),
		"example.com/shimbad/shimsgeneratedincluded",
		"example.com/shimbad/shimstestsincluded",
	)
}

func TestShimAnalyzerExclusions(t *testing.T) {
	config := newTestConfiguration(t, map[string]any{
		"exclude-files":     []string{`file_excluded\.go$`},
		"exclude-functions": []string{`\.Excluded$`},
	})
	analysistest.Run(t, testdataDir(t), newShimAnalyzer(config),
		"example.com/shimbad/shimsexcluded",
	)
}

func TestShimAnalyzerDisabled(t *testing.T) {
	config := newTestConfiguration(t, map[string]any{
		"disabled-rules": []string{"trivial-forwarder"},
	})
	analysistest.Run(t, testdataDir(t), newShimAnalyzer(config),
		"example.com/shimbad/shimsdisabled",
	)
}
