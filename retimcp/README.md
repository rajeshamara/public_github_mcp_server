# RetiMCP development endpoint

> **Public edition scope:** This repository contains selected RetiMCP source and documentation for the development gateway. Private configuration, operational details, and installation-specific integrations are excluded; this is not a complete distribution of the privately maintained project. Upstream components retain their original attribution and licenses.

The custom `/dev` endpoint is designed for one operator-approved execution worktree and one authorized GitHub account. Its configuration is loaded at startup; tool callers cannot change paths, origins, branches, credentials, commands, or the test folder.

## Workflow

1. Use the normal source-control workflow to author, review, and commit changes to the configured execution branch.
2. Obtain the full lowercase 40-character commit SHA from that branch.
3. Call `git_status` to inspect the execution worktree.
4. Call `git_pull(expected_commit=SHA)`. The remote branch tip must equal that SHA, and the update must be a fast-forward.
5. Call `git_status` to verify the commit and clean state.
6. Call `run_unit_tests(expected_commit=SHA)` when the committed test folder is present. Inspect `status`, `exit_code`, and output; a successful transport response is not a passing test result.

This endpoint cannot author source files or run arbitrary commands. Refresh or reconnect the client after an operator changes the server's tool catalog; `server_info` reports metadata, not proof that the client has registered a callable tool.

## Configuration and authentication

Copy `config/devbox.example.json` to the ignored `config/devbox.json` and fill in:

| Field | Meaning |
| --- | --- |
| `worktree` | Absolute path to a dedicated real Git worktree; redirected/symlinked roots are refused. |
| `branch` | An explicit execution branch; `main` and `master` are refused. |
| `remote` | The exact configured `origin` URL. |
| `github_user_id` | Positive numeric ID of the only account allowed to use `/dev`. |
| `state_dir` | Private absolute directory outside the execution worktree for locks and audit records. |
| `public_url` | Deployment origin for protected-resource metadata; HTTPS is required except on loopback. |

Tokens are validated against `https://api.github.com/user` with redirects disabled. The returned numeric account ID must match configuration. Tokens are not persisted or included in audit records. The gateway does not issue tokens or implement an authorization server. A compatible MCP client or operator must provide a valid token and manage its permissions.

Git access by the execution service is a separate operator responsibility. The service must already have access to the configured `origin`; interactive prompts are disabled. The incoming MCP bearer token is not passed to Git subprocesses.

## Git boundary

Operations verify the repository root, branch, exact origin, clean state, and absence of in-progress merges/rebases. Dirty, wrong-branch, wrong-origin, ahead/diverged, and busy worktrees are refused. Pull uses a dedicated fetch ref without modifying shared `FETCH_HEAD`, disables hooks and recursive submodule updates, and performs a fast-forward-only merge to the exact expected SHA.

Calls use a cross-process lock, a 90-second operation deadline, bounded command output, and private audit files. The lock coordinates this server's calls only; unrelated processes must not modify the dedicated worktree concurrently. Operators are responsible for reviewing all code and repository configuration before allowing execution.

## Test boundary

`run_unit_tests` supports only Python standard-library `unittest` tests named `test_*.py` in the committed folder `proof_of_work/chatgpt`.

- The worktree must be clean and at the requested commit.
- Only committed regular files in that folder are exported. Symlinks, submodules, ignored local files, and Git metadata are excluded or refused.
- The snapshot is limited to 10 MiB and 500 files. Git archive export-ignore attributes apply.
- Bubblewrap creates isolated namespaces, disables network access, mounts the test snapshot read-only, clears the environment, and does not mount the host home directory.
- Python runs with `-I -B`, a 60-second deadline, resource limits, and a 64 KiB output limit.
- Missing or empty tests fail. Test output is untrusted data.
- Missing bubblewrap or unavailable namespaces cause the production operation to fail; there is no unsandboxed fallback.

The runner does not install dependencies, access application modules outside the test snapshot, connect to a database, perform browser tests, or expose general process execution.

## Source layout

Each tool owns its registration and implementation in `internal/devtools/<tool_name>.go`. Shared configuration, worktree verification, Git execution, locking, and auditing live beside the tools. `http.go` owns transport, owner authentication, and the explicit tool registrations. `server_info.go` lists only tools actually registered by this public edition.

The transport/catalog test uses authenticated MCP requests to verify that the callable tool list matches the advertised catalog. Git tests use temporary local repositories. Authentication tests use fake tokens and mocked HTTP responses; they do not require real credentials.

```bash
go test ./retimcp/...
```

Sandbox execution tests report a skip if the host cannot provide bubblewrap namespaces. Configuration, authentication, Git safety checks, and tool catalog tests remain runnable without that capability.

## Edition scope

The public edition contains `git_status`, `git_pull`, `run_unit_tests`, and `server_info`. Installation-specific browser login, private application regression suites, desktop plugin refresh, and self-deployment integrations are excluded. Their private helper code, credentials, service state, and infrastructure identifiers are not distributed.
