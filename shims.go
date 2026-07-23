package shimbad

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

func newShimAnalyzer(config *configuration) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "shimbadshims",
		Doc:  "detect trivial forwarding and call-composition functions",
		Run: func(pass *analysis.Pass) (any, error) {
			if !config.ruleEnabled(ruleTrivialForwarder) {
				return nil, nil
			}
			forEachFunction(pass, config, func(function functionInfo) {
				if isTrivialForwardingFunction(pass, function) {
					report(pass, function, ruleTrivialForwarder, "avoid a trivial forwarding function; call the underlying function directly")
				}
			})
			return nil, nil
		},
	}
}

func isTrivialForwardingFunction(pass *analysis.Pass, function functionInfo) bool {
	if len(function.body.List) != 1 {
		return false
	}
	call, ok := forwardingCall(function.body.List[0])
	if !ok {
		return false
	}

	parameters, valid := functionParameters(pass, function)
	if !valid || len(parameters) == 0 {
		return false
	}
	seen := make(map[types.Object]int, len(parameters))
	if !isComposableCall(pass, call, function.object, parameters, seen) {
		return false
	}
	for parameter := range parameters {
		if seen[parameter] != 1 {
			return false
		}
	}
	return true
}

func isComposableCall(
	pass *analysis.Pass,
	call *ast.CallExpr,
	declared types.Object,
	parameters map[types.Object]struct{},
	seen map[types.Object]int,
) bool {
	switch callee := calledObject(pass, call).(type) {
	case *types.Builtin:
		if callee.Name() == "panic" || callee.Name() == "recover" {
			return false
		}
	case *types.Func:
		if callee == declared {
			return false
		}
		signature, ok := callee.Type().(*types.Signature)
		if !ok || signature.Recv() != nil {
			return false
		}
	case *types.TypeName:
		// Type conversions can be trivial forwarding implementations.
	default:
		return false
	}

	for _, argument := range call.Args {
		if !isComposableArgument(pass, argument, declared, parameters, seen) {
			return false
		}
	}
	return true
}

func isComposableArgument(
	pass *analysis.Pass,
	expression ast.Expr,
	declared types.Object,
	parameters map[types.Object]struct{},
	seen map[types.Object]int,
) bool {
	if parenthesized, ok := expression.(*ast.ParenExpr); ok {
		return isComposableArgument(pass, parenthesized.X, declared, parameters, seen)
	}
	if identifier, ok := expression.(*ast.Ident); ok {
		if parameter := pass.TypesInfo.Uses[identifier]; parameter != nil {
			if _, exists := parameters[parameter]; exists {
				seen[parameter]++
				return true
			}
		}
	}
	if call, ok := expression.(*ast.CallExpr); ok {
		return isComposableCall(pass, call, declared, parameters, seen)
	}
	return isStaticValue(pass, expression)
}

func forwardingCall(statement ast.Stmt) (*ast.CallExpr, bool) {
	switch value := statement.(type) {
	case *ast.ReturnStmt:
		if len(value.Results) != 1 {
			return nil, false
		}
		call, ok := value.Results[0].(*ast.CallExpr)
		return call, ok
	case *ast.ExprStmt:
		call, ok := value.X.(*ast.CallExpr)
		return call, ok
	default:
		return nil, false
	}
}
