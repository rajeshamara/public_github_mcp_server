package devtools

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitTest(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.invalid", "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.invalid")
	b, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("git %v: %v %s", args, e, b)
	}
	return strings.TrimSpace(string(b))
}
func fixture(t *testing.T) (*Service, string, string) {
	t.Helper()
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	source := filepath.Join(root, "source")
	work := filepath.Join(root, "work")
	gitTest(t, root, "init", "--bare", remote)
	gitTest(t, root, "init", "-b", "agent-execution", source)
	os.WriteFile(filepath.Join(source, "hello.txt"), []byte("one"), 0600)
	gitTest(t, source, "add", ".")
	gitTest(t, source, "commit", "-m", "initial")
	gitTest(t, source, "remote", "add", "origin", remote)
	gitTest(t, source, "push", "origin", "agent-execution")
	gitTest(t, root, "clone", "-b", "agent-execution", remote, work)
	return &Service{Config: Config{Worktree: work, Branch: "agent-execution", Remote: remote, StateDir: filepath.Join(root, "state"), GitHubUserID: 1, PublicURL: "https://mcp.example.invalid"}}, source, gitTest(t, source, "rev-parse", "HEAD")
}
func advance(t *testing.T, source string) string {
	t.Helper()
	os.WriteFile(filepath.Join(source, "hello.txt"), []byte("two"), 0600)
	gitTest(t, source, "add", ".")
	gitTest(t, source, "commit", "-m", "second")
	gitTest(t, source, "push", "origin", "agent-execution")
	return gitTest(t, source, "rev-parse", "HEAD")
}
