package devtools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func commitTests(t *testing.T, s *Service, code string) string {
	t.Helper()
	dir := filepath.Join(s.Config.Worktree, testFolder)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "test_example.py"), []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	gitTest(t, s.Config.Worktree, "add", ".")
	gitTest(t, s.Config.Worktree, "commit", "-m", "test fixture")
	return gitTest(t, s.Config.Worktree, "rev-parse", "HEAD")
}
func TestUnitTestFailureFixAndIsolation(t *testing.T) {
	requireSandbox(t)
	s, _, _ := fixture(t)
	sha := commitTests(t, s, "import unittest\nclass Example(unittest.TestCase):\n def test_add(self): self.assertEqual(2-3,5)\n")
	r, e := s.RunUnitTests(context.Background(), sha)
	if e != nil || r.Status != "failed" || r.ExitCode != 1 || !strings.Contains(r.Output, "FAIL") {
		t.Fatalf("%+v %v", r, e)
	}
	t.Setenv("RETI_TEST_SECRET", "not-for-child")
	code := `import os, socket, unittest
class Example(unittest.TestCase):
 def test_add(self): self.assertEqual(2+3,5)
 def test_environment(self):
  self.assertNotIn("RETI_TEST_SECRET",os.environ)
  self.assertFalse(os.path.exists("/home"))
  self.assertFalse(os.path.exists("/tests/.git"))
 def test_readonly(self):
  with self.assertRaises(OSError):
   open("/tests/mutation","w")
 def test_network(self):
  s=socket.socket()
  s.settimeout(0.1)
  try:
   with self.assertRaises(OSError): s.connect(("1.1.1.1",443))
  finally: s.close()
`
	sha = commitTests(t, s, code)
	r, e = s.RunUnitTests(context.Background(), sha)
	if e != nil || r.Status != "passed" || r.ExitCode != 0 || !strings.Contains(r.Output, "Ran 4 tests") || !r.Clean || r.RunID == "" {
		t.Fatalf("%+v %v", r, e)
	}
}
func TestUnitTestRefusals(t *testing.T) {
	for _, mode := range []string{"wrong_sha", "dirty", "missing", "symlink", "no_tests"} {
		t.Run(mode, func(t *testing.T) {
			if mode == "no_tests" {
				requireSandbox(t)
			}
			s, _, old := fixture(t)
			sha := old
			if mode != "missing" {
				sha = commitTests(t, s, "# no tests\n")
			}
			switch mode {
			case "wrong_sha":
				sha = old
			case "dirty":
				os.WriteFile(filepath.Join(s.Config.Worktree, "dirty"), []byte("x"), 0600)
			case "symlink":
				os.Symlink("/etc/passwd", filepath.Join(s.Config.Worktree, testFolder, "secret"))
				gitTest(t, s.Config.Worktree, "add", ".")
				gitTest(t, s.Config.Worktree, "commit", "-m", "symlink")
				sha = gitTest(t, s.Config.Worktree, "rev-parse", "HEAD")
			}
			r, e := s.RunUnitTests(context.Background(), sha)
			if mode == "no_tests" {
				if e != nil || r.Status != "failed" || r.ExitCode != 5 {
					t.Fatalf("%+v %v", r, e)
				}
			} else if e == nil {
				t.Fatalf("refusal expected: %+v", r)
			}
		})
	}
}
func TestUnitTestTimeoutAndOutputBound(t *testing.T) {
	requireSandbox(t)
	s, _, _ := fixture(t)
	sha := commitTests(t, s, "import time\ntime.sleep(10)\n")
	r, e := s.runUnitTests(context.Background(), sha, 200*time.Millisecond)
	if e != nil || r.Status != "timed_out_or_cancelled" {
		t.Fatalf("%+v %v", r, e)
	}
	sha = commitTests(t, s, "import unittest\nprint('x'*100000)\nclass Example(unittest.TestCase):\n def test_ok(self): pass\n")
	r, e = s.RunUnitTests(context.Background(), sha)
	if e != nil || r.Status != "passed" || !r.OutputTruncated || len(r.Output) != 65536 {
		t.Fatalf("status=%s len=%d truncated=%v error=%v", r.Status, len(r.Output), r.OutputTruncated, e)
	}
}
