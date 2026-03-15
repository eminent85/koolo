package discordv2

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hectorgimenez/d2go/pkg/data"
)

func (b *Bot) handleStart(s *discordgo.Session, m *discordgo.MessageCreate) {
	words := strings.Fields(m.Content)

	if len(words) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !start <supervisor1> [supervisor2] ...")
		return
	}

	for _, supervisor := range words[1:] {
		if !b.supervisorExists(supervisor) {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' not found.", supervisor))
			continue
		}

		b.manager.Start(supervisor, false, false)
		time.Sleep(1 * time.Second)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' has been started.", supervisor))
	}
}

func (b *Bot) handleStop(s *discordgo.Session, m *discordgo.MessageCreate) {
	words := strings.Fields(m.Content)

	if len(words) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !stop <supervisor1> [supervisor2] ...")
		return
	}

	for _, supervisor := range words[1:] {
		if !b.supervisorExists(supervisor) {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' not found.", supervisor))
			continue
		}

		status := b.manager.Status(supervisor)
		if status.SupervisorStatus == StatusNotStarted || status.SupervisorStatus == "" {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' is not running.", supervisor))
			continue
		}

		b.manager.Stop(supervisor)
		time.Sleep(1 * time.Second)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' has been stopped.", supervisor))
	}
}

func (b *Bot) handleStatus(s *discordgo.Session, m *discordgo.MessageCreate) {
	words := strings.Fields(m.Content)

	if len(words) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !status <supervisor1> [supervisor2] ...")
		return
	}

	for _, supervisor := range words[1:] {
		if !b.supervisorExists(supervisor) {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' not found.", supervisor))
			continue
		}

		status := b.manager.Status(supervisor)
		if status.SupervisorStatus == StatusNotStarted || status.SupervisorStatus == "" {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' is offline.", supervisor))
			continue
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' is %s", supervisor, status.SupervisorStatus))
	}
}

func (b *Bot) handleStats(s *discordgo.Session, m *discordgo.MessageCreate) {
	words := strings.Fields(m.Content)

	if len(words) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !stats <supervisor1> [supervisor2] ...")
		return
	}

	for _, supervisor := range words[1:] {
		if !b.supervisorExists(supervisor) {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' not found.", supervisor))
			continue
		}

		embed := buildStatsEmbed(supervisor, b.manager.Status(supervisor), b.manager.GetSupervisorStats(supervisor))
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
	}
}

func (b *Bot) handleList(s *discordgo.Session, m *discordgo.MessageCreate) {
	supervisors := b.manager.AvailableSupervisors()

	if len(supervisors) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No supervisors available.")
		return
	}

	embed := buildListEmbed(supervisors, b.manager)
	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

func (b *Bot) handleHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSendEmbed(m.ChannelID, buildHelpEmbed())
}

func (b *Bot) handleDrops(s *discordgo.Session, m *discordgo.MessageCreate) {
	words := strings.Fields(m.Content)

	if len(words) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !drops <supervisor> [count]\nExample: `!drops Koza 10`")
		return
	}

	supervisor := words[1]

	if !b.supervisorExists(supervisor) {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' not found.", supervisor))
		return
	}

	count := 5
	if len(words) > 2 {
		fmt.Sscanf(words[2], "%d", &count)
		if count < 1 {
			count = 5
		}
		if count > 20 {
			count = 20
		}
	}

	stats := b.manager.GetSupervisorStats(supervisor)

	if len(stats.Drops) == 0 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No drops recorded for '%s' yet.", supervisor))
		return
	}

	embed := buildDropsEmbed(supervisor, stats.Drops, count)
	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

// ---------------------------------------------------------------------------
// Embed builders — extracted for testability
// ---------------------------------------------------------------------------

// buildStatsEmbed creates the embed for the !stats command.
func buildStatsEmbed(supervisor string, status, stats SupervisorStats) *discordgo.MessageEmbed {
	supStatus := string(status.SupervisorStatus)
	if supStatus == string(StatusNotStarted) || supStatus == "" {
		supStatus = "Offline"
	}

	return &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Stats for %s", supervisor),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Status", Value: supStatus, Inline: true},
			{Name: "Uptime", Value: time.Since(status.StartedAt).String(), Inline: true},
			{Name: "Games", Value: fmt.Sprintf("%d", stats.TotalGames), Inline: true},
			{Name: "Drops", Value: fmt.Sprintf("%d", len(stats.Drops)), Inline: true},
			{Name: "Deaths", Value: fmt.Sprintf("%d", stats.TotalDeaths), Inline: true},
			{Name: "Chickens", Value: fmt.Sprintf("%d", stats.TotalChickens), Inline: true},
			{Name: "Errors", Value: fmt.Sprintf("%d", stats.TotalErrors), Inline: true},
		},
	}
}

