package devtools

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	Worktree     string `json:"worktree"`
	Branch       string `json:"branch"`
	Remote       string `json:"remote"`
	GitHubUserID int64  `json:"github_user_id"`
	StateDir     string `json:"state_dir"`
	PublicURL    string `json:"public_url"`
}
type Result struct {
	RunID     string `json:"run_id"`
	Operation string `json:"operation"`
	Devbox    string `json:"devbox"`
	Worktree  string `json:"worktree"`
	Branch    string `json:"branch"`
	Before    string `json:"before_commit,omitempty"`
	Commit    string `json:"commit"`
	Clean     bool   `json:"clean"`
	Status    string `json:"status"`
}
type Service struct{ Config Config }

var shaPattern = regexp.MustCompile("^[0-9a-f]{40}$")

type boundedBuffer struct{ bytes.Buffer }

func (b *boundedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	left := 65536 - b.Len()
	if left > 0 {
		if len(p) > left {
			p = p[:left]
		}
		_, _ = b.Buffer.Write(p)
	}
	return n, nil
}
func (s *Service) git(ctx context.Context, args ...string) (string, error) {
	base := []string{"-c", "core.hooksPath=/dev/null", "-c", "submodule.recurse=false", "-C", s.Config.Worktree}
	cmd := exec.CommandContext(ctx, "git", append(base, args...)...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_SSH_COMMAND=ssh -o BatchMode=yes -o StrictHostKeyChecking=yes -o ConnectTimeout=10")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = 2 * time.Second
	var out boundedBuffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("operation timed out or cancelled")
		}
		return "", fmt.Errorf("git %s failed; inspect server-side repository configuration", args[0])
	}
	return strings.TrimSpace(out.String()), nil
}
func (s *Service) inspect(ctx context.Context) (Result, error) {
	r := Result{Devbox: "configured-host", Worktree: s.Config.Worktree, Branch: s.Config.Branch, Status: "ready"}
	root, err := filepath.EvalSymlinks(s.Config.Worktree)
	if err != nil || root != filepath.Clean(s.Config.Worktree) {
		return r, errors.New("worktree path is missing or redirected")
	}
	actual, err := s.git(ctx, "rev-parse", "--show-toplevel")
	if err != nil || actual != root {
		return r, errors.New("unexpected repository root")
	}
	branch, err := s.git(ctx, "symbolic-ref", "--short", "HEAD")
	if err != nil || branch != s.Config.Branch || branch == "main" || branch == "master" {
		return r, errors.New("unexpected or protected branch")
	}
	remote, err := s.git(ctx, "remote", "get-url", "origin")
	if err != nil || remote != s.Config.Remote {
		return r, errors.New("unexpected origin remote")
	}
	for _, marker := range []string{"MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD", "rebase-merge", "rebase-apply"} {
		p, e := s.git(ctx, "rev-parse", "--git-path", marker)
		if e != nil {
			return r, e
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(root, p)
		}
		if _, e = os.Stat(p); e == nil {
			return r, errors.New("repository operation already in progress")
		} else if !os.IsNotExist(e) {
			return r, e
		}
	}
	r.Commit, err = s.git(ctx, "rev-parse", "HEAD")
	if err != nil {
		return r, err
	}
	dirty, err := s.git(ctx, "status", "--porcelain=v1", "--untracked-files=normal")
	if err != nil {
		return r, err
	}
	r.Clean = dirty == ""
	return r, nil
}
func (s *Service) locked(ctx context.Context, op string, fn func(context.Context) (Result, error)) (Result, error) {
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	if err := os.MkdirAll(s.Config.StateDir, 0700); err != nil {
		return Result{}, err
	}
	lock, err := os.OpenFile(filepath.Join(s.Config.StateDir, "worktree.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return Result{}, err
	}
	defer lock.Close()
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return Result{}, errors.New("worktree busy; retry after current operation finishes")
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	audit, err := os.OpenFile(filepath.Join(s.Config.StateDir, "runs.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return Result{}, err
	}
	defer audit.Close()
	id := make([]byte, 12)
	if _, err = io.ReadFull(rand.Reader, id); err != nil {
		return Result{}, err
	}
	runID := hex.EncodeToString(id)
	if err = json.NewEncoder(audit).Encode(map[string]any{"run_id": runID, "operation": op, "status": "started", "at": time.Now().UTC(), "github_user_id": s.Config.GitHubUserID}); err != nil {
		return Result{}, err
	}
	r, callErr := fn(ctx)
	r.RunID = runID
	r.Operation = op
	entry := map[string]any{"run": r, "at": time.Now().UTC(), "github_user_id": s.Config.GitHubUserID}
	if callErr != nil {
		entry["error"] = callErr.Error()
	}
	if err = json.NewEncoder(audit).Encode(entry); err != nil {
		return r, fmt.Errorf("operation completed but audit write failed: %w", err)
	}
	return r, callErr
}
