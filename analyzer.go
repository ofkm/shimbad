package shimbad

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

type functionInfo struct {
	name         string
	pos          token.Pos
	body         *ast.BlockStmt
	functionType *ast.FuncType
	object       types.Object
}

func forEachFunction(pass *analysis.Pass, config *configuration, visit func(functionInfo)) {
	for _, file := range pass.Files {
		filename := normalizedFilename(pass, file.Pos())
		if (!config.includeTests && strings.HasSuffix(filename, "_test.go")) ||
			(!config.includeGenerated && ast.IsGenerated(file)) ||
			matchesAny(config.excludedFiles, filename) {
			continue
		}

		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil || (function.Recv != nil && !config.includeMethods) {
				continue
			}
			info := declaredFunctionInfo(pass, function)
			if !matchesAny(config.excludedFunctions, info.name) {
				visit(info)
			}
		}

		if !config.includeFunctionLiterals {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			literal, ok := node.(*ast.FuncLit)
			if !ok {
				return true
			}
			position := pass.Fset.PositionFor(literal.Type.Func, false)
			name := fmt.Sprintf("%s.<func@%s:%d:%d>", pass.Pkg.Path(), filepath.Base(position.Filename), position.Line, position.Column)
			if !matchesAny(config.excludedFunctions, name) {
				visit(functionInfo{
					name:         name,
					pos:          literal.Type.Func,
					body:         literal.Body,
					functionType: literal.Type,
				})
			}
			return true
		})
	}
}

func declaredFunctionInfo(pass *analysis.Pass, function *ast.FuncDecl) functionInfo {
	object := pass.TypesInfo.Defs[function.Name]
	name := pass.Pkg.Path() + "." + function.Name.Name
	if typedFunction, ok := object.(*types.Func); ok {
		if signature, ok := typedFunction.Type().(*types.Signature); ok && signature.Recv() != nil {
			name = pass.Pkg.Path() + "." + receiverTypeName(signature.Recv().Type()) + "." + function.Name.Name
		}
	}
	return functionInfo{
		name:         name,
		pos:          function.Name.Pos(),
		body:         function.Body,
		functionType: function.Type,
		object:       object,
	}
}

func receiverTypeName(receiver types.Type) string {
	if pointer, ok := receiver.(*types.Pointer); ok {
		receiver = pointer.Elem()
	}
	receiver = types.Unalias(receiver)
	if named, ok := receiver.(*types.Named); ok {
		return named.Obj().Name()
	}
	return types.TypeString(receiver, func(*types.Package) string { return "" })
}

func normalizedFilename(pass *analysis.Pass, pos token.Pos) string {
	file := pass.Fset.File(pos)
	if file == nil {
		return ""
	}
	return filepath.ToSlash(file.Name())
}

func calledObject(pass *analysis.Pass, call *ast.CallExpr) types.Object {
	return calledExpressionObject(pass, call.Fun)
}

func calledExpressionObject(pass *analysis.Pass, expression ast.Expr) types.Object {
	switch value := expression.(type) {
	case *ast.Ident:
		return pass.TypesInfo.Uses[value]
	case *ast.SelectorExpr:
		return pass.TypesInfo.Uses[value.Sel]
	case *ast.IndexExpr:
		return calledExpressionObject(pass, value.X)
	case *ast.IndexListExpr:
		return calledExpressionObject(pass, value.X)
	case *ast.ParenExpr:
		return calledExpressionObject(pass, value.X)
	default:
		return nil
	}
}

func objectMatches(object types.Object, packagePath string, names ...string) bool {
	return object != nil &&
		object.Pkg() != nil &&
		object.Pkg().Path() == packagePath &&
		slices.Contains(names, object.Name())
}

func functionParameters(pass *analysis.Pass, function functionInfo) (map[types.Object]struct{}, bool) {
	parameters := make(map[types.Object]struct{})
	if function.functionType.Params == nil {
		return parameters, true
	}
	for _, field := range function.functionType.Params.List {
		if len(field.Names) == 0 {
			return nil, false
		}
		for _, name := range field.Names {
			if name.Name == "_" {
				return nil, false
			}
			parameter := pass.TypesInfo.Defs[name]
			if parameter == nil {
				return nil, false
			}
			parameters[parameter] = struct{}{}
		}
	}
	return parameters, true
}

func functionParameterCount(function functionInfo) int {
	if function.functionType.Params == nil {
		return 0
	}
	count := 0
	for _, field := range function.functionType.Params.List {
		if len(field.Names) == 0 {
			count++
		} else {
			count += len(field.Names)
		}
	}
	return count
}

func functionResultCount(function functionInfo) int {
	if function.functionType.Results == nil {
		return 0
	}
	count := 0
	for _, field := range function.functionType.Results.List {
		if len(field.Names) == 0 {
			count++
		} else {
			count += len(field.Names)
		}
	}
	return count
}

func statementsWithoutParameterDiscards(pass *analysis.Pass, function functionInfo) []ast.Stmt {
	parameters, _ := functionParameters(pass, function)
	statements := make([]ast.Stmt, 0, len(function.body.List))
	for _, statement := range function.body.List {
		if _, empty := statement.(*ast.EmptyStmt); empty || isParameterDiscard(pass, statement, parameters) {
			continue
		}
		statements = append(statements, statement)
	}
	return statements
}

func isParameterDiscard(pass *analysis.Pass, statement ast.Stmt, parameters map[types.Object]struct{}) bool {
	assignment, ok := statement.(*ast.AssignStmt)
	if !ok || assignment.Tok != token.ASSIGN || len(assignment.Lhs) == 0 || len(assignment.Rhs) == 0 {
		return false
	}
	for _, expression := range assignment.Lhs {
		identifier, ok := expression.(*ast.Ident)
		if !ok || identifier.Name != "_" {
			return false
		}
	}
	for _, expression := range assignment.Rhs {
		identifier, ok := expression.(*ast.Ident)
		if !ok {
			return false
		}
		object := pass.TypesInfo.Uses[identifier]
		if _, exists := parameters[object]; !exists {
			return false
		}
	}
	return true
}

func isStaticValue(pass *analysis.Pass, expression ast.Expr) bool {
	if parenthesized, ok := expression.(*ast.ParenExpr); ok {
		return isStaticValue(pass, parenthesized.X)
	}
	if value, exists := pass.TypesInfo.Types[expression]; exists && value.Value != nil {
		return true
	}
	if identifier, ok := expression.(*ast.Ident); ok {
		return identifier.Name == "nil"
	}
	composite, ok := expression.(*ast.CompositeLit)
	return ok && len(composite.Elts) == 0
}

func staticString(pass *analysis.Pass, expression ast.Expr) (string, bool) {
	value, exists := pass.TypesInfo.Types[expression]
	if !exists || value.Value == nil || value.Value.Kind() != constant.String {
		return "", false
	}
	return constant.StringVal(value.Value), true
}

func containsPlaceholder(pass *analysis.Pass, config *configuration, expression ast.Expr) bool {
	found := false
	ast.Inspect(expression, func(node ast.Node) bool {
		if found {
			return false
		}
		value, ok := node.(ast.Expr)
		if !ok {
			return true
		}
		text, ok := staticString(pass, value)
		if ok && matchesAny(config.placeholderPatterns, text) {
			found = true
			return false
		}
		return true
	})
	return found
}

func report(pass *analysis.Pass, function functionInfo, rule ruleID, message string) {
	pass.Report(analysis.Diagnostic{
		Pos:      function.pos,
		Category: string(rule),
		Message:  message,
	})
}
