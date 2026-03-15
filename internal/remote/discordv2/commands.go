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

		if err := b.manager.Start(supervisor, false, false); err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Failed to start supervisor '%s': %s", supervisor, err))
			continue
		}
		time.Sleep(1 * time.Second)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' started successfully.", supervisor))
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

func (b *Bot) handleStats(s *discordgo.Session, m *discordgo.MessageCreate) {
	words := strings.Fields(m.Content)

	verbose := false
	var supervisors []string
	for _, w := range words[1:] {
		if w == "-v" || w == "--verbose" {
			verbose = true
		} else {
			supervisors = append(supervisors, w)
		}
	}

	if len(supervisors) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !stats [-v] <supervisor1> [supervisor2] ...")
		return
	}

	for _, supervisor := range supervisors {
		if !b.supervisorExists(supervisor) {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Supervisor '%s' not found.", supervisor))
			continue
		}

		embed := buildStatsEmbed(supervisor, b.manager.Status(supervisor), verbose)
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
	s.ChannelMessageSendEmbed(m.ChannelID, buildHelpEmbed(b.commandPrefix()))
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

// buildStatsEmbed creates the embed for the !stats / !status command.
// When verbose is true and live character data is available, additional
// in-game fields (character, location, vitals, resists, MF/GF, ping) are appended.
func buildStatsEmbed(supervisor string, stats SupervisorStats, verbose bool) *discordgo.MessageEmbed {
	supStatus := string(stats.SupervisorStatus)
	if supStatus == string(StatusNotStarted) || supStatus == "" {
		supStatus = "Offline"
	}

	fields := []*discordgo.MessageEmbedField{
		{Name: "Status", Value: supStatus, Inline: true},
		{Name: "Uptime", Value: time.Since(stats.StartedAt).String(), Inline: true},
		{Name: "Games", Value: fmt.Sprintf("%d", stats.TotalGames), Inline: true},
		{Name: "Drops", Value: fmt.Sprintf("%d", len(stats.Drops)), Inline: true},
		{Name: "Deaths", Value: fmt.Sprintf("%d", stats.TotalDeaths), Inline: true},
		{Name: "Chickens", Value: fmt.Sprintf("%d", stats.TotalChickens), Inline: true},
		{Name: "Errors", Value: fmt.Sprintf("%d", stats.TotalErrors), Inline: true},
	}

	if verbose && stats.Character.Class != "" {
		ch := stats.Character
		fields = append(fields,
			&discordgo.MessageEmbedField{Name: "Character", Value: fmt.Sprintf("%s (Level %d)", ch.Class, ch.Level), Inline: true},
			&discordgo.MessageEmbedField{Name: "Location", Value: fmt.Sprintf("%s, %s", ch.Area, ch.Difficulty), Inline: true},
			&discordgo.MessageEmbedField{Name: "Life", Value: fmt.Sprintf("%d / %d", ch.Life, ch.MaxLife), Inline: true},
			&discordgo.MessageEmbedField{Name: "Mana", Value: fmt.Sprintf("%d / %d", ch.Mana, ch.MaxMana), Inline: true},
			&discordgo.MessageEmbedField{Name: "MF / GF", Value: fmt.Sprintf("%d%% / %d%%", ch.MagicFind, ch.GoldFind), Inline: true},
			&discordgo.MessageEmbedField{Name: "Ping", Value: fmt.Sprintf("%dms", ch.Ping), Inline: true},
			&discordgo.MessageEmbedField{Name: "Resistances", Value: fmt.Sprintf("FR: %d | CR: %d | LR: %d | PR: %d", ch.FireResist, ch.ColdResist, ch.LightningResist, ch.PoisonResist), Inline: false},
		)
	}

	return &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("Stats for %s", supervisor),
		Fields: fields,
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

// buildHelpEmbed creates the embed for the help command.
func buildHelpEmbed(prefix string) *discordgo.MessageEmbed {
	p := prefix
	return &discordgo.MessageEmbed{
		Title:       "🤖 Koolo Discord Bot Commands",
		Description: "Control and monitor your Diablo II bot supervisors",
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: p + "list", Value: "Show all available supervisors with their status and uptime", Inline: false},
			{Name: p + "start <supervisor1> [supervisor2] ...", Value: fmt.Sprintf("Start one or more supervisors\nExample: `%sstart Koza Ovca`", p), Inline: false},
			{Name: p + "stop <supervisor1> [supervisor2] ...", Value: fmt.Sprintf("Stop one or more supervisors\nExample: `%sstop Koza`", p), Inline: false},
			{Name: p + "stats [-v] <supervisor1> [supervisor2] ...", Value: fmt.Sprintf("Show supervisor stats. Add `-v` for verbose output including live character info\nExamples: `%sstats Koza` · `%sstats -v Koza` · `%sstatus Koza`", p, p, p), Inline: false},
			{Name: p + "drops <supervisor> [count]", Value: fmt.Sprintf("Show recent drops for a supervisor\nExample: `%sdrops Koza 10`\nDefault count: 5", p), Inline: false},
			{Name: p + "help", Value: "Show this help message", Inline: false},
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
