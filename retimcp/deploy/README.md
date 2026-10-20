# Local build and operator deployment

> **Public demo scope:** This repository is a curated public demo of the privately maintained RetiMCP project. It includes selected code and documentation to showcase the architecture and engineering approach. Private configuration, operational details, and installation-specific integrations are intentionally excluded. This repository is intended for portfolio review and demonstration, and is not a complete distribution of the private project. Upstream components retain their original attribution and licenses.

The public source does not contain a hosted service or preconfigured credentials. All addresses below are loopback values or documentation placeholders.

## Requirements

- Linux, Git, and Go 1.25.12 or newer.
- `/usr/bin/bwrap` (bubblewrap), `/usr/bin/python3`, and working user/network namespaces for the isolated runner.
- Node.js and npm if building the upstream browser UI.
- A dedicated execution worktree on a non-protected branch, with an explicitly configured `origin` and working noninteractive Git authentication.
- A private state directory outside that worktree.

## Build the custom gateway

From the repository root:

```bash
bash retimcp/deploy/build.sh
cp retimcp/config/devbox.example.json retimcp/config/devbox.json
```

The build script runs the custom Go tests and builds the gateway. It does not install services, start listeners, fetch an application repository, or deploy anything.

Edit `devbox.json` locally. Set your positive numeric GitHub account ID; the example's zero ID fails validation. Use the exact worktree path, branch, and origin URL. Keep configuration and state private. For local testing the public origin can be `http://127.0.0.1:8082`; for a remotely reachable endpoint use the HTTPS origin served by your own reverse proxy.

## Build and run the upstream GitHub server

The custom gateway proxies ordinary GitHub MCP requests to `127.0.0.1:8084`. To build that upstream binary and its UI:

```bash
(cd ui && npm ci && npm run build)
go build -buildvcs=false -o bin/github-mcp-server ./cmd/github-mcp-server
```

Run it with a base URL that matches your deployment:

```bash
./bin/github-mcp-server http \
  --listen-host 127.0.0.1 \
  --port 8084 \
  --base-url http://127.0.0.1:8082 \
  --scope-challenge \
  --toolsets repos,git,issues,pull_requests,actions,users
```

See [the upstream README](../../UPSTREAM_README.md) and upstream HTTP documentation for authentication, supported clients, and configuration. Review the selected upstream toolsets and token permissions for your own use case. The gateway preserves upstream request authentication; it does not add an authorization server or grant GitHub permissions.

In a separate process, start the custom gateway:

```bash
./bin/retimcp-gateway -config retimcp/config/devbox.json
```

The gateway binds only to `127.0.0.1:8082`. Its `/dev` endpoint additionally validates the GitHub token and requires the account ID configured in `devbox.json`. Ordinary upstream GitHub tools and the custom execution tools have distinct authorization paths.

## Remote deployment

Use your own TLS reverse proxy to reach the loopback gateway. Set the upstream `--base-url` and local `public_url` to the same public HTTPS origin. Forward MCP streaming requests without response buffering and preserve the Authorization header. Validate your proxy's timeouts against the 90-second operation deadline.

Manage services using your normal deployment process and a dedicated unprivileged account. This public edition includes no automatic service installation, self-deployment, tunnel provisioning, browser credentials, or desktop automation.

## Verification boundaries

`go test ./retimcp/...` exercises the local implementation with mocked token validation and temporary Git repositories. Tests that execute bubblewrap require host namespace support and explicitly skip when unavailable. A passing unit test run does not verify your TLS proxy, GitHub OAuth/client setup, live credentials, remote Git access, or running production deployment; verify those in your own environment.
