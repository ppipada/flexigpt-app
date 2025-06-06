package conversationstore

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ppipada/flexigpt-app/pkg/conversationstore/spec"
)

func newCollection(t *testing.T, dir string, withFTS bool) *ConversationCollection {
	t.Helper()
	cc, err := NewConversationCollection(
		dir,
		WithFTS(withFTS),
	)
	if err != nil {
		t.Fatalf("NewConversationCollection: %v", err)
	}
	return cc
}

func newConv(t *testing.T, title string) *spec.Conversation {
	t.Helper()
	id, _ := uuid.NewV7()
	now := time.Now()
	return &spec.Conversation{
		ConversationItem: spec.ConversationItem{ID: id.String(), Title: title, CreatedAt: now},
		ModifiedAt:       now,
		Messages:         []spec.ConversationMessage{},
	}
}

/* ------------------------------------------------------------------ */
/*  1. Search happy-path                                              */
/* ------------------------------------------------------------------ */

func TestFTSSearchHappyPath(t *testing.T) {
	dir := t.TempDir()
	cc := newCollection(t, dir, true)

	c1 := newConv(t, "Banana split") // query in title
	_, _ = cc.SaveConversation(t.Context(), &spec.SaveConversationRequest{Body: c1})

	resp, err := cc.SearchConversations(t.Context(), &spec.SearchConversationsRequest{
		Query: "banana",
		Token: "",
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(resp.Body.ConversationItems) != 1 {
		t.Fatalf("want 1 hit, got %d", len(resp.Body.ConversationItems))
	}
	if resp.Body.ConversationItems[0].ID != c1.ID {
		t.Fatalf("unexpected hit %q", resp.Body.ConversationItems[0].ID)
	}
}

/* ------------------------------------------------------------------ */
/*  2. Ranking: title beats body                                      */
/* ------------------------------------------------------------------ */

func TestFTSRankingTitleVsBody(t *testing.T) {
	dir := t.TempDir()
	cc := newCollection(t, dir, true)

	// A: query term in title
	a := newConv(t, "Alpha winner")
	_, _ = cc.SaveConversation(t.Context(), &spec.SaveConversationRequest{Body: a})

	// B: query term only inside a message
	b := newConv(t, "No match in title")
	b.Messages = []spec.ConversationMessage{{
		ID:      "m1",
		Role:    spec.ConversationRoleUser,
		Content: "alpha appears in body",
	}}
	_, _ = cc.SaveConversation(t.Context(), &spec.SaveConversationRequest{Body: b})

	resp, _ := cc.SearchConversations(t.Context(),
		&spec.SearchConversationsRequest{Query: "alpha"})
	if len(resp.Body.ConversationItems) != 2 ||
		resp.Body.ConversationItems[0].ID != a.ID {
		t.Fatalf("ranking wrong, hits=%v", resp.Body.ConversationItems)
	}
}

/* ------------------------------------------------------------------ */
/*  3. Pagination token                                               */
/* ------------------------------------------------------------------ */

func TestFTSPagination(t *testing.T) {
	dir := t.TempDir()
	cc := newCollection(t, dir, true)

	for i := range 12 {
		c := newConv(t, "Kiwi talk "+strconv.Itoa(i))
		c.Messages = []spec.ConversationMessage{{
			ID:      "m",
			Role:    spec.ConversationRoleUser,
			Content: "kiwi everywhere",
		}}
		_, _ = cc.SaveConversation(t.Context(), &spec.SaveConversationRequest{Body: c})
	}

	var (
		token string
		total int
	)
	for {
		res, err := cc.SearchConversations(t.Context(),
			&spec.SearchConversationsRequest{Query: "kiwi", Token: token})
		if err != nil {
			t.Fatalf("search page: %v", err)
		}
		total += len(res.Body.ConversationItems)
		if res.Body.NextPageToken == nil || *res.Body.NextPageToken == "" {
			if len(res.Body.ConversationItems) != 2 { // 12−10
				t.Fatalf("last page size want 2, got %d", len(res.Body.ConversationItems))
			}
			break
		}
		if len(res.Body.ConversationItems) != 10 {
			t.Fatalf("full page size want 10 got %d", len(res.Body.ConversationItems))
		}
		token = *res.Body.NextPageToken
	}
	if total != 12 {
		t.Fatalf("expected 12 hits over all pages, got %d", total)
	}
}

/* ------------------------------------------------------------------ */
/*  4. Delete removes from index                                      */
/* ------------------------------------------------------------------ */

func TestFTSDeletePurgesIndex(t *testing.T) {
	dir := t.TempDir()
	cc := newCollection(t, dir, true)

	c := newConv(t, "Cherry pie")
	_, _ = cc.SaveConversation(t.Context(), &spec.SaveConversationRequest{Body: c})

	// ensure hit exists
	_, err := cc.SearchConversations(t.Context(),
		&spec.SearchConversationsRequest{Query: "cherry"})
	if err != nil {
		t.Fatalf("initial search: %v", err)
	}

	// delete conversation
	_, _ = cc.DeleteConversation(t.Context(),
		&spec.DeleteConversationRequest{ID: c.ID, Title: c.Title})

	// expect zero hits
	res, _ := cc.SearchConversations(t.Context(),
		&spec.SearchConversationsRequest{Query: "cherry"})
	if len(res.Body.ConversationItems) != 0 {
		t.Fatalf("expected 0 hits after delete, got %d", len(res.Body.ConversationItems))
	}
}

/* ------------------------------------------------------------------ */
/*  5. AddMessage updates index                                       */
/* ------------------------------------------------------------------ */

func TestFTSAddMessageUpdatesIndex(t *testing.T) {
	dir := t.TempDir()
	cc := newCollection(t, dir, true)

	c := newConv(t, "Silent")
	_, _ = cc.SaveConversation(t.Context(), &spec.SaveConversationRequest{Body: c})

	// no hit yet
	res, _ := cc.SearchConversations(t.Context(),
		&spec.SearchConversationsRequest{Query: "whisper"})
	if len(res.Body.ConversationItems) != 0 {
		t.Fatal("should have no hits before message added")
	}

	msg := spec.ConversationMessage{
		ID:      "m1",
		Role:    spec.ConversationRoleAssistant,
		Content: "let me whisper a secret",
	}
	_, _ = cc.AddMessageToConversation(t.Context(),
		&spec.AddMessageToConversationRequest{
			ID: c.ID,
			Body: &spec.AddMessageToConversationRequestBody{
				Title:      c.Title,
				NewMessage: msg,
			},
		})

	// now should hit
	res, _ = cc.SearchConversations(t.Context(),
		&spec.SearchConversationsRequest{Query: "whisper"})
	if len(res.Body.ConversationItems) != 1 {
		t.Fatalf("want 1 hit after adding message, got %d", len(res.Body.ConversationItems))
	}
}

/* ------------------------------------------------------------------ */
/*  6. Search disabled when FTS off                                   */
/* ------------------------------------------------------------------ */

func TestFTSDisabled(t *testing.T) {
	dir := t.TempDir()
	cc := newCollection(t, dir, false)

	_, err := cc.SearchConversations(t.Context(),
		&spec.SearchConversationsRequest{Query: "x"})
	if err == nil {
		t.Fatal("expected error when FTS disabled")
	}
}

/* ------------------------------------------------------------------ */
/*  7. Scale: 100 conversations, 20 messages each                     */
/* ------------------------------------------------------------------ */

func TestFTSScaleSearch(t *testing.T) {
	dir := t.TempDir()
	cc := newCollection(t, dir, true)

	const (
		nConvos   = 100
		nMessages = 20
		pageSz    = 15
	)

	/* ---- create & save conversations ---------------------------------- */
	for i := range nConvos {
		c := newConv(t, fmt.Sprintf("Scale test fruit%d", i))
		for m := range nMessages {
			role := spec.ConversationRoleUser
			if m%2 == 1 {
				role = spec.ConversationRoleAssistant
			}
			c.Messages = append(c.Messages, spec.ConversationMessage{
				ID:      fmt.Sprintf("m%02d", m),
				Role:    role,
				Content: fmt.Sprintf("This is message %d with common keyword", m),
			})
		}
		if _, err := cc.SaveConversation(t.Context(),
			&spec.SaveConversationRequest{Body: c}); err != nil {
			t.Fatalf("save %d: %v", i, err)
		}
	}

	/* ---- single-doc search ------------------------------------------- */
	res, err := cc.SearchConversations(t.Context(),
		&spec.SearchConversationsRequest{Query: "fruit42", PageSize: pageSz})
	if err != nil {
		t.Fatalf("search fruit42: %v", err)
	}
	if len(res.Body.ConversationItems) != 1 {
		t.Fatalf("expect 1 hit for fruit42, got %d", len(res.Body.ConversationItems))
	}

	/* ---- bulk search with pagination --------------------------------- */
	token := ""
	seen := map[string]bool{}
	total := 0
	for page := 0; ; page++ {
		r, err := cc.SearchConversations(t.Context(),
			&spec.SearchConversationsRequest{
				Query:    "common",
				Token:    token,
				PageSize: pageSz,
			})
		if err != nil {
			t.Fatalf("page %d search: %v", page, err)
		}
		for _, it := range r.Body.ConversationItems {
			if seen[it.ID] {
				t.Fatalf("duplicate id %s on page %d", it.ID, page)
			}
			seen[it.ID] = true
		}
		total += len(r.Body.ConversationItems)

		if r.Body.NextPageToken == nil || *r.Body.NextPageToken == "" {
			break
		}
		if len(r.Body.ConversationItems) != pageSz {
			t.Fatalf("full page expected %d items, got %d", pageSz,
				len(r.Body.ConversationItems))
		}
		token = *r.Body.NextPageToken
	}

	if total != nConvos {
		t.Fatalf("expected %d total hits for 'common', got %d", nConvos, total)
	}
}
