package server

import (
	"strings"
	"testing"
	"time"

	"github.com/digitalex/unison/internal/store"
)

// A member may pre-assign the invited agent's name, emoji and model: they
// are checked, quoted into the join line, and can be changed for an unused
// pass without minting another (#306).
func TestPassPrefill(t *testing.T) {
	e := newAuthEnv(t)
	ctx := t.Context()
	ws, _, _ := e.st.CreateWorkspace(ctx, "Unison", "https://github.com/digitalex/unison", "")
	other, _, _ := e.st.CreateWorkspace(ctx, "Other", "", "")
	base := "/api/workspaces/" + ws.ID
	ada, _ := e.login("github", "/", adaGitHub, adaGitHubEmails)
	bob, _ := e.login("google", "/", bobGoogle, nil)
	e.memberJoins(ada, ws.ID, "Ada")
	e.memberJoins(ada, other.ID, "Ada")
	e.memberJoins(bob, ws.ID, "Bob")
	asAda := []reqOpt{withCookie(ada), withCSRF()}
	joinLine := func(prompt any) string {
		t.Helper()
		for _, line := range strings.Split(prompt.(string), "\n") {
			if strings.HasPrefix(line, "unison join ") {
				return line
			}
		}
		t.Fatalf("no join line in %q", prompt)
		return ""
	}
	url := e.ts.URL + "/w/" + ws.ID

	// Nothing given: every flag is a placeholder, and the agent is asked for all three.
	resp, out := e.do("POST", base+"/passes", "", asAda...)
	pass, _ := out["pass"].(string)
	prompt, _ := out["invite_prompt"].(string)
	if resp.StatusCode != 201 || joinLine(prompt) != "unison join "+url+" --pass "+pass+" --name <YourName> --emoji <emoji> --model <model>" {
		t.Fatalf("plain mint = %d %q", resp.StatusCode, prompt)
	}
	for _, want := range []string{
		"Replace <YourName> with a name of your own, <emoji> with an emoji of your own and <model> with the model you run on.",
		"Then run `unison --agent-help` and follow it.",
		"expires in 15 minutes", "agent of Ada Lovelace", url + "/agent.md?pass=" + pass,
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt lacks %q:\n%s", want, prompt)
		}
	}
	// It is short: the laps live in the guide, not here.
	if strings.Contains(prompt, "auto-merge") || strings.Contains(prompt, "unison wait") || len(prompt) > 700 {
		t.Errorf("prompt repeats the guide (%d characters):\n%s", len(prompt), prompt)
	}

	// Re-rendering the same pass with values: quoted where a shell needs it,
	// and the agent is asked only for what is still blank.
	resp, out = e.do("POST", base+"/passes/prompt", `{"pass":"`+pass+`","name":" Quill ","emoji":"🪶"}`, asAda...)
	prompt, _ = out["invite_prompt"].(string)
	if resp.StatusCode != 200 || out["pass"] != pass || out["expires_at"] == nil ||
		joinLine(prompt) != "unison join "+url+" --pass "+pass+" --name Quill --emoji '🪶' --model <model>" ||
		!strings.Contains(prompt, "Replace <model> with the model you run on. Then run") || strings.Contains(prompt, "<YourName> with") {
		t.Fatalf("re-render = %d %q", resp.StatusCode, prompt)
	}
	_, out = e.do("POST", base+"/passes/prompt", `{"pass":"`+pass+`","name":"Quill","emoji":"🪶","model":"Fable 5.2"}`, asAda...)
	prompt, _ = out["invite_prompt"].(string)
	if joinLine(prompt) != "unison join "+url+" --pass "+pass+" --name Quill --emoji '🪶' --model 'Fable 5.2'" || strings.Contains(prompt, "Replace") {
		t.Fatalf("all three given: %q", prompt)
	}
	// Minting with values does the same in one step.
	_, out = e.do("POST", base+"/passes", `{"name":"Lark","model":"sonnet-5"}`, asAda...)
	second, _ := out["pass"].(string)
	if second == pass || joinLine(out["invite_prompt"]) != "unison join "+url+" --pass "+second+" --name Lark --emoji <emoji> --model sonnet-5" {
		t.Fatalf("mint with values = %v", out)
	}

	// Values that break the name rules or a shell are refused.
	for _, bad := range []string{
		`{"name":"two words"}`, `{"name":"a@b"}`, `{"name":"O'Brien"}`, `{"name":"$(id)"}`, `{"name":"a;b"}`,
		`{"emoji":"🦊 🦊"}`, `{"emoji":"` + "`x`" + `"}`, `{"model":"a\nb"}`, `{"model":"x\"y"}`, `{"model":"` + strings.Repeat("m", 65) + `"}`,
	} {
		if resp, out := e.do("POST", base+"/passes", bad, asAda...); resp.StatusCode != 400 {
			t.Errorf("mint %s = %d %v, want 400", bad, resp.StatusCode, out)
		}
		if resp, out := e.do("POST", base+"/passes/prompt", `{"pass":"`+pass+`",`+bad[1:], asAda...); resp.StatusCode != 400 {
			t.Errorf("re-render %s = %d %v, want 400", bad, resp.StatusCode, out)
		}
	}

	// Only the member who made a pass may re-render it, only in its
	// workspace, and only until it is used.
	for name, tc := range map[string]struct {
		path string
		opts []reqOpt
		want int
	}{
		"anonymous":         {base, nil, 401},
		"no CSRF header":    {base, []reqOpt{withCookie(ada)}, 403},
		"another member":    {base, []reqOpt{withCookie(bob), withCSRF()}, 403},
		"another workspace": {"/api/workspaces/" + other.ID, asAda, 403},
	} {
		if resp, out := e.do("POST", tc.path+"/passes/prompt", `{"pass":"`+pass+`"}`, tc.opts...); resp.StatusCode != tc.want {
			t.Errorf("re-render (%s) = %d %v, want %d", name, resp.StatusCode, out, tc.want)
		}
	}
	if resp, out := e.do("POST", base+"/passes/prompt", `{"pass":"up_nope"}`, asAda...); resp.StatusCode != 403 || !strings.Contains(out["error"].(string), "unknown pass") {
		t.Errorf("unknown pass = %d %v", resp.StatusCode, out)
	}
	// The agent joins with the pre-assigned values; after that the pass is spent.
	resp, joined := e.do("POST", base+"/join", `{"pass":"`+pass+`","name":"Quill","emoji":"🪶","model":"Fable 5.2"}`)
	if agent, _ := joined["participant"].(map[string]any); resp.StatusCode != 201 || agent["name"] != "Quill" || agent["model"] != "Fable 5.2" {
		t.Fatalf("join = %d %v", resp.StatusCode, joined)
	}
	if resp, out := e.do("POST", base+"/passes/prompt", `{"pass":"`+pass+`"}`, asAda...); resp.StatusCode != 403 || !strings.Contains(out["error"].(string), "already used") {
		t.Errorf("re-render of a used pass = %d %v", resp.StatusCode, out)
	}
	// An expired one too.
	e.st.SetClock(func() time.Time { return time.Now().Add(store.AgentPassTTL + time.Minute) })
	if resp, out := e.do("POST", base+"/passes/prompt", `{"pass":"`+second+`"}`, asAda...); resp.StatusCode != 403 || !strings.Contains(out["error"].(string), "expired") {
		t.Errorf("re-render of an expired pass = %d %v", resp.StatusCode, out)
	}
}
