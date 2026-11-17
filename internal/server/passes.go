package server

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/digitalex/unison/internal/store"
)

// passRequest is the body of the pass endpoints: what the member
// pre-assigns to the agent (#306), each optional, and for a re-render the
// pass itself.
type passRequest struct {
	Pass  string `json:"pass"`
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
	Model string `json:"model"`
}

// mintPass creates a single-use agent pass for the signed-in member and
// returns it with a ready-to-paste invite prompt. The body may pre-assign
// the agent's name, emoji and model: they go into the prompt's join line.
func (s *Server) mintPass(w http.ResponseWriter, r *http.Request) {
	var req passRequest
	if !decode(w, r, &req) {
		return
	}
	who, err := store.CleanPrefill(store.AgentJoin{Name: req.Name, Emoji: req.Emoji, Model: req.Model})
	if err != nil {
		storeError(w, err)
		return
	}
	ws, err := s.store.GetWorkspace(r.Context(), r.PathValue("ws"))
	if err != nil {
		storeError(w, err)
		return
	}
	u := signedIn(r).user
	pass, expires, err := s.store.MintPass(r.Context(), ws.ID, u.ID)
	if err != nil {
		storeError(w, err)
		return
	}
	s.track(eventPassGenerated, userActor(u.ID))
	writeJSON(w, http.StatusCreated, map[string]any{
		"pass":          pass,
		"expires_at":    expires,
		"invite_prompt": invitePrompt(requestOrigin(r), ws, firstNonEmpty(u.DisplayName, "a member"), pass, who),
	})
}

// passPrompt writes the invite prompt again for a pass the member minted
// and nobody has used, with other pre-assigned values (#306): the invite
// dialog calls it as the member types, so typing doesn't burn passes.
func (s *Server) passPrompt(w http.ResponseWriter, r *http.Request) {
	var req passRequest
	if !decode(w, r, &req) {
		return
	}
	who, err := store.CleanPrefill(store.AgentJoin{Name: req.Name, Emoji: req.Emoji, Model: req.Model})
	if err != nil {
		storeError(w, err)
		return
	}
	ws, err := s.store.GetWorkspace(r.Context(), r.PathValue("ws"))
	if err != nil {
		storeError(w, err)
		return
	}
	u := signedIn(r).user
	expires, err := s.store.OpenPass(r.Context(), ws.ID, req.Pass, u.ID)
	if err != nil {
		storeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"pass":          req.Pass,
		"expires_at":    expires,
		"invite_prompt": invitePrompt(requestOrigin(r), ws, firstNonEmpty(u.DisplayName, "a member"), req.Pass, who),
	})
}

// invitePrompt is what a member pastes into a coding agent: get the CLI,
// join with the pass, then follow the agent guide, which has the rest (the
// laps, pull requests, parking on `unison wait`). Agents can't join
// without a pass, so there is no passless prompt.
//
// The join line carries what the member pre-assigned (who); the agent is
// asked only for what was left blank.
func invitePrompt(origin string, ws *store.Workspace, sponsor, pass string, who store.AgentJoin) string {
	url := origin + "/w/" + ws.ID
	// The doorway is private; the pass opens it (reading doesn't use it up).
	guide := url + "/agent.md?pass=" + pass
	var blanks []string
	flag := func(value, placeholder, ask string) string {
		if value != "" {
			return shellQuote(value)
		}
		blanks = append(blanks, ask)
		return placeholder
	}
	join := fmt.Sprintf("unison join %s --pass %s --name %s --emoji %s --model %s", url, pass,
		flag(who.Name, "<YourName>", "<YourName> with a name of your own"),
		flag(who.Emoji, "<emoji>", "<emoji> with an emoji of your own"),
		flag(who.Model, "<model>", "<model> with the model you run on"))
	fill := ""
	if len(blanks) > 0 {
		fill = "Replace " + joinList(blanks) + ". "
	}
	return fmt.Sprintf("You're invited to collaborate in the Unison workspace %q as an agent of %s. "+
		"Install the unison CLI (instructions: curl -s '%s'), then join with this pass "+
		"(it works once and expires in %d minutes):\n\n%s\n\n"+
		"%sThen run `unison --agent-help` and follow it.",
		ws.Name, sponsor, guide, int(store.AgentPassTTL/time.Minute), join, fill)
}

// shellSafe matches a value that needs no quoting in a shell command.
var shellSafe = regexp.MustCompile(`^[A-Za-z0-9._:/+=-]+$`)

// shellQuote writes v as one shell word. Pre-assigned values hold no quotes
// (store.CleanPrefill), so single quotes around anything unusual, such as
// an emoji or a model with a space in it, are enough.
func shellQuote(v string) string {
	if shellSafe.MatchString(v) {
		return v
	}
	return "'" + v + "'"
}

// joinList writes "a", "a and b" or "a, b and c".
func joinList(parts []string) string {
	if len(parts) < 2 {
		return strings.Join(parts, "")
	}
	return strings.Join(parts[:len(parts)-1], ", ") + " and " + parts[len(parts)-1]
}
