# Security policy

This policy applies to Rajesh Amara's independent RetiMCP public edition. The upstream GitHub MCP Server is credited in [NOTICE.md](NOTICE.md); GitHub's security team does not maintain this edition or its custom gateway.

## Reporting a suspected vulnerability

Do not include vulnerability details, exploit code, credentials, private repository data, or personal information in public issues, discussions, or pull requests.

Use a private vulnerability reporting option only if this repository explicitly offers one. This repository does not publish a dedicated security email address or promise that private reporting is enabled. If no private option is available, open an issue containing only a request for a private security contact. Wait for the maintainer to establish a suitable channel before sharing sensitive details.

Once a private channel is agreed, useful information includes the affected commit and component, a description of the impact, and a minimal reproduction using synthetic data. Never send working credentials or data from another person's systems.

For an issue confined to the original upstream implementation, consult the upstream project's own security policy. Do not assume that an upstream reporting channel accepts reports about RetiMCP additions or a separately operated deployment.

## Maintenance expectations

This is an independently maintained project. There is no guaranteed response or remediation timeline, published supported-version schedule, or bounty program for this edition. This policy does not extend any upstream legal safe-harbor commitment to the project.

Operators are responsible for reviewing changes, updating dependencies, protecting credentials, and deploying the gateway with appropriate isolation and access controls. See the [development endpoint guide](retimcp/README.md) for the implemented boundaries and limitations.
