package discordv2

import (
	"context"
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/hectorgimenez/koolo/internal/event"
)

// newTestImage creates a minimal 1x1 image for testing screenshot paths.
func newTestImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	return img
}

// ---------------------------------------------------------------------------
// shouldPublish
// ---------------------------------------------------------------------------

func TestShouldPublish_GameFinished(t *testing.T) {
	tests := []struct {
		name          string
		reason        event.FinishReason
		enableError   bool
		enableChicken bool
		want          bool
	}{
		{"error enabled", event.FinishedError, true, false, true},
		{"error disabled", event.FinishedError, false, false, false},
		{"chicken enabled", event.FinishedChicken, false, true, true},
		{"chicken disabled", event.FinishedChicken, false, false, false},
		{"merc chicken enabled", event.FinishedMercChicken, false, true, true},
		{"merc chicken disabled", event.FinishedMercChicken, false, false, false},
		{"died enabled", event.FinishedDied, false, true, true},
		{"died disabled", event.FinishedDied, false, false, false},
		{"ok always suppressed", event.FinishedOK, true, true, false},
		{"unknown reason always published", event.FinishReason("unknown"), false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bot{opts: Options{
				EnableDiscordErrorMessages:   tt.enableError,
				EnableDiscordChickenMessages: tt.enableChicken,
			}}

			evt := event.GameFinished(event.Text("test", "game finished"), tt.reason)
			if got := b.shouldPublish(evt); got != tt.want {
				t.Errorf("shouldPublish() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldPublish_GameCreated(t *testing.T) {
	tests := []struct {
		enabled bool
		want    bool
	}{
		{true, true},
		{false, false},
	}
	for _, tt := range tests {
		b := &Bot{opts: Options{EnableGameCreatedMessages: tt.enabled}}
		evt := event.GameCreated(event.Text("test", "created"), "game1", "pass1")
		if got := b.shouldPublish(evt); got != tt.want {
			t.Errorf("EnableGameCreatedMessages=%v: shouldPublish() = %v, want %v", tt.enabled, got, tt.want)
		}
	}
}

func TestShouldPublish_RunStarted(t *testing.T) {
	tests := []struct {
		enabled bool
		want    bool
	}{
		{true, true},
		{false, false},
	}
	for _, tt := range tests {
		b := &Bot{opts: Options{EnableNewRunMessages: tt.enabled}}
		evt := event.RunStarted(event.Text("test", "run started"), "Mephisto")
		if got := b.shouldPublish(evt); got != tt.want {
			t.Errorf("EnableNewRunMessages=%v: shouldPublish() = %v, want %v", tt.enabled, got, tt.want)
		}
	}
}

func TestShouldPublish_RunFinished(t *testing.T) {
	tests := []struct {
		enabled bool
		want    bool
	}{
		{true, true},
		{false, false},
	}
	for _, tt := range tests {
		b := &Bot{opts: Options{EnableRunFinishMessages: tt.enabled}}
		evt := event.RunFinished(event.Text("test", "run finished"), "Mephisto", event.FinishedOK)
		if got := b.shouldPublish(evt); got != tt.want {
			t.Errorf("EnableRunFinishMessages=%v: shouldPublish() = %v, want %v", tt.enabled, got, tt.want)
		}
	}
}

func TestShouldPublish_NgrokTunnel_AlwaysTrue(t *testing.T) {
	b := &Bot{opts: Options{}}
	evt := event.NgrokTunnel("https://example.ngrok.io")
	if !b.shouldPublish(evt) {
		t.Error("NgrokTunnelEvent should always be published")
	}
}

func TestShouldPublish_DefaultEvent_NoImage(t *testing.T) {
	b := &Bot{opts: Options{}}
	evt := event.Text("test", "some text")
	if b.shouldPublish(evt) {
		t.Error("default event without image should not be published")
	}
}

func TestShouldPublish_DefaultEvent_WithImage(t *testing.T) {
	b := &Bot{opts: Options{}}
	evt := event.WithScreenshot("test", "screenshot", newTestImage())
	if !b.shouldPublish(evt) {
		t.Error("default event with image should be published")
	}
}

// ---------------------------------------------------------------------------
// Handle — message routing and formatting
// ---------------------------------------------------------------------------

func TestHandle_GameCreated(t *testing.T) {
	b, main, _ := newTestBot(Options{EnableGameCreatedMessages: true})
	evt := event.GameCreated(event.Text("Koza", "Game created"), "game1", "pass1")

	if err := b.Handle(context.Background(), evt); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if main.callCount() != 1 {
		t.Fatalf("expected 1 call, got %d", main.callCount())
	}
	call := main.lastCall()
	if call.kind != "send" {
		t.Fatalf("expected send call, got %s", call.kind)
	}
	if !strings.Contains(call.content, "Koza") {
		t.Errorf("content should contain supervisor name, got: %s", call.content)
	}
	if !strings.Contains(call.content, "game1") {
		t.Errorf("content should contain game name, got: %s", call.content)
	}
	if !strings.Contains(call.content, "pass1") {
		t.Errorf("content should contain password, got: %s", call.content)
	}
}

func TestHandle_GameFinished(t *testing.T) {
	b, main, _ := newTestBot(Options{EnableDiscordChickenMessages: true})
	evt := event.GameFinished(event.Text("Koza", "Game finished"), event.FinishedChicken)

	if err := b.Handle(context.Background(), evt); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if main.callCount() != 1 {
		t.Fatalf("expected 1 call, got %d", main.callCount())
	}
	call := main.lastCall()
	if !strings.Contains(call.content, "Koza") {
		t.Errorf("content should contain supervisor name, got: %s", call.content)
	}
}

func TestHandle_RunStarted(t *testing.T) {
	b, main, _ := newTestBot(Options{EnableNewRunMessages: true})
	evt := event.RunStarted(event.Text("Koza", "Run started"), "Mephisto")

	if err := b.Handle(context.Background(), evt); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if main.callCount() != 1 {
		t.Fatalf("expected 1 call, got %d", main.callCount())
	}
	call := main.lastCall()
	if !strings.Contains(call.content, "Mephisto") {
		t.Errorf("content should contain run name, got: %s", call.content)
	}
}

func TestHandle_RunFinished(t *testing.T) {
	b, main, _ := newTestBot(Options{EnableRunFinishMessages: true})
	evt := event.RunFinished(event.Text("Koza", "Run finished"), "Mephisto", event.FinishedOK)

	if err := b.Handle(context.Background(), evt); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if main.callCount() != 1 {
		t.Fatalf("expected 1 call, got %d", main.callCount())
	}
	call := main.lastCall()
	if !strings.Contains(call.content, "Mephisto") {
		t.Errorf("content should contain run name, got: %s", call.content)
	}
	if !strings.Contains(call.content, string(event.FinishedOK)) {
		t.Errorf("content should contain reason, got: %s", call.content)
	}
}

func TestHandle_NgrokTunnel(t *testing.T) {
	b, main, _ := newTestBot(Options{})
	evt := event.NgrokTunnel("https://example.ngrok.io")

	if err := b.Handle(context.Background(), evt); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if main.callCount() != 1 {
		t.Fatalf("expected 1 call, got %d", main.callCount())
	}
	call := main.lastCall()
	if !strings.Contains(call.content, "https://example.ngrok.io") {
		t.Errorf("content should contain ngrok URL, got: %s", call.content)
	}
}

func TestHandle_SuppressedEvent_NoCalls(t *testing.T) {
	// All toggles off, no image — nothing should be sent.
	b, main, _ := newTestBot(Options{})
	evt := event.GameCreated(event.Text("Koza", "created"), "g", "p")

	if err := b.Handle(context.Background(), evt); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if main.callCount() != 0 {
		t.Errorf("expected 0 calls for suppressed event, got %d", main.callCount())
	}
}

func TestHandle_DefaultEvent_WithImage_SendsScreenshot(t *testing.T) {
	b, main, _ := newTestBot(Options{})
	evt := event.WithScreenshot("Koza", "something happened", newTestImage())

	if err := b.Handle(context.Background(), evt); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if main.callCount() != 1 {
		t.Fatalf("expected 1 call, got %d", main.callCount())
	}
	call := main.lastCall()
	if call.fileName != screenshotFileName {
		t.Errorf("expected filename %q, got %q", screenshotFileName, call.fileName)
	}
	if len(call.fileData) == 0 {
		t.Error("expected non-empty JPEG file data")
	}
}

func TestHandle_DefaultEvent_NoImage_NoCalls(t *testing.T) {
	b, main, _ := newTestBot(Options{})
	// A plain BaseEvent with no image and no matching type switch case
	evt := event.Text("Koza", "nothing special")

	if err := b.Handle(context.Background(), evt); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if main.callCount() != 0 {
		t.Errorf("expected 0 calls, got %d", main.callCount())
	}
}

// ---------------------------------------------------------------------------
// Handle — item stash routing
// ---------------------------------------------------------------------------

func TestHandle_ItemStashed_Screenshot_UsesItemSender(t *testing.T) {
	b, main, itemSpy := newTestBot(Options{
		EnableGameCreatedMessages: true, // irrelevant but ensures shouldPublish doesn't block
	})
	// ItemStashedEvent with an image — default behavior is screenshot mode.
	evt := event.ItemStashed(
		event.WithScreenshot("Koza", "Item stashed", newTestImage()),
		makeTestDrop(),
	)

	if err := b.Handle(context.Background(), evt); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	// Should route through the item sender, not the main sender.
	if itemSpy.callCount() != 1 {
		t.Fatalf("expected 1 item sender call, got %d", itemSpy.callCount())
	}
	if main.callCount() != 0 {
		t.Errorf("expected 0 main sender calls, got %d", main.callCount())
	}
	call := itemSpy.lastCall()
	if call.fileName != screenshotFileName {
		t.Errorf("expected filename %q, got %q", screenshotFileName, call.fileName)
	}
}

func TestHandle_ItemStashed_Embed_UsesItemSender(t *testing.T) {
	b, main, itemSpy := newTestBot(Options{
		DisableItemStashScreenshots: true,
	})
	evt := event.ItemStashed(
		event.WithScreenshot("Koza", "Item stashed", newTestImage()),
		makeTestDrop(),
	)

	if err := b.Handle(context.Background(), evt); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if itemSpy.callCount() != 1 {
		t.Fatalf("expected 1 item sender call, got %d", itemSpy.callCount())
	}
	if main.callCount() != 0 {
		t.Errorf("expected 0 main sender calls, got %d", main.callCount())
	}
	call := itemSpy.lastCall()
	if call.kind != "embed" {
		t.Fatalf("expected embed call, got %s", call.kind)
	}
	if call.embed == nil {
		t.Fatal("embed should not be nil")
	}
}

func TestHandle_ItemStashed_NoItemSender_FallsBackToMain(t *testing.T) {
	b, main := newTestBotNoItemSender(Options{
		DisableItemStashScreenshots: true,
	})
	evt := event.ItemStashed(
		event.WithScreenshot("Koza", "Item stashed", newTestImage()),
		makeTestDrop(),
	)

	if err := b.Handle(context.Background(), evt); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if main.callCount() != 1 {
		t.Fatalf("expected 1 main sender call, got %d", main.callCount())
	}
}

func TestHandle_ItemStashed_NoImage_NoScreenshot(t *testing.T) {
	// Screenshot mode, but event has no image — should be a no-op.
	b, _, itemSpy := newTestBot(Options{})
	evt := event.ItemStashed(
		event.Text("Koza", "Item stashed"),
		makeTestDrop(),
	)

	if err := b.Handle(context.Background(), evt); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
	if itemSpy.callCount() != 0 {
		t.Errorf("expected 0 calls when no image, got %d", itemSpy.callCount())
	}
}
