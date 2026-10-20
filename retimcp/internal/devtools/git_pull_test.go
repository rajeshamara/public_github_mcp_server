package devtools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestPullFastForwardAndIdempotent(t *testing.T) {
	s, src, old := fixture(t)
	next := advance(t, src)
	r, e := s.Pull(context.Background(), next)
	if e != nil {
		t.Fatal(e)
	}
	if r.Commit != next || r.Before != old || r.Status != "updated" || !r.Clean || r.RunID == "" {
		t.Fatalf("%+v", r)
	}
	r, e = s.Pull(context.Background(), next)
	if e != nil || r.Status != "up_to_date" {
		t.Fatalf("%+v %v", r, e)
	}
}
func TestPullRefusals(t *testing.T) {
	for _, mode := range []string{"dirty", "wrong_sha", "diverged", "wrong_branch", "wrong_remote", "symlink", "busy"} {
		t.Run(mode, func(t *testing.T) {
			s, src, old := fixture(t)
			next := advance(t, src)
			switch mode {
			case "dirty":
				os.WriteFile(filepath.Join(s.Config.Worktree, "untracked"), []byte("preserve"), 0600)
			case "wrong_sha":
				next = old
			case "diverged":
				os.WriteFile(filepath.Join(s.Config.Worktree, "local"), []byte("preserve"), 0600)
				gitTest(t, s.Config.Worktree, "add", ".")
				gitTest(t, s.Config.Worktree, "commit", "-m", "local")
			case "wrong_branch":
				gitTest(t, s.Config.Worktree, "checkout", "-b", "main")
			case "wrong_remote":
				s.Config.Remote = "https://example.invalid/other.git"
			case "symlink":
				link := s.Config.Worktree + "-link"
				os.Symlink(s.Config.Worktree, link)
				s.Config.Worktree = link
			case "busy":
				os.MkdirAll(s.Config.StateDir, 0700)
				f, e := os.OpenFile(filepath.Join(s.Config.StateDir, "worktree.lock"), os.O_CREATE|os.O_RDWR, 0600)
				if e != nil {
					t.Fatal(e)
				}
				defer f.Close()
				syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
				defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
			}
			before := gitTest(t, s.Config.Worktree, "rev-parse", "HEAD")
			if _, e := s.Pull(context.Background(), next); e == nil {
				t.Fatal("unsafe pull allowed")
			}
			if after := gitTest(t, s.Config.Worktree, "rev-parse", "HEAD"); after != before {
				t.Fatal("refused operation changed HEAD")
			}
		})
	}
}
func TestInvalidCommit(t *testing.T) {
	s, _, _ := fixture(t)
	for _, v := range []string{"", "main", ";touch /tmp/pwn", strings.Repeat("a", 41)} {
		if _, e := s.Pull(context.Background(), v); e == nil {
			t.Fatal(v)
		}
	}
}
