package shimbad

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

func newStubAnalyzer(config *configuration) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "shimbadstubs",
		Doc:  "detect empty, constant, panic, and placeholder-result function stubs",
		Run: func(pass *analysis.Pass) (any, error) {
			forEachFunction(pass, config, func(function functionInfo) {
				inspectStub(pass, config, function)
			})
			return nil, nil
		},
	}
}

func inspectStub(pass *analysis.Pass, config *configuration, function functionInfo) {
	statements := statementsWithoutParameterDiscards(pass, function)
	if len(statements) == 0 {
		if config.ruleEnabled(ruleEmptyStub) {
			report(pass, function, ruleEmptyStub, "function has no implementation")
		}
		return
	}
	if len(statements) != 1 {
		return
	}

	if config.ruleEnabled(rulePanicStub) && isPlaceholderPanic(pass, config, statements[0]) {
		report(pass, function, rulePanicStub, "function is implemented only by a placeholder panic")
		return
	}

	returnStatement, ok := statements[0].(*ast.ReturnStmt)
	if !ok {
		return
	}
	if config.ruleEnabled(rulePlaceholder) && isPlaceholderResult(pass, config, returnStatement) {
		report(pass, function, rulePlaceholder, "function returns only a placeholder not-implemented result")
		return
	}
	if config.ruleEnabled(ruleConstantStub) && isConstantStub(pass, function, returnStatement) {
		report(pass, function, ruleConstantStub, "function ignores its inputs and returns only static values")
	}
}

func isPlaceholderPanic(pass *analysis.Pass, config *configuration, statement ast.Stmt) bool {
	expression, ok := statement.(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := expression.X.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return false
	}
	builtin, ok := calledObject(pass, call).(*types.Builtin)
	return ok && builtin.Name() == "panic" && containsPlaceholder(pass, config, call.Args[0])
}

func isPlaceholderResult(pass *analysis.Pass, config *configuration, statement *ast.ReturnStmt) bool {
	if len(statement.Results) == 0 {
		return false
	}
	foundPlaceholder := false
	for _, result := range statement.Results {
		call, ok := result.(*ast.CallExpr)
		if ok && isPlaceholderErrorCall(pass, config, call) {
			foundPlaceholder = true
			continue
		}
		if !isStaticValue(pass, result) {
			return false
		}
	}
	return foundPlaceholder
}

func isPlaceholderErrorCall(pass *analysis.Pass, config *configuration, call *ast.CallExpr) bool {
	if len(call.Args) == 0 {
		return false
	}
	object := calledObject(pass, call)
	if !objectMatches(object, "errors", "New") && !objectMatches(object, "fmt", "Errorf") {
		return false
	}
	text, ok := staticString(pass, call.Args[0])
	return ok && matchesAny(config.placeholderPatterns, text)
}

func isConstantStub(pass *analysis.Pass, function functionInfo, statement *ast.ReturnStmt) bool {
	if functionParameterCount(function) == 0 {
		return false
	}
	if len(statement.Results) == 0 {
		return functionResultCount(function) > 0
	}
	for _, result := range statement.Results {
		if !isStaticValue(pass, result) {
			return false
		}
	}
	return true
}
