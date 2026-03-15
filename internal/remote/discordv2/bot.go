package discordv2

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hectorgimenez/d2go/pkg/data"
)

// SupervisorStatus mirrors bot.SupervisorStatus without importing the bot
// package (which transitively depends on Windows-only packages).
type SupervisorStatus string

const (
	StatusNotStarted SupervisorStatus = "Not Started"
	StatusStarting   SupervisorStatus = "Starting"
	StatusInGame     SupervisorStatus = "In game"
	StatusPaused     SupervisorStatus = "Paused"
	StatusCrashed    SupervisorStatus = "Crashed"
)

// SupervisorStats holds the subset of supervisor statistics that the discord
// bot needs. The caller (application wiring) maps bot.Stats to this struct.
type SupervisorStats struct {
	SupervisorStatus SupervisorStatus
	StartedAt        time.Time
	Drops            []data.Drop
	TotalGames       int
	TotalDeaths      int
	TotalChickens    int
	TotalErrors      int
}

// SupervisorControl is the subset of bot.SupervisorManager that the discord
// bot needs. Using an interface breaks the transitive dependency on
// Windows-only packages, allowing the package to be tested on any OS.
type SupervisorControl interface {
	AvailableSupervisors() []string
	Start(supervisorName string, attachToExisting bool, manualMode bool, pidHwnd ...uint32) error
	Stop(supervisor string)
	Status(characterName string) SupervisorStats
	GetSupervisorStats(supervisor string) SupervisorStats
}

// Options holds all configuration needed to construct a Bot. Passing config
// explicitly keeps the implementation free of global state reads.
type Options struct {
	// Token is the Discord bot token. Required when UseWebhook is false.
	Token string
	// ChannelID is the main notification channel.
	ChannelID string
	// ItemChannelID is an optional separate channel for item drop messages.
	// When empty, item messages go to ChannelID.
	ItemChannelID string

	// UseWebhook enables webhook delivery mode. When true, the bot cannot
	// receive commands — it only publishes events.
	UseWebhook bool
	// WebhookURL is the main webhook URL. Required when UseWebhook is true.
	WebhookURL string
	// ItemWebhookURL is an optional webhook for item messages.
	ItemWebhookURL string

	// BotAdmins is the list of Discord user IDs allowed to run commands.
	BotAdmins []string

	// Event publishing toggles — mirror the v1 config surface.
	EnableGameCreatedMessages    bool
	EnableNewRunMessages         bool
	EnableRunFinishMessages      bool
	EnableDiscordErrorMessages   bool
	EnableDiscordChickenMessages bool
	DisableItemStashScreenshots  bool
	IncludePickitInfoInItemText  bool
}

// Bot is the v2 Discord integration. It implements the event.Handler signature
// and optionally listens for user commands when running in bot-token mode.
type Bot struct {
	session *discordgo.Session
	manager SupervisorControl
	opts    Options

	// sender is the default message sender (main channel / webhook).
	sender MessageSender
	// itemSender delivers item-specific messages. Falls back to sender when
	// no separate item channel/webhook is configured.
	itemSender MessageSender
}

// New creates a Bot from the given options and supervisor manager.
func New(opts Options, manager SupervisorControl) (*Bot, error) {
	b := &Bot{
		manager: manager,
		opts:    opts,
	}

	if opts.UseWebhook {
		if opts.WebhookURL == "" {
			return nil, fmt.Errorf("discordv2: webhook URL is required in webhook mode")
		}
		b.sender = newWebhookSender(opts.WebhookURL)
		if strings.TrimSpace(opts.ItemWebhookURL) != "" {
			b.itemSender = newWebhookSender(opts.ItemWebhookURL)
		}
		return b, nil
	}

	dg, err := discordgo.New("Bot " + opts.Token)
	if err != nil {
		return nil, fmt.Errorf("discordv2: create session: %w", err)
	}
	b.session = dg
	b.sender = newSessionSender(dg, opts.ChannelID)

	itemCh := strings.TrimSpace(opts.ItemChannelID)
	if itemCh != "" {
		b.itemSender = newSessionSender(dg, itemCh)
	}

	return b, nil
}

// Start opens the Discord gateway connection (bot-token mode) and blocks until
// ctx is cancelled. In webhook mode it simply blocks on context cancellation.
func (b *Bot) Start(ctx context.Context) error {
	if b.opts.UseWebhook {
		<-ctx.Done()
		return nil
	}

	b.session.AddHandler(b.onMessageCreated)
	b.session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentMessageContent

	if err := b.session.Open(); err != nil {
		return fmt.Errorf("discordv2: open connection: %w", err)
	}

	<-ctx.Done()
	return b.session.Close()
}

// getItemSender returns the item-specific sender, falling back to the default.
func (b *Bot) getItemSender() MessageSender {
	if b.itemSender != nil {
		return b.itemSender
	}
	return b.sender
}

// onMessageCreated routes incoming Discord messages to command handlers.
func (b *Bot) onMessageCreated(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if !slices.Contains(b.opts.BotAdmins, m.Author.ID) {
		return
	}
	if !strings.HasPrefix(m.Content, "!") {
		return
	}

	prefix := strings.Split(m.Content, " ")[0]
	switch prefix {
	case "!start":
		b.handleStart(s, m)
	case "!stop":
		b.handleStop(s, m)
	case "!status":
		b.handleStatus(s, m)
	case "!stats":
		b.handleStats(s, m)
	case "!list":
		b.handleList(s, m)
	case "!help":
		b.handleHelp(s, m)
	case "!drops":
		b.handleDrops(s, m)
	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unknown command: `%s`. Type `!help` for available commands.", prefix))
	}
}

// supervisorExists checks whether a supervisor name is configured.
func (b *Bot) supervisorExists(name string) bool {
	return slices.Contains(b.manager.AvailableSupervisors(), name)
}
