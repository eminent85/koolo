package discordv2

import (
	"strings"
	"testing"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
)

// ---------------------------------------------------------------------------
// buildHelpEmbed
// ---------------------------------------------------------------------------

func TestBuildHelpEmbed(t *testing.T) {
	embed := buildHelpEmbed("!")

	if embed.Title == "" {
		t.Error("help embed should have a title")
	}
	if embed.Color != 0x5865F2 {
		t.Errorf("help embed color = 0x%06x, want 0x5865F2", embed.Color)
	}
	if len(embed.Fields) != 6 {
		t.Errorf("help embed should have 6 fields, got %d", len(embed.Fields))
	}
	if embed.Footer == nil {
		t.Error("help embed should have a footer")
	}

	// Verify all commands are listed
	commands := []string{"!list", "!start", "!stop", "!stats", "!drops", "!help"}
	for _, cmd := range commands {
		found := false
		for _, f := range embed.Fields {
			if strings.Contains(f.Name, cmd) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("help embed missing command %q", cmd)
		}
	}
}

func TestBuildHelpEmbed_CustomPrefix(t *testing.T) {
	embed := buildHelpEmbed("$")

	for _, f := range embed.Fields {
		if !strings.HasPrefix(f.Name, "$") {
			t.Errorf("field name %q should start with custom prefix '$'", f.Name)
		}
		// Examples in field values should also use the custom prefix, not "!"
		if strings.Contains(f.Value, "`!") {
			t.Errorf("field %q value contains stale '!' prefix in examples: %s", f.Name, f.Value)
		}
	}
}

// ---------------------------------------------------------------------------
// buildStatsEmbed
// ---------------------------------------------------------------------------

func TestBuildStatsEmbed_Online(t *testing.T) {
	stats := SupervisorStats{
		SupervisorStatus: StatusInGame,
		StartedAt:        time.Now().Add(-1 * time.Hour),
		TotalGames:       50,
		TotalDeaths:      2,
		TotalChickens:    3,
		TotalErrors:      1,
		Drops:            make([]data.Drop, 10),
	}

	embed := buildStatsEmbed("Koza", stats, false)

	if embed.Title != "Stats for Koza" {
		t.Errorf("title = %q, want %q", embed.Title, "Stats for Koza")
	}
	if len(embed.Fields) != 7 {
		t.Fatalf("expected 7 fields, got %d", len(embed.Fields))
	}

	fieldValues := map[string]string{}
	for _, f := range embed.Fields {
		fieldValues[f.Name] = f.Value
	}

	if fieldValues["Status"] != string(StatusInGame) {
		t.Errorf("Status = %q, want %q", fieldValues["Status"], StatusInGame)
	}
	if fieldValues["Games"] != "50" {
		t.Errorf("Games = %q, want %q", fieldValues["Games"], "50")
	}
	if fieldValues["Drops"] != "10" {
		t.Errorf("Drops = %q, want %q", fieldValues["Drops"], "10")
	}
	if fieldValues["Deaths"] != "2" {
		t.Errorf("Deaths = %q, want %q", fieldValues["Deaths"], "2")
	}
	if fieldValues["Chickens"] != "3" {
		t.Errorf("Chickens = %q, want %q", fieldValues["Chickens"], "3")
	}
	if fieldValues["Errors"] != "1" {
		t.Errorf("Errors = %q, want %q", fieldValues["Errors"], "1")
	}
}

func TestBuildStatsEmbed_Offline(t *testing.T) {
	stats := SupervisorStats{
		SupervisorStatus: StatusNotStarted,
	}

	embed := buildStatsEmbed("Koza", stats, false)

	for _, f := range embed.Fields {
		if f.Name == "Status" && f.Value != "Offline" {
			t.Errorf("offline status should show 'Offline', got %q", f.Value)
		}
	}
}

func TestBuildStatsEmbed_EmptyStatus(t *testing.T) {
	stats := SupervisorStats{
		SupervisorStatus: "",
	}

	embed := buildStatsEmbed("Koza", stats, false)

	for _, f := range embed.Fields {
		if f.Name == "Status" && f.Value != "Offline" {
			t.Errorf("empty status should show 'Offline', got %q", f.Value)
		}
	}
}

func TestBuildStatsEmbed_AllFieldsInline(t *testing.T) {
	embed := buildStatsEmbed("Koza", SupervisorStats{SupervisorStatus: StatusInGame}, false)
	for _, f := range embed.Fields {
		if !f.Inline {
			t.Errorf("field %q should be inline", f.Name)
		}
	}
}

