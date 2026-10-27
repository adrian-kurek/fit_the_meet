package group_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/adrian-kurek/fit_the_meet/internal/group"
)

func TestInfoJSONContract(t *testing.T) {
	t.Parallel()

	info := group.Info{
		ID:           8,
		Name:         "Project group",
		Role:         "owner",
		MembersCount: 2,
		CreatedAt:    time.Date(2026, 10, 27, 20, 0, 0, 0, time.UTC),
		Participants: []group.Participant{
			{ID: 12, Username: "adriank", Email: "adrian@example.com", Role: "owner"},
			{ID: 15, Username: "maya", Role: "member"},
		},
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var response map[string]any
	if err = json.Unmarshal(data, &response); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if response["member_count"] != float64(2) {
		t.Fatalf("member_count = %v, want 2", response["member_count"])
	}
	if _, exists := response["MembersCount"]; exists {
		t.Fatal("unexpected Go field name MembersCount in JSON response")
	}

	participants, ok := response["participants"].([]any)
	if !ok || len(participants) != 2 {
		t.Fatalf("participants = %v, want two participants", response["participants"])
	}
	member, ok := participants[1].(map[string]any)
	if !ok {
		t.Fatalf("member = %v, want JSON object", participants[1])
	}
	if _, exists := member["email"]; exists {
		t.Fatal("empty member email must be omitted")
	}
}
