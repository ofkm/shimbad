package shimbad

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/golangci/plugin-module-register/register"
)

type ruleID string

const (
	ruleTrivialForwarder ruleID = "trivial-forwarder"
	ruleEmptyStub        ruleID = "empty-stub"
	ruleConstantStub     ruleID = "constant-stub"
	rulePanicStub        ruleID = "panic-stub"
	rulePlaceholder      ruleID = "placeholder-result"
)

var allRuleIDs = []ruleID{
	ruleTrivialForwarder,
	ruleEmptyStub,
	ruleConstantStub,
	rulePanicStub,
	rulePlaceholder,
}

var defaultPlaceholderPatterns = []string{
	`(?i)\bTODO\b`,
	`(?i)\bNYI\b`,
	`(?i)\bnot(?: yet)? implemented\b`,
	`(?i)\bunimplemented\b`,
}

// Settings configures shimbad. Unknown fields are rejected.
type Settings struct {
	DisabledRules                 []string `json:"disabled-rules"`
	IncludeTests                  bool     `json:"include-tests"`
	IncludeGenerated              bool     `json:"include-generated"`
	IncludeMethods                bool     `json:"include-methods"`
	IncludeFunctionLiterals       bool     `json:"include-function-literals"`
	ExcludeFiles                  []string `json:"exclude-files"`
	ExcludeFunctions              []string `json:"exclude-functions"`
	PlaceholderPatterns           []string `json:"placeholder-patterns"`
	AdditionalPlaceholderPatterns []string `json:"additional-placeholder-patterns"`
}

type configuration struct {
	disabledRules           map[ruleID]struct{}
	includeTests            bool
	includeGenerated        bool
	includeMethods          bool
	includeFunctionLiterals bool
	excludedFiles           []*regexp.Regexp
	excludedFunctions       []*regexp.Regexp
	placeholderPatterns     []*regexp.Regexp
}

func newConfiguration(rawSettings any) (*configuration, error) {
	settings, err := register.DecodeSettings[Settings](rawSettings)
	if err != nil {
		return nil, err
	}

	disabled, err := validateDisabledRules(settings.DisabledRules)
	if err != nil {
		return nil, err
	}
	excludedFiles, err := compilePatterns("exclude-files", settings.ExcludeFiles)
	if err != nil {
		return nil, err
	}
	excludedFunctions, err := compilePatterns("exclude-functions", settings.ExcludeFunctions)
	if err != nil {
		return nil, err
	}

	placeholderPatterns := settings.PlaceholderPatterns
	if len(placeholderPatterns) == 0 {
		placeholderPatterns = defaultPlaceholderPatterns
	}
	placeholderPatterns = append(slices.Clone(placeholderPatterns), settings.AdditionalPlaceholderPatterns...)
	compiledPlaceholders, err := compilePatterns("placeholder patterns", placeholderPatterns)
	if err != nil {
		return nil, err
	}

	return &configuration{
		disabledRules:           disabled,
		includeTests:            settings.IncludeTests,
		includeGenerated:        settings.IncludeGenerated,
		includeMethods:          settings.IncludeMethods,
		includeFunctionLiterals: settings.IncludeFunctionLiterals,
		excludedFiles:           excludedFiles,
		excludedFunctions:       excludedFunctions,
		placeholderPatterns:     compiledPlaceholders,
	}, nil
}

func validateDisabledRules(values []string) (map[ruleID]struct{}, error) {
	disabled := make(map[ruleID]struct{}, len(values))
	for _, value := range values {
		rule := ruleID(value)
		if !slices.Contains(allRuleIDs, rule) {
			return nil, fmt.Errorf("unknown disabled rule %q (valid rules: %s)", value, ruleList())
		}
		if _, exists := disabled[rule]; exists {
			return nil, fmt.Errorf("disabled rule %q is listed more than once", value)
		}
		disabled[rule] = struct{}{}
	}
	return disabled, nil
}

func compilePatterns(setting string, values []string) ([]*regexp.Regexp, error) {
	patterns := make([]*regexp.Regexp, 0, len(values))
	for _, value := range values {
		if value == "" {
			return nil, fmt.Errorf("%s contains an empty regular expression", setting)
		}
		pattern, err := regexp.Compile(value)
		if err != nil {
			return nil, fmt.Errorf("compile %s pattern %q: %w", setting, value, err)
		}
		patterns = append(patterns, pattern)
	}
	return patterns, nil
}

func ruleList() string {
	values := make([]string, 0, len(allRuleIDs))
	for _, rule := range allRuleIDs {
		values = append(values, string(rule))
	}
	return strings.Join(values, ", ")
}

func (config *configuration) ruleEnabled(rule ruleID) bool {
	_, disabled := config.disabledRules[rule]
	return !disabled
}

func matchesAny(patterns []*regexp.Regexp, value string) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}
