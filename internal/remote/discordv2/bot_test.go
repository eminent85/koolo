package discordv2

import (
	"testing"
)

func TestNew_WebhookMode_RequiresURL(t *testing.T) {
	_, err := New(Options{UseWebhook: true, WebhookURL: ""}, nil, discardLogger)
	if err == nil {
		t.Fatal("expected error when webhook URL is empty")
	}
}

func TestNew_WebhookMode_OK(t *testing.T) {
	b, err := New(Options{UseWebhook: true, WebhookURL: "https://example.com/webhook"}, nil, discardLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.sender == nil {
		t.Fatal("sender should be set")
	}
	if b.itemSender != nil {
		t.Fatal("itemSender should be nil when no item webhook URL is set")
	}
}

func TestNew_WebhookMode_WithItemWebhook(t *testing.T) {
	b, err := New(Options{
		UseWebhook:     true,
		WebhookURL:     "https://example.com/webhook",
		ItemWebhookURL: "https://example.com/items",
	}, nil, discardLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.itemSender == nil {
		t.Fatal("itemSender should be set when item webhook URL is provided")
	}
}

func TestNew_BotTokenMode_OK(t *testing.T) {
	// discordgo.New succeeds with any non-empty token string
	b, err := New(Options{Token: "test-token"}, nil, discardLogger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.session == nil {
		t.Fatal("session should be set in bot-token mode")
	}
	if b.sender == nil {
		t.Fatal("sender should be set")
	}
}

func TestGetItemSender_FallsBackToSender(t *testing.T) {
	main := &spySender{}
	b := &Bot{sender: main}

	if got := b.getItemSender(); got != main {
		t.Fatal("getItemSender should return main sender when itemSender is nil")
	}
}

func TestGetItemSender_ReturnsItemSender(t *testing.T) {
	main := &spySender{}
	item := &spySender{}
	b := &Bot{sender: main, itemSender: item}

	if got := b.getItemSender(); got != item {
		t.Fatal("getItemSender should return itemSender when set")
	}
}
