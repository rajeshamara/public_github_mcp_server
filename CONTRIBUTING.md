# Contributing

Contributions to Rajesh Amara's independent RetiMCP public edition are welcome. The custom gateway and development tools live in `retimcp/`; the included GitHub MCP Server implementation is credited in [NOTICE.md](NOTICE.md).

Contributions are made under the repository's [MIT license](LICENSE). Preserve existing copyright and attribution notices. Follow the [Code of Conduct](CODE_OF_CONDUCT.md), and use the [security policy](SECURITY.md) for suspected vulnerabilities.

## Proposing a change

Search [this repository's issues](https://github.com/rajeshamara/public_github_mcp_server/issues) before opening a new one. For a substantial feature, describe the use case, expected behavior, and scope before starting implementation. Acceptance or a review timeline is not guaranteed.

Keep each change focused. Use generic examples and synthetic data. Do not commit credentials, personal data, private application code, local configuration, or infrastructure identifiers.

## Development and verification

Follow the [build requirements](README.md#build-and-configure) and use the Go version required by `go.mod`.

For a change to the custom gateway or tools:

```bash
go test ./retimcp/...
go build -buildvcs=false -o bin/retimcp-gateway ./retimcp/cmd/gateway
```

Format changed Go files with `gofmt`. Add or update meaningful tests for changed behavior, particularly authentication, Git state checks, and execution boundaries. Bubblewrap execution tests can skip when the host lacks namespace support; report those skips explicitly rather than treating them as successful execution checks.

For changes to the included upstream code, run the affected package tests and relevant upstream checks. See the local [lint configuration](.golangci.yml), [tool renaming guide](docs/tool-renaming.md), and [toolsets guide](docs/toolsets-and-icons.md) where applicable. UI changes require the relevant build checks in `ui/`. Documentation-only changes do not require a full application test run.

This edition does not include automatic GitHub Actions workflows or Dependabot configuration. Record the checks you actually ran and any limitations in the pull request. Preserve the RetiMCP root README; the upstream documentation generator must not overwrite it.

## Submitting a pull request

1. [Fork this repository](https://github.com/rajeshamara/public_github_mcp_server/fork) and create a branch for your change.
2. Implement the change, update relevant documentation, and run the appropriate checks.
3. [Open a pull request in this repository](https://github.com/rajeshamara/public_github_mcp_server/compare) targeting its default branch.
4. Explain the problem, resulting behavior, validation, and any remaining limitations.

Changes intended for the upstream GitHub project should follow its separate contribution process. A pull request here does not submit a change upstream.
