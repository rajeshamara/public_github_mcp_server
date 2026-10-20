package devtools

// Shared argument types keep the existing input schemas identical across tools.
type PullArgs struct {
	ExpectedCommit string `json:"expected_commit" jsonschema:"Full commit SHA that ChatGPT committed to the configured origin branch. Pull refuses a different remote SHA."`
}
type EmptyArgs struct{}