func TestBuildStatsEmbed_Verbose_WithCharacterData(t *testing.T) {
	stats := SupervisorStats{
		SupervisorStatus: StatusInGame,
		StartedAt:        time.Now().Add(-30 * time.Minute),
		TotalGames:       10,
		Character: CharacterInfo{
			Class:           "Sorceress",
			Level:           85,
			Area:            "Throne of Destruction",
			Difficulty:      "Hell",
			Life:            1200,
			MaxLife:         1500,
			Mana:            400,
			MaxMana:         600,
			MagicFind:       250,
			GoldFind:        120,
			FireResist:      75,
			ColdResist:      60,
			LightningResist: 75,
			PoisonResist:    55,
			Ping:            42,
		},
	}

	embed := buildStatsEmbed("Koza", stats, true)

	// 7 base fields + 7 verbose fields
	if len(embed.Fields) != 14 {
		t.Fatalf("verbose embed with character data should have 14 fields, got %d", len(embed.Fields))
	}

	fieldValues := map[string]string{}
	for _, f := range embed.Fields {
		fieldValues[f.Name] = f.Value
	}

	if !strings.Contains(fieldValues["Character"], "Sorceress") {
		t.Errorf("Character field should contain class, got: %s", fieldValues["Character"])
	}
	if !strings.Contains(fieldValues["Character"], "85") {
		t.Errorf("Character field should contain level, got: %s", fieldValues["Character"])
	}
	if !strings.Contains(fieldValues["Location"], "Throne of Destruction") {
		t.Errorf("Location field should contain area, got: %s", fieldValues["Location"])
	}
	if !strings.Contains(fieldValues["Location"], "Hell") {
		t.Errorf("Location field should contain difficulty, got: %s", fieldValues["Location"])
	}
	if !strings.Contains(fieldValues["Life"], "1200") || !strings.Contains(fieldValues["Life"], "1500") {
		t.Errorf("Life field should contain current/max life, got: %s", fieldValues["Life"])
	}
	if !strings.Contains(fieldValues["Mana"], "400") || !strings.Contains(fieldValues["Mana"], "600") {
		t.Errorf("Mana field should contain current/max mana, got: %s", fieldValues["Mana"])
	}
	if !strings.Contains(fieldValues["MF / GF"], "250") || !strings.Contains(fieldValues["MF / GF"], "120") {
		t.Errorf("MF / GF field should contain MF and GF values, got: %s", fieldValues["MF / GF"])
	}
	if !strings.Contains(fieldValues["Ping"], "42") {
		t.Errorf("Ping field should contain value, got: %s", fieldValues["Ping"])
	}
	if !strings.Contains(fieldValues["Resistances"], "FR: 75") || !strings.Contains(fieldValues["Resistances"], "CR: 60") {
		t.Errorf("Resistances field should contain all resist values, got: %s", fieldValues["Resistances"])
	}
}

func TestBuildStatsEmbed_Verbose_WithoutCharacterData(t *testing.T) {
	// verbose=true but Class is empty (supervisor offline / no game data yet)
	stats := SupervisorStats{
		SupervisorStatus: StatusNotStarted,
	}

	embed := buildStatsEmbed("Koza", stats, true)

	if len(embed.Fields) != 7 {
		t.Fatalf("verbose embed without character data should have 7 fields, got %d", len(embed.Fields))
	}
}

// ---------------------------------------------------------------------------
// buildListEmbed
// ---------------------------------------------------------------------------

func TestBuildListEmbed_MixedStatus(t *testing.T) {
	mgr := newMockManager()
	mgr.addSupervisor("Koza", StatusInGame, time.Now().Add(-2*time.Hour), nil, 0, 0, 0, 0)
	mgr.addSupervisor("Ovca", StatusNotStarted, time.Time{}, nil, 0, 0, 0, 0)

	embed := buildListEmbed([]string{"Koza", "Ovca"}, mgr)

	if embed.Title == "" {
		t.Error("list embed should have a title")
	}
	if embed.Color != 0x5865F2 {
		t.Errorf("list embed color = 0x%06x, want 0x5865F2", embed.Color)
	}
	if len(embed.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(embed.Fields))
	}

	// Find Koza's field
	for _, f := range embed.Fields {
		if f.Name == "Koza" {
			if !strings.Contains(f.Value, "✅") {
				t.Errorf("running supervisor should have ✅, got: %s", f.Value)
			}
			if strings.Contains(f.Value, "Uptime: -") {
				t.Errorf("running supervisor should have uptime, got: %s", f.Value)
			}
		}
		if f.Name == "Ovca" {
			if !strings.Contains(f.Value, "❌ Offline") {
				t.Errorf("offline supervisor should have ❌ Offline, got: %s", f.Value)
			}
			if !strings.Contains(f.Value, "Uptime: -") {
				t.Errorf("offline supervisor should have Uptime: -, got: %s", f.Value)
			}
		}
	}
}

