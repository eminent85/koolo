package discordv2

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"

	"github.com/hectorgimenez/koolo/internal/event"
)

const screenshotFileName = "Screenshot.jpeg"
const jpegQuality = 80

// Handle satisfies the event.Handler signature. It receives internal events and
// publishes them to Discord when the corresponding config toggle is enabled.
func (b *Bot) Handle(ctx context.Context, e event.Event) error {
	if !b.shouldPublish(e) {
		return nil
	}

	switch evt := e.(type) {
	case event.GameCreatedEvent:
		msg := fmt.Sprintf("**[%s]** %s\nGame: %s\nPassword: %s", evt.Supervisor(), evt.Message(), evt.Name, evt.Password)
		return b.sender.Send(ctx, msg, "", nil)

	case event.GameFinishedEvent:
		msg := fmt.Sprintf("**[%s]** %s", evt.Supervisor(), evt.Message())
		return b.sender.Send(ctx, msg, "", nil)

	case event.RunStartedEvent:
		msg := fmt.Sprintf("**[%s]** started a new run: **%s**", evt.Supervisor(), evt.RunName)
		return b.sender.Send(ctx, msg, "", nil)

	case event.RunFinishedEvent:
		msg := fmt.Sprintf("**[%s]** finished run: **%s** (%s)", evt.Supervisor(), evt.RunName, evt.Reason)
		return b.sender.Send(ctx, msg, "", nil)

	case event.NgrokTunnelEvent:
		return b.sender.Send(ctx, evt.Message(), "", nil)

	case event.ItemStashedEvent:
		return b.handleItemStashed(ctx, evt)
	}

	// Default: if the event carries an image, send it as a screenshot.
	return b.sendScreenshot(ctx, e, b.sender)
}

// handleItemStashed sends an item stash event as either an embed or a
// screenshot, depending on configuration.
func (b *Bot) handleItemStashed(ctx context.Context, evt event.ItemStashedEvent) error {
	if b.opts.DisableItemStashScreenshots {
		embed := buildItemStashEmbed(evt, b.opts.IncludePickitInfoInItemText)
		return b.getItemSender().SendEmbed(ctx, embed)
	}

	return b.sendScreenshot(ctx, evt, b.getItemSender())
}

// sendScreenshot encodes the event image as JPEG and sends it through the
// given sender. Returns nil if the event has no image.
func (b *Bot) sendScreenshot(ctx context.Context, e event.Event, sender MessageSender) error {
	if e.Image() == nil {
		return nil
	}

	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, e.Image(), &jpeg.Options{Quality: jpegQuality}); err != nil {
		return err
	}

	msg := fmt.Sprintf("**[%s]** %s", e.Supervisor(), e.Message())
	return sender.Send(ctx, msg, screenshotFileName, buf.Bytes())
}

// shouldPublish determines whether an event should be sent to Discord based on
// the bot's configuration toggles.
func (b *Bot) shouldPublish(e event.Event) bool {
	switch evt := e.(type) {
	case event.GameFinishedEvent:
		if evt.Reason == event.FinishedError {
			return b.opts.EnableDiscordErrorMessages
		}
		if evt.Reason == event.FinishedChicken || evt.Reason == event.FinishedMercChicken || evt.Reason == event.FinishedDied {
			return b.opts.EnableDiscordChickenMessages
		}
		if evt.Reason == event.FinishedOK {
			return false
		}
		return true
	case event.GameCreatedEvent:
		_ = evt
		return b.opts.EnableGameCreatedMessages
	case event.RunStartedEvent:
		_ = evt
		return b.opts.EnableNewRunMessages
	case event.RunFinishedEvent:
		_ = evt
		return b.opts.EnableRunFinishMessages
	case event.NgrokTunnelEvent:
		_ = evt
		return true
	default:
		return e.Image() != nil
	}
}
