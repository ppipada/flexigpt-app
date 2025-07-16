package filenameprovider

import (
	"testing"
	"time"
)

func TestUUIDv7Provider_Build(t *testing.T) {
	p := &UUIDv7Provider{}

	longTitle := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec. Ipsto facto."
	sanitisedLong := "Lorem_ipsum_dolor_sit_amet__consectetur_adipiscing_elit__Donec__Ipsto_facto"

	tests := []struct {
		name        string
		id          string
		title       string
		want        string
		expectError bool
	}{
		{
			"simple",
			"018f1e3e-7c89-7b4b-8a3b-6f8e8f8e8f8e",
			"Chat",
			"018f1e3e-7c89-7b4b-8a3b-6f8e8f8e8f8e_Chat.json",
			false,
		},
		{
			"special-chars",
			"018f1e3e-7c89-7b4b-8a3b-6f8e8f8e8f8e",
			"Chat with AI!",
			"018f1e3e-7c89-7b4b-8a3b-6f8e8f8e8f8e_Chat_with_AI_.json",
			false,
		},
		{
			"empty title -> default",
			"018f1e3e-7c89-7b4b-8a3b-6f8e8f8e8f8e",
			"",
			"018f1e3e-7c89-7b4b-8a3b-6f8e8f8e8f8e_New_Conversation.json",
			false,
		},
		{
			"long title truncated to 64",
			"018f1e3e-7c89-7b4b-8a3b-6f8e8f8e8f8e",
			longTitle,
			"018f1e3e-7c89-7b4b-8a3b-6f8e8f8e8f8e_" + sanitisedLong[:64] + ".json",
			false,
		},
		{
			"missing id",
			"",
			"Chat",
			"",
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := p.Build(FileInfo{ID: tc.id, Title: tc.title})
			if (err != nil) != tc.expectError {
				t.Fatalf("want error %v, got %v", tc.expectError, err)
			}
			if err == nil && got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestUUIDv7Provider_Parse(t *testing.T) {
	p := &UUIDv7Provider{}

	validID := "018f1e3e-7c89-7b4b-8a3b-6f8e8f8e8f8e"
	validFile := validID + "_Chat_with_AI_.json"

	tests := []struct {
		name      string
		filename  string
		expectErr bool
		wantID    string
		wantTitle string
		// Assert that CreatedAt != zero.
		checkTime bool
	}{
		{
			"valid",
			validFile,
			false,
			validID,
			"Chat with AI ",
			true,
		},
		{
			"invalid filename no underscore",
			"whatever.json",
			true,
			"",
			"",
			false,
		},
		{
			"invalid uuid",
			"not-a-uuid_Chat.json",
			true,
			"",
			"",
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info, err := p.Parse(tc.filename)
			if (err != nil) != tc.expectErr {
				t.Fatalf("want err=%v got %v", tc.expectErr, err)
			}
			if err != nil {
				return
			}
			if info.ID != tc.wantID {
				t.Fatalf("want id %q got %q", tc.wantID, info.ID)
			}
			if info.Title != tc.wantTitle {
				t.Fatalf("want title %q got %q", tc.wantTitle, info.Title)
			}
			if tc.checkTime && info.CreatedAt.IsZero() {
				t.Fatalf("CreatedAt should not be zero")
			}
		})
	}
}

func TestUUIDv7Provider_BuildParse_RoundTrip(t *testing.T) {
	p := &UUIDv7Provider{}
	id := "018f1e3e-7c89-7b4b-8a3b-6f8e8f8e8f8e"
	title := "Round Trip Chat!"
	expectedTitle := "Round Trip Chat "
	file, err := p.Build(FileInfo{ID: id, Title: title})
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	info, err := p.Parse(file)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if info.ID != id || info.Title != expectedTitle {
		t.Fatalf("round-trip failed: got file:%s %+v", file, info)
	}
}

func TestUUIDv7Provider_CreatedAt(t *testing.T) {
	p := &UUIDv7Provider{}
	file := "018f1e3e-7c89-7b4b-8a3b-6f8e8f8e8f8e_MyChat.json"

	created, err := p.CreatedAt(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.IsZero() || created.After(time.Now()) {
		t.Fatalf("got suspicious time %v", created)
	}
}