// buildListEmbed creates the embed for the !list command.
func buildListEmbed(supervisors []string, manager SupervisorControl) *discordgo.MessageEmbed {
	var fields []*discordgo.MessageEmbedField

	for _, supervisor := range supervisors {
		status := manager.Status(supervisor)
		var statusText, uptimeText string

		if status.SupervisorStatus == StatusNotStarted || status.SupervisorStatus == "" {
			statusText = "❌ Offline"
			uptimeText = "-"
		} else {
			statusText = fmt.Sprintf("✅ %s", status.SupervisorStatus)
			uptime := time.Since(status.StartedAt)
			if uptime < time.Minute {
				uptimeText = fmt.Sprintf("%ds", int(uptime.Seconds()))
			} else if uptime < time.Hour {
				uptimeText = fmt.Sprintf("%dm", int(uptime.Minutes()))
			} else {
				uptimeText = fmt.Sprintf("%dh %dm", int(uptime.Hours()), int(uptime.Minutes())%60)
			}
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   supervisor,
			Value:  fmt.Sprintf("Status: %s\nUptime: %s", statusText, uptimeText),
			Inline: true,
		})
	}

	return &discordgo.MessageEmbed{
		Title:  "📋 Available Supervisors",
		Fields: fields,
		Color:  0x5865F2,
	}
}

// buildHelpEmbed creates the embed for the !help command.
func buildHelpEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "🤖 Koolo Discord Bot Commands",
		Description: "Control and monitor your Diablo II bot supervisors",
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "!list", Value: "Show all available supervisors with their status and uptime", Inline: false},
			{Name: "!start <supervisor1> [supervisor2] ...", Value: "Start one or more supervisors\nExample: `!start Koza Ovca`", Inline: false},
			{Name: "!stop <supervisor1> [supervisor2] ...", Value: "Stop one or more supervisors\nExample: `!stop Koza`", Inline: false},
			{Name: "!status <supervisor1> [supervisor2] ...", Value: "Check the current status of supervisors\nExample: `!status Koza Ovca`", Inline: false},
			{Name: "!stats <supervisor1> [supervisor2] ...", Value: "Get detailed statistics for supervisors\nExample: `!stats Koza`", Inline: false},
			{Name: "!drops <supervisor> [count]", Value: "Show recent drops for a supervisor\nExample: `!drops Koza 10`\nDefault count: 5", Inline: false},
			{Name: "!help", Value: "Show this help message", Inline: false},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "💡 Tip: You can control multiple supervisors at once with most commands",
		},
	}
}

// dropQualityEmoji returns the emoji for a given item quality string.
func dropQualityEmoji(qualityStr string, itemName string) string {
	if strings.Contains(strings.ToLower(itemName), "rune") {
		return "🟣"
	}

	switch strings.ToLower(qualityStr) {
	case "unique":
		return "🟠"
	case "set":
		return "🟢"
	case "rare":
		return "🟡"
	case "magic":
		return "🔵"
	default:
		return "⚪"
	}
}

// buildDropsEmbed creates the embed for the !drops command.
func buildDropsEmbed(supervisor string, drops []data.Drop, count int) *discordgo.MessageEmbed {
	startIdx := len(drops) - count
	if startIdx < 0 {
		startIdx = 0
	}
	recentDrops := drops[startIdx:]

	var description strings.Builder

	// Reverse to show newest first
	for i := len(recentDrops) - 1; i >= 0; i-- {
		drop := recentDrops[i]
		item := drop.Item

		qualityStr := item.Quality.ToString()
		emoji := dropQualityEmoji(qualityStr, string(item.Name))

		itemName := string(item.Name)
		if qualityStr != "" && qualityStr != "Normal" {
			itemName = fmt.Sprintf("%s %s", qualityStr, string(item.Name))
		}

		description.WriteString(fmt.Sprintf("%s **%s**", emoji, itemName))

		desc := item.Desc()
		if desc.Name != "" && desc.Name != string(item.Name) {
			description.WriteString(fmt.Sprintf(" (%s)", desc.Name))
		}

		description.WriteString("\n")
	}

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("💎 Recent Drops for %s", supervisor),
		Description: description.String(),
		Color:       0xFFD700,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Showing last %d of %d total drops", len(recentDrops), len(drops)),
		},
	}
}
