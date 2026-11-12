package cli

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/digitalex/unison/internal/store"
)

// TestPassPrefill: `unison pass` can pre-assign the invited agent's name,
// emoji and model, and print an unused pass's prompt again with other
// values without minting a new pass (#306).
func TestPassPrefill(t *testing.T) {
	e := newTestEnv(t)
	human, agent := t.TempDir(), t.TempDir()
	e.login("Alex", "")
	var ws store.Workspace
	if err := json.Unmarshal([]byte(e.ok(human, "workspace", "create", "Unison", "--json")), &ws); err != nil {
		t.Fatal(err)
	}
	url := e.url + "/w/" + ws.ID
	e.ok(human, "join", url, "--name", "Alex", "--emoji", "🦊", "--account")
	joinLine := regexp.MustCompile(`(?m)^unison join \S+ --pass (up_\S+) (.*)$`)

	out := e.ok(human, "pass", "--name", "Quill", "--model", "Fable 5.2")
	m := joinLine.FindStringSubmatch(out)
	if m == nil || m[2] != "--name Quill --emoji <emoji> --model 'Fable 5.2'" || !strings.Contains(out, "Replace <emoji> with an emoji of your own.") {
		t.Fatalf("pass --name --model:\n%s", out)
	}
	pass := m[1]

	// The same pass, other values: nothing new is minted.
	again := joinLine.FindStringSubmatch(e.ok(human, "pass", "--pass", pass, "--name", "Lark", "--emoji", "🐦"))
	if again == nil || again[1] != pass || again[2] != "--name Lark --emoji '🐦' --model <model>" {
		t.Fatalf("pass --pass = %v", again)
	}
	var resp struct {
		Pass         string `json:"pass"`
		InvitePrompt string `json:"invite_prompt"`
	}
	if err := json.Unmarshal([]byte(e.ok(human, "pass", "--pass", pass, "--json")), &resp); err != nil || resp.Pass != pass || !strings.Contains(resp.InvitePrompt, "--name <YourName>") {
		t.Errorf("pass --pass --json = %+v, %v", resp, err)
	}

	// The server's refusals come through.
	if _, errOut, code := e.run(human, "pass", "--name", "two words"); code == 0 || !strings.Contains(errOut, "names cannot contain whitespace") {
		t.Errorf("pass --name 'two words': exit %d %q", code, errOut)
	}
	if _, errOut, code := e.run(human, "pass", "--model", "$(id)"); code == 0 || !strings.Contains(errOut, "shell characters") {
		t.Errorf("pass --model '$(id)': exit %d %q", code, errOut)
	}

	// The printed line works as pasted (the quotes are the shell's to remove).
	e.ok(agent, "join", url, "--pass", pass, "--name", "Lark", "--emoji", "🐦", "--model", "Fable 5.2")
	if out := e.ok(human, "who"); !strings.Contains(out, "🐦 Lark (Fable 5.2) · agent of Alex") {
		t.Errorf("who = %q", out)
	}
	if _, errOut, code := e.run(human, "pass", "--pass", pass); code == 0 || !strings.Contains(errOut, "already used") {
		t.Errorf("pass --pass on a used pass: exit %d %q", code, errOut)
	}
}
