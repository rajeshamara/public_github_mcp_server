package devtools

import (
	"context"
	"errors"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerGitPull keeps the git_pull MCP contract beside its implementation.
func registerGitPull(server *mcp.Server, s *Service) {
	no := false
	mcp.AddTool(server, &mcp.Tool{Name: "git_pull", Description: "Pulls a commit already created through your normal GitHub contribution workflow on the configured execution branch. Do not use this tool to author code. Expected flow: edit GitHub -> commit -> obtain full SHA -> call git_pull(expected_commit=SHA). Fast-forwards only the dedicated configured execution host execution worktree to that exact commit. Refuses dirty, busy, ahead, diverged, or unexpected branches. Does not create or edit source files independently of that commit, run tests, or restart services.", Annotations: &mcp.ToolAnnotations{DestructiveHint: &no, IdempotentHint: true}}, func(ctx context.Context, _ *mcp.CallToolRequest, a PullArgs) (*mcp.CallToolResult, Result, error) {
		r, e := s.Pull(ctx, a.ExpectedCommit)
		return nil, r, e
	})

}

// Pull fast-forwards only to the expected remote commit, preserving all existing checks.
func (s *Service) Pull(ctx context.Context, expected string) (Result, error) {
	if !shaPattern.MatchString(expected) {
		return Result{}, errors.New("expected_commit must be a full lowercase 40-character commit SHA")
	}
	return s.locked(ctx, "git_pull", func(ctx context.Context) (Result, error) {
		before, err := s.inspect(ctx)
		if err != nil {
			return before, err
		}
		if !before.Clean {
			return before, errors.New("dirty worktree; refusing to pull")
		}
		ref := "refs/retimcp/" + s.Config.Branch
		_, err = s.git(ctx, "fetch", "--no-tags", "--no-write-fetch-head", "origin", "+refs/heads/"+s.Config.Branch+":"+ref)
		if err != nil {
			return before, err
		}
		target, err := s.git(ctx, "rev-parse", ref)
		if err != nil {
			return before, err
		}
		if target != expected {
			return before, errors.New("remote branch moved or expected_commit is incorrect; inspect GitHub and retry")
		}
		if _, err = s.git(ctx, "merge-base", "--is-ancestor", before.Commit, target); err != nil {
			return before, errors.New("local branch is ahead or diverged; refusing non-fast-forward update")
		}
		current, err := s.inspect(ctx)
		if err != nil {
			return current, err
		}
		if !current.Clean || current.Commit != before.Commit {
			return current, errors.New("worktree changed during fetch; refusing update")
		}
		if _, err = s.git(ctx, "merge", "--ff-only", "--no-edit", target); err != nil {
			return before, err
		}
		after, err := s.inspect(ctx)
		if err != nil {
			return after, err
		}
		after.Before = before.Commit
		after.Status = "updated"
		if after.Commit == before.Commit {
			after.Status = "up_to_date"
		}
		if after.Commit != expected || !after.Clean {
			return after, errors.New("post-pull verification failed")
		}
		return after, nil
	})
}
