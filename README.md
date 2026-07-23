# shimbad

`shimbad` is a type-aware [golangci-lint module plugin](https://golangci-lint.run/docs/plugins/module-plugins/)
that detects suspicious Go function implementations: thin forwarding shims,
empty bodies, constant-return stubs, placeholder panics, and
not-yet-implemented results.

It intentionally stays focused on function-body smells. Use established
golangci-lint analyzers such as `govet`, `staticcheck`, `errcheck`, `unparam`,
`revive`, and `godox` for broader correctness and style checks.

## Rules

All rules are enabled by default.

| Rule | Diagnostic |
| --- | --- |
| `trivial-forwarder` | A named function only forwards each parameter once through a call or nested call composition. Static default arguments and type conversions are supported. |
| `empty-stub` | A function body is empty or contains only assignments that discard its parameters. |
| `constant-stub` | A function ignores at least one input and immediately returns only constants, zero values, `nil`, empty composite literals, or a bare named-result return. |
| `panic-stub` | A function is implemented only by `panic` with a recognized placeholder message. |
| `placeholder-result` | A function immediately returns `errors.New` or `fmt.Errorf` with a recognized placeholder message, plus only static companion results. |

The default placeholder patterns recognize `TODO`, `NYI`, `not implemented`,
`not yet implemented`, and `unimplemented`, case-insensitively.

The analyzers deliberately accept functions that validate, branch, mutate
state, call methods, use parameters more than once, ignore only some forwarding
parameters, or contain other domain behavior. Bodyless declarations are also
accepted.

## Build a custom golangci-lint

Module plugins are linked into a custom golangci-lint binary. They are not
loaded dynamically by the standard binary.

Add `.custom-gcl.yml` to the consuming repository:

```yaml
version: v2.12.2
name: golangci-lint-custom
destination: ./.bin

plugins:
  - module: go.ofkm.dev/shimbad
    version: v0.1.0
```

Use an available `shimbad` release in place of `v0.1.0`. For local development,
replace `version` with a path:

```yaml
plugins:
  - module: go.ofkm.dev/shimbad
    path: ../shimbad
```

Build the binary:

```sh
mkdir -p .bin
golangci-lint custom
```

## Configure golangci-lint

Enable the custom linter in `.golangci.yml`:

```yaml
version: "2"

linters:
  enable:
    - shimbad
  settings:
    custom:
      shimbad:
        type: module
        description: Detects suspicious Go function implementations.
        original-url: https://go.ofkm.dev/shimbad
        settings:
          disabled-rules:
            - constant-stub
          include-tests: false
          include-generated: false
          include-methods: false
          include-function-literals: false
          exclude-files:
            - /mocks/
          exclude-functions:
            - '\.IntentionalNoop$'
          additional-placeholder-patterns:
            - '(?i)\bpending implementation\b'
```

Then run the custom binary:

```sh
./.bin/golangci-lint-custom run ./...
```

Unknown settings, unknown rule names, duplicate disabled rules, empty regular
expressions, and invalid regular expressions are configuration errors.

### Settings

| Setting | Default | Description |
| --- | --- | --- |
| `disabled-rules` | `[]` | Rule identifiers to disable. |
| `include-tests` | `false` | Analyze files ending in `_test.go`. |
| `include-generated` | `false` | Analyze files recognized by `go/ast.IsGenerated`. |
| `include-methods` | `false` | Analyze named methods in addition to free functions. |
| `include-function-literals` | `false` | Analyze function literals and closures. |
| `exclude-files` | `[]` | Go regular expressions matched against slash-normalized file names from the compiler file set. |
| `exclude-functions` | `[]` | Go regular expressions matched against fully qualified function names such as `example.com/project/pkg.Function` or `example.com/project/pkg.Type.Method`. |
| `placeholder-patterns` | built-in patterns | Replace the built-in placeholder regular expressions. |
| `additional-placeholder-patterns` | `[]` | Add regular expressions after the selected built-in or replacement patterns. |

Function literals use synthetic names such as
`example.com/project/pkg.<func@file.go:12:10>`. File exclusions are usually more
stable when excluding groups of callbacks.

For a one-off intentional implementation, golangci-lint's normal suppression
mechanism remains available:

```go
func IntentionalNoop() { //nolint:shimbad // Required callback with no work.
}
```

Prefer a narrow function exclusion or a justified `nolint` directive over
disabling a rule globally.

## Development

Format and test the plugin:

```sh
gofmt -w .
go test ./...
go vet ./...
```

The tests use `golang.org/x/tools/go/analysis/analysistest`. Each detected
pattern has accepted counterexamples to protect against broadening a rule
without considering false positives.

Keep the versions of golangci-lint, `plugin-module-register`, and `x/tools`
compatible when updating dependencies.
