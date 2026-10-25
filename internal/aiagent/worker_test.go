package aiagent

import (
	"reflect"
	"testing"

	"github.com/abhinavxd/libredesk/internal/ai"
)

func TestSplitSuggestions(t *testing.T) {
	body, suggestions := splitSuggestions("Which plan?\n[[suggestions]] [\"Basic\",\"Pro\",\"Enterprise\",\"Extra\"]")
	if body != "Which plan?" {
		t.Fatalf("unexpected body: %q", body)
	}
	want := []string{"Basic", "Pro", "Enterprise"}
	if !reflect.DeepEqual(suggestions, want) {
		t.Fatalf("unexpected suggestions: %#v", suggestions)
	}
}

func TestSplitSuggestionsRejectsInvalidPayload(t *testing.T) {
	body, suggestions := splitSuggestions("Tell me more\n[[suggestions]] not-json")
	if body != "Tell me more" || suggestions != nil {
		t.Fatalf("unexpected result: %q %#v", body, suggestions)
	}
}

func TestWithinWrapUp(t *testing.T) {
	for _, tc := range []struct {
		turns, max int
		want       bool
	}{{12, 12, true}, {15, 12, true}, {16, 12, false}, {20, 12, false}} {
		if got := withinWrapUp(tc.turns, tc.max); got != tc.want {
			t.Fatalf("withinWrapUp(%d, %d) = %v, want %v", tc.turns, tc.max, got, tc.want)
		}
	}
}

func TestWrapUpToolsKeepsVerificationAndAddsHandoff(t *testing.T) {
	tools := []ai.Tool{
		&searchKnowledgeTool{},
		&resolveTool{},
		&sendEmailVerificationTool{},
		&checkEmailVerificationTool{},
		&setContactEmailTool{},
	}
	got := wrapUpTools(tools, &handoffTool{})
	var names []string
	for _, tool := range got {
		names = append(names, tool.Name())
	}
	want := []string{"send_email_verification", "check_email_verification", "set_contact_email", "hand_off_to_human"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("unexpected tools: %v", names)
	}
}

func TestWrapUpToolsKeepsExistingHandoffOnce(t *testing.T) {
	got := wrapUpTools([]ai.Tool{&handoffTool{}, &sendEmailVerificationTool{}}, &handoffTool{})
	if len(got) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(got))
	}
}
