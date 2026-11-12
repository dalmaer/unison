package store

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// What a member pre-assigns to an invited agent is tidied and checked, and
// a pass can be looked up while it is still unused (#306).
func TestCleanPrefillAndOpenPass(t *testing.T) {
	got, err := CleanPrefill(AgentJoin{Name: " Quill ", Emoji: " 🪶 ", Model: " Fable 5.2 "})
	if err != nil || got != (AgentJoin{Name: "Quill", Emoji: "🪶", Model: "Fable 5.2"}) {
		t.Errorf("CleanPrefill = %+v, %v", got, err)
	}
	if got, err := CleanPrefill(AgentJoin{}); err != nil || got != (AgentJoin{}) {
		t.Errorf("nothing pre-assigned = %+v, %v", got, err)
	}
	for _, bad := range []AgentJoin{
		{Name: "two words"}, {Name: "a@b"}, {Name: "O'Brien"}, {Name: `a"b`}, {Name: "$(id)"}, {Name: "a;b"}, {Name: "a|b"}, {Name: `a\b`},
		{Emoji: "🦊 🦊"}, {Emoji: "`x`"}, {Emoji: "a&b"},
		{Model: "a\nb"}, {Model: "x>y"}, {Model: "x*"}, {Model: strings.Repeat("m", ModelMax+1)},
	} {
		if _, err := CleanPrefill(bad); !errors.Is(err, ErrInvalid) {
			t.Errorf("CleanPrefill(%+v) = %v, want ErrInvalid", bad, err)
		}
	}

	s, now, ws, alex, bob := passSetup(t)
	ctx := context.Background()
	if _, _, _, _, err := s.JoinAsUser(ctx, ws.ID, bob.ID, "Bob", "🐻"); err != nil {
		t.Fatal(err)
	}
	pass, expires, err := s.MintPass(ctx, ws.ID, alex.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := s.OpenPass(ctx, ws.ID, pass, alex.ID); err != nil || got != expires {
		t.Errorf("OpenPass = %d, %v; want %d", got, err, expires)
	}
	// Somebody else's pass, another workspace, or no such pass: unknown.
	other, _, _ := s.CreateWorkspaceFor(ctx, alex.ID, "Other", "", "")
	for what, try := range map[string]func() (int64, error){
		"another member":    func() (int64, error) { return s.OpenPass(ctx, ws.ID, pass, bob.ID) },
		"another workspace": func() (int64, error) { return s.OpenPass(ctx, other.ID, pass, alex.ID) },
		"no such pass":      func() (int64, error) { return s.OpenPass(ctx, ws.ID, "up_nope", alex.ID) },
	} {
		if _, err := try(); !errors.Is(err, ErrForbidden) || !strings.Contains(err.Error(), "unknown pass") {
			t.Errorf("OpenPass (%s) = %v, want unknown pass", what, err)
		}
	}
	// Looking doesn't spend it; using it does; time runs out on the next one.
	if _, _, _, err := s.JoinWithPass(ctx, ws.ID, pass, "Quill", "🪶"); err != nil {
		t.Fatal(err)
	}
	wantForbidden(t, "a used pass", second(s.OpenPass(ctx, ws.ID, pass, alex.ID)), "already used")
	late, _, err := s.MintPass(ctx, ws.ID, alex.ID)
	if err != nil {
		t.Fatal(err)
	}
	*now = now.Add(AgentPassTTL)
	wantForbidden(t, "an expired pass", second(s.OpenPass(ctx, ws.ID, late, alex.ID)), "expired")
}

func second(_ int64, err error) error { return err }
