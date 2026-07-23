package shimbad

import (
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/golangci/plugin-module-register/register"
)

func TestPluginRegistration(t *testing.T) {
	newPlugin, err := register.GetPlugin("shimbad")
	if err != nil {
		t.Fatalf("get plugin: %v", err)
	}

	registered, err := newPlugin(nil)
	if err != nil {
		t.Fatalf("construct plugin: %v", err)
	}
	if got := registered.GetLoadMode(); got != register.LoadModeTypesInfo {
		t.Fatalf("load mode = %q, want %q", got, register.LoadModeTypesInfo)
	}

	analyzers, err := registered.BuildAnalyzers()
	if err != nil {
		t.Fatalf("build analyzers: %v", err)
	}
	names := make([]string, 0, len(analyzers))
	for _, analyzer := range analyzers {
		names = append(names, analyzer.Name)
	}
	for _, expected := range []string{"shimbadshims", "shimbadstubs"} {
		if !slices.Contains(names, expected) {
			t.Errorf("analyzers %v do not include %q", names, expected)
		}
	}
}

func TestDefaultConfiguration(t *testing.T) {
	config := newTestConfiguration(t, nil)

	for _, rule := range allRuleIDs {
		if !config.ruleEnabled(rule) {
			t.Errorf("rule %q is disabled by default", rule)
		}
	}
	if config.includeTests || config.includeGenerated || config.includeMethods || config.includeFunctionLiterals {
		t.Fatal("tests, generated files, methods, or function literals are included by default")
	}
	if len(config.placeholderPatterns) != len(defaultPlaceholderPatterns) {
		t.Fatalf("placeholder pattern count = %d, want %d", len(config.placeholderPatterns), len(defaultPlaceholderPatterns))
	}
}

func TestCustomConfiguration(t *testing.T) {
	config := newTestConfiguration(t, map[string]any{
		"disabled-rules":                  []string{"constant-stub"},
		"include-tests":                   true,
		"include-generated":               true,
		"include-methods":                 true,
		"include-function-literals":       true,
		"exclude-files":                   []string{`vendor/`},
		"exclude-functions":               []string{`\.Generated$`},
		"placeholder-patterns":            []string{`(?i)pending`},
		"additional-placeholder-patterns": []string{`(?i)later`},
	})

	if config.ruleEnabled(ruleConstantStub) {
		t.Fatal("constant-stub was not disabled")
	}
	if !config.includeTests || !config.includeGenerated || !config.includeMethods || !config.includeFunctionLiterals {
		t.Fatal("declaration inclusion settings were not enabled")
	}
	if len(config.excludedFiles) != 1 || len(config.excludedFunctions) != 1 {
		t.Fatal("exclusion patterns were not compiled")
	}
	if len(config.placeholderPatterns) != 2 {
		t.Fatalf("placeholder pattern count = %d, want 2", len(config.placeholderPatterns))
	}
}

func TestInvalidConfiguration(t *testing.T) {
	tests := map[string]struct {
		settings map[string]any
		want     string
	}{
		"unknown field": {
			settings: map[string]any{"unknown": true},
			want:     "unknown field",
		},
		"unknown rule": {
			settings: map[string]any{"disabled-rules": []string{"mystery"}},
			want:     "unknown disabled rule",
		},
		"duplicate rule": {
			settings: map[string]any{"disabled-rules": []string{"empty-stub", "empty-stub"}},
			want:     "listed more than once",
		},
		"invalid file expression": {
			settings: map[string]any{"exclude-files": []string{"["}},
			want:     "compile exclude-files",
		},
		"empty function expression": {
			settings: map[string]any{"exclude-functions": []string{""}},
			want:     "empty regular expression",
		},
		"invalid placeholder expression": {
			settings: map[string]any{"placeholder-patterns": []string{"("}},
			want:     "compile placeholder patterns",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := New(test.settings)
			if err == nil {
				t.Fatal("New() accepted invalid settings")
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("New() error = %q, want it to contain %q", err, test.want)
			}
		})
	}
}

func newTestConfiguration(t *testing.T, raw any) *configuration {
	t.Helper()
	config, err := newConfiguration(raw)
	if err != nil {
		t.Fatalf("new configuration: %v", err)
	}
	return config
}

func testdataDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve testdata directory")
	}
	return filepath.Join(filepath.Dir(filename), "testdata")
}
