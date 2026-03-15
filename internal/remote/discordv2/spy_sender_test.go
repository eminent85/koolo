package discordv2

import (
	"context"
	"io"
	"log/slog"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

// spySender is a test double for MessageSender that records calls.
type spySender struct {
	mu    sync.Mutex
	calls []senderCall
}

type senderCall struct {
	kind     string // "send" or "embed"
	content  string
	fileName string
	fileData []byte
	embed    *discordgo.MessageEmbed
}

func (s *spySender) Send(_ context.Context, content, fileName string, fileData []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, senderCall{
		kind:     "send",
		content:  content,
		fileName: fileName,
		fileData: fileData,
	})
	return nil
}

func (s *spySender) SendEmbed(_ context.Context, embed *discordgo.MessageEmbed) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, senderCall{
		kind:  "embed",
		embed: embed,
	})
	return nil
}

func (s *spySender) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

func (s *spySender) lastCall() senderCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.calls) == 0 {
		return senderCall{}
	}
	return s.calls[len(s.calls)-1]
}

// newTestBot constructs a Bot with spy senders for testing Handle and event
// routing without a real Discord session.
func newTestBot(opts Options) (*Bot, *spySender, *spySender) {
	main := &spySender{}
	item := &spySender{}
	return &Bot{
		opts:       opts,
		sender:     main,
		itemSender: item,
		logger:     discardLogger,
	}, main, item
}

// newTestBotNoItemSender constructs a Bot with only a main sender.
// getItemSender will fall back to the main sender.
func newTestBotNoItemSender(opts Options) (*Bot, *spySender) {
	main := &spySender{}
	return &Bot{
		opts:   opts,
		sender: main,
		logger: discardLogger,
	}, main
}
