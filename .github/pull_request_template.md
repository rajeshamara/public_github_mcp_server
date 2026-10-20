<!--
Copilot: Fill all sections. Prefer short, concrete answers.
If a checkbox is selected, add a brief explanation.
-->

## Summary
<!-- In 1–2 sentences: what does this PR do? -->

## Why
<!-- Why is this change needed? Link issues or discussions. -->
Fixes #

## What changed
<!-- Bullet list of concrete changes. -->
- 
- 

## MCP impact
<!-- Select one or more. If selected, add 1–2 sentences. -->
- [ ] No tool or API changes
- [ ] Tool schema or behavior changed
- [ ] New tool added

## Prompts tested (tool changes only)
<!-- If you changed or added tools, list example prompts you tested. -->
<!-- Include prompts that trigger the tool and describe the use case. -->
<!-- Example: "List all open issues in the repo assigned to me" -->
- 

## Security / limits
<!-- Select if relevant. Add a short note if checked. -->
- [ ] No security or limits impact
- [ ] Auth / permissions considered
- [ ] Data exposure, filtering, or token/size limits considered

## Tool renaming
- [ ] I am renaming tools as part of this PR (e.g. a part of a consolidation effort)
   - [ ] I have added the new tool aliases in `deprecated_tool_aliases.go` 
- [ ] I am not renaming tools as part of this PR

For renames in the included upstream toolsets, follow the tool renaming guide in `docs/tool-renaming.md`. For custom RetiMCP tools, keep the callable catalog, `server_info`, documentation, and tests consistent.

## Lint & tests
<!-- Check only what you ran. Explain skipped or unavailable checks. This edition has no automatic CI workflows. -->
- [ ] Custom gateway tests: `go test ./retimcp/...`
- [ ] Custom gateway build: `go build -buildvcs=false -o bin/retimcp-gateway ./retimcp/cmd/gateway`
- [ ] Linted locally with `./script/lint`
- [ ] Tested locally with `./script/test`

Sandbox tests skipped, other checks, and limitations:

## Docs

- [ ] Not needed
- [ ] Updated (README / docs / examples)
