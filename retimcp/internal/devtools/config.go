package devtools

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

var branchPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)

// Validate accepts only explicit operator configuration. MCP callers cannot
// supply or override these values through tool arguments.
func (c Config) Validate() error {
	if c.GitHubUserID <= 0 || c.Remote == "" ||
		!filepath.IsAbs(c.Worktree) || !filepath.IsAbs(c.StateDir) ||
		filepath.Clean(c.Worktree) != c.Worktree || filepath.Clean(c.StateDir) != c.StateDir ||
		c.Worktree == "/" || c.StateDir == "/" {
		return fmt.Errorf("set an authorized GitHub user ID, origin, and absolute worktree/state paths")
	}
	if !branchPattern.MatchString(c.Branch) || c.Branch == "main" || c.Branch == "master" ||
		strings.Contains(c.Branch, "..") || strings.Contains(c.Branch, "//") ||
		strings.HasSuffix(c.Branch, "/") || strings.HasSuffix(c.Branch, ".") || strings.HasSuffix(c.Branch, ".lock") {
		return fmt.Errorf("configure an explicit non-protected execution branch")
	}
	if c.StateDir == c.Worktree || strings.HasPrefix(c.StateDir, c.Worktree+string(filepath.Separator)) {
		return fmt.Errorf("state_dir must be outside the execution worktree")
	}
	u, err := url.Parse(c.PublicURL)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" ||
		(u.Path != "" && u.Path != "/") {
		return fmt.Errorf("public_url must be an origin without credentials, path, query, or fragment")
	}
	loopback := u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1"
	if u.Scheme != "https" && !(u.Scheme == "http" && loopback) {
		return fmt.Errorf("public_url requires HTTPS except on loopback")
	}
	return nil
}
