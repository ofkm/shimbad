// Package shimbad provides static analysis for suspicious Go function implementations.
package shimbad

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() { //nolint:gochecknoinits // Module plugins register through init.
	register.Plugin("shimbad", New)
}

// Plugin provides the shimbad analyzers to golangci-lint.
type Plugin struct {
	config *configuration
}

// New constructs the shimbad golangci-lint module plugin.
func New(rawSettings any) (register.LinterPlugin, error) {
	config, err := newConfiguration(rawSettings)
	if err != nil {
		return nil, err
	}

	return &Plugin{config: config}, nil
}

// BuildAnalyzers returns the shimbad analyzers.
func (plugin *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		newShimAnalyzer(plugin.config),
		newStubAnalyzer(plugin.config),
	}, nil
}

// GetLoadMode requests type information for reliable symbol matching.
func (*Plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