func TestBuildListEmbed_AllFieldsInline(t *testing.T) {
	mgr := newMockManager()
	mgr.addSupervisor("Koza", StatusInGame, time.Now(), nil, 0, 0, 0, 0)

	embed := buildListEmbed([]string{"Koza"}, mgr)
	for _, f := range embed.Fields {
		if !f.Inline {
			t.Errorf("field %q should be inline", f.Name)
		}
	}
}

func TestBuildListEmbed_UptimeFormatting(t *testing.T) {
	mgr := newMockManager()

	// 30 seconds ago
	mgr.addSupervisor("Quick", StatusInGame, time.Now().Add(-30*time.Second), nil, 0, 0, 0, 0)
	// 15 minutes ago
	mgr.addSupervisor("Medium", StatusInGame, time.Now().Add(-15*time.Minute), nil, 0, 0, 0, 0)
	// 3 hours 25 minutes ago
	mgr.addSupervisor("Long", StatusInGame, time.Now().Add(-3*time.Hour-25*time.Minute), nil, 0, 0, 0, 0)

	embed := buildListEmbed([]string{"Quick", "Medium", "Long"}, mgr)

	for _, f := range embed.Fields {
		switch f.Name {
		case "Quick":
			if !strings.Contains(f.Value, "s") {
				t.Errorf("short uptime should use seconds, got: %s", f.Value)
			}
		case "Medium":
			if !strings.Contains(f.Value, "m") {
				t.Errorf("medium uptime should use minutes, got: %s", f.Value)
			}
		case "Long":
			if !strings.Contains(f.Value, "h") || !strings.Contains(f.Value, "m") {
				t.Errorf("long uptime should use hours and minutes, got: %s", f.Value)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// dropQualityEmoji
// ---------------------------------------------------------------------------

func TestDropQualityEmoji(t *testing.T) {
	tests := []struct {
		quality  string
		itemName string
		want     string
	}{
		{"Unique", "SharkstoothArmor", "🟠"},
		{"Set", "Ring", "🟢"},
		{"Rare", "Amulet", "🟡"},
		{"Magic", "Shield", "🔵"},
		{"Superior", "Sword", "⚪"},
		{"Normal", "Gem", "⚪"},
		{"", "Unknown", "⚪"},
		// Rune detection overrides quality
		{"Normal", "ElRune", "🟣"},
		{"Normal", "BerRune", "🟣"},
		{"Normal", "RUNE", "🟣"},
	}
	for _, tt := range tests {
		t.Run(tt.quality+"_"+tt.itemName, func(t *testing.T) {
			got := dropQualityEmoji(tt.quality, tt.itemName)
			if got != tt.want {
				t.Errorf("dropQualityEmoji(%q, %q) = %q, want %q", tt.quality, tt.itemName, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// buildDropsEmbed
// ---------------------------------------------------------------------------

func TestBuildDropsEmbed_BasicFields(t *testing.T) {
	drops := []data.Drop{
		{Item: data.Item{Name: item.Name("Shako"), Quality: item.QualityUnique}},
		{Item: data.Item{Name: item.Name("Arachnid"), Quality: item.QualityUnique}},
		{Item: data.Item{Name: item.Name("Harlequin"), Quality: item.QualitySet}},
	}

	embed := buildDropsEmbed("Koza", drops, 5)

	if !strings.Contains(embed.Title, "Koza") {
		t.Errorf("title should contain supervisor name, got: %s", embed.Title)
	}
	if embed.Color != 0xFFD700 {
		t.Errorf("color = 0x%06x, want 0xFFD700", embed.Color)
	}
	if embed.Footer == nil {
		t.Fatal("drops embed should have a footer")
	}
	if !strings.Contains(embed.Footer.Text, "3") {
		t.Errorf("footer should mention total drops, got: %s", embed.Footer.Text)
	}
}

func TestBuildDropsEmbed_RespectsCount(t *testing.T) {
	drops := make([]data.Drop, 10)
	for i := range drops {
		drops[i] = data.Drop{Item: data.Item{
			Name:    item.Name("Item" + string(rune('A'+i))),
			Quality: item.QualityNormal,
		}}
	}

	embed := buildDropsEmbed("Koza", drops, 3)

	// Should only show last 3
	lines := strings.Split(strings.TrimSpace(embed.Description), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %v", len(lines), lines)
	}
}

func TestBuildDropsEmbed_CountLargerThanDrops(t *testing.T) {
	drops := []data.Drop{
		{Item: data.Item{Name: item.Name("Shako"), Quality: item.QualityUnique}},
	}

	embed := buildDropsEmbed("Koza", drops, 20)

	lines := strings.Split(strings.TrimSpace(embed.Description), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(lines))
	}
}

func TestBuildDropsEmbed_NewestFirst(t *testing.T) {
	drops := []data.Drop{
		{Item: data.Item{Name: item.Name("OldItem"), Quality: item.QualityNormal}},
		{Item: data.Item{Name: item.Name("NewItem"), Quality: item.QualityNormal}},
	}

	embed := buildDropsEmbed("Koza", drops, 5)

	lines := strings.Split(strings.TrimSpace(embed.Description), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}
	// NewItem should appear before OldItem
	newIdx := strings.Index(embed.Description, "NewItem")
	oldIdx := strings.Index(embed.Description, "OldItem")
	if newIdx > oldIdx {
		t.Error("newest drop should appear first")
	}
}

func TestBuildDropsEmbed_QualityInName(t *testing.T) {
	drops := []data.Drop{
		{Item: data.Item{Name: item.Name("Shako"), Quality: item.QualityUnique}},
	}

	embed := buildDropsEmbed("Koza", drops, 5)

	// "Unique Shako" should appear
	if !strings.Contains(embed.Description, "Unique Shako") {
		t.Errorf("expected 'Unique Shako' in description, got: %s", embed.Description)
	}
}

func TestBuildDropsEmbed_NormalQuality_NoPrefix(t *testing.T) {
	drops := []data.Drop{
		{Item: data.Item{Name: item.Name("Gem"), Quality: item.QualityNormal}},
	}

	embed := buildDropsEmbed("Koza", drops, 5)

	// Should NOT have "Normal Gem", just "Gem"
	if strings.Contains(embed.Description, "Normal Gem") {
		t.Errorf("normal quality should not prefix name, got: %s", embed.Description)
	}
	if !strings.Contains(embed.Description, "Gem") {
		t.Errorf("should contain item name, got: %s", embed.Description)
	}
}

func TestBuildDropsEmbed_RuneEmoji(t *testing.T) {
	drops := []data.Drop{
		{Item: data.Item{Name: item.Name("BerRune"), Quality: item.QualityNormal}},
	}

	embed := buildDropsEmbed("Koza", drops, 5)

	if !strings.Contains(embed.Description, "🟣") {
		t.Errorf("rune should use 🟣 emoji, got: %s", embed.Description)
	}
}

func TestBuildDropsEmbed_QualityEmojis(t *testing.T) {
	drops := []data.Drop{
		{Item: data.Item{Name: item.Name("UniqueItem"), Quality: item.QualityUnique}},
		{Item: data.Item{Name: item.Name("SetItem"), Quality: item.QualitySet}},
		{Item: data.Item{Name: item.Name("RareItem"), Quality: item.QualityRare}},
		{Item: data.Item{Name: item.Name("MagicItem"), Quality: item.QualityMagic}},
	}

	embed := buildDropsEmbed("Koza", drops, 10)

	expected := map[string]string{
		"UniqueItem": "🟠",
		"SetItem":    "🟢",
		"RareItem":   "🟡",
		"MagicItem":  "🔵",
	}

	for name, emoji := range expected {
		// Find the line with this item
		for _, line := range strings.Split(embed.Description, "\n") {
			if strings.Contains(line, name) {
				if !strings.Contains(line, emoji) {
					t.Errorf("item %q should have emoji %s, got line: %s", name, emoji, line)
				}
				break
			}
		}
	}
}

// ---------------------------------------------------------------------------
// supervisorExists (via mockManager)
// ---------------------------------------------------------------------------

func TestSupervisorExists(t *testing.T) {
	mgr := newMockManager()
	mgr.addSupervisor("Koza", StatusInGame, time.Now(), nil, 0, 0, 0, 0)

	b := &Bot{manager: mgr}

	if !b.supervisorExists("Koza") {
		t.Error("supervisorExists should return true for existing supervisor")
	}
	if b.supervisorExists("NonExistent") {
		t.Error("supervisorExists should return false for non-existing supervisor")
	}
}
