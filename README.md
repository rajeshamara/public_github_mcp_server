# RetiMCP — a controlled development gateway for GitHub MCP

RetiMCP is an independent extension of the [GitHub MCP Server](https://github.com/github/github-mcp-server). It adds an authenticated execution endpoint for inspecting a dedicated Git worktree, pulling an exact reviewed commit, and running a small Python unit-test snapshot in an isolated Linux environment.

The project demonstrates a separation between **authoring code** and **executing a known revision**. An AI client uses its normal GitHub workflow to create a commit; RetiMCP accepts the full commit SHA and verifies the configured worktree before performing a limited operation.

## What Rajesh built

The custom implementation lives in [`retimcp/`](retimcp/):

- A Go HTTP gateway with an embedded public landing page and a reverse proxy to the upstream GitHub MCP server.
- An owner-restricted `/dev` MCP endpoint that validates a GitHub bearer token against GitHub and compares the account's numeric ID with explicit operator configuration.
- Exact-commit Git synchronization with branch/origin verification, dirty-worktree refusal, fast-forward-only updates, cross-process locking, and an audit trail.
- A Python standard-library test runner that exports committed files into a read-only bubblewrap sandbox, with no network or mounted home directory, bounded output, time limits, and resource limits.
- A protocol-level tool catalog and tests covering authentication, unsafe Git states, configuration, and execution boundaries.

## Public tool catalog

| Tool | Behavior |
| --- | --- |
| `git_status` | Inspect the configured worktree, branch, commit, and clean state. |
| `git_pull(expected_commit)` | Fast-forward the configured execution branch only when its remote tip matches the full expected SHA. |
| `run_unit_tests(expected_commit)` | Run committed `test_*.py` files under `proof_of_work/chatgpt` in the isolated runner. |
| `server_info` | Return the running version, tool catalog, capabilities, and limits. |

The public edition intentionally excludes installation-specific browser login, desktop plugin refresh, application-specific regression suites, and self-deployment helpers. Those tools depended on private application code, credentials, or infrastructure. They are not advertised as portable functionality.

## Architecture

The gateway listens on `127.0.0.1:8082`. It serves the landing page for browser requests, routes `/dev` to the custom authenticated tools, and forwards other MCP requests to the upstream server on `127.0.0.1:8084`. An operator may place a TLS reverse proxy in front of these loopback services.

The custom endpoint has a separate identity check from the upstream server. It does not provide an OAuth authorization server; a compatible client or operator must supply a valid GitHub token for the configured owner. Deployment and client authentication setup are described in [`retimcp/deploy/README.md`](retimcp/deploy/README.md).

## Build and configure

Requirements for the custom gateway: Linux, Git, and Go 1.25.12 or newer. Running sandboxed tests also requires `/usr/bin/bwrap`, `/usr/bin/python3`, and working Linux user/network namespaces. Building the upstream browser UI additionally requires Node.js and npm.

```bash
go build -buildvcs=false -o bin/retimcp-gateway ./retimcp/cmd/gateway
cp retimcp/config/devbox.example.json retimcp/config/devbox.json
```

Edit the local configuration with a dedicated worktree, its exact `origin` URL, a non-protected execution branch, your numeric GitHub user ID, a private state directory outside that worktree, and your deployment origin. `github_user_id: 0` is deliberately invalid: the example cannot authorize an account until configured. The local `devbox.json` is ignored by Git.

Start the gateway only after configuring and reviewing the local environment:

```bash
./bin/retimcp-gateway -config retimcp/config/devbox.json
```

See the [deployment guide](retimcp/deploy/README.md) for the upstream server and reverse proxy setup, and the [development endpoint guide](retimcp/README.md) for the execution workflow.

## Verification

```bash
go test ./retimcp/...
go build -buildvcs=false -o bin/retimcp-gateway ./retimcp/cmd/gateway
```

Sandbox execution tests require functioning bubblewrap namespaces. Tests report a skip when the environment cannot provide that sandbox. The production runner fails closed when it cannot start the sandbox; it never substitutes unsandboxed execution. Git tests use temporary local repositories, and authentication tests use mocked HTTP responses.

## Scope and limitations

This is a development tool, not a general command runner or a claim of production certification. The test runner supports only Python standard-library tests in the fixed committed folder; it does not install dependencies or run application integration tests. Operators remain responsible for reviewing code, managing service credentials and token permissions, providing TLS for remote use, and keeping the dedicated worktree free of concurrent edits by other processes.

## Upstream attribution

This repository contains GitHub's MCP Server source from public upstream commit [`6e00cee58c808b352a05c9e84d8d490eb61349d5`](https://github.com/github/github-mcp-server/tree/6e00cee58c808b352a05c9e84d8d490eb61349d5), together with Rajesh Amara's RetiMCP additions. The upstream implementation remains under the [MIT license](LICENSE), with the original copyright notice retained. See [UPSTREAM_README.md](UPSTREAM_README.md) for upstream documentation and [NOTICE.md](NOTICE.md) for provenance. This is an independent project and is not an official GitHub product release.

The public repository starts from a reviewed source snapshot with fresh history. Deployment state, private configuration, operational logs, private application code, and the original private Git history are not included.
