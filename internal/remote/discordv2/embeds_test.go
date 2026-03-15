package discordv2

import (
	"strings"
	"testing"

	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/event"
)

// ---------------------------------------------------------------------------
// qualityColor
// ---------------------------------------------------------------------------

func TestQualityColor(t *testing.T) {
	tests := []struct {
		quality string
		want    int
	}{
		{"LowQuality", 0x666666},
		{"Normal", 0xffffff},
		{"Superior", 0xc0c0c0},
		{"Magic", 0x6969ff},
		{"Set", 0x00ff00},
		{"Rare", 0xffff77},
		{"Unique", 0xbfa969},
		{"Crafted", 0xff8000},
		{"", 0x999999},
		{"InvalidQuality", 0x999999},
	}

	for _, tt := range tests {
		t.Run(tt.quality, func(t *testing.T) {
			if got := qualityColor(tt.quality); got != tt.want {
				t.Errorf("qualityColor(%q) = 0x%06x, want 0x%06x", tt.quality, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// formatDamageLine
// ---------------------------------------------------------------------------

func TestFormatDamageLine(t *testing.T) {
	tests := []struct {
		name       string
		min, max   int
		damageType string
		want       string
	}{
		{"both zero", 0, 0, "Fire", ""},
		{"min only", 5, 0, "Fire", "Adds 5-0 Fire Damage\n"},
		{"max only", 0, 10, "Lightning", "Adds 0-10 Lightning Damage\n"},
		{"both set", 3, 7, "Cold", "Adds 3-7 Cold Damage\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDamageLine(tt.min, tt.max, tt.damageType)
			if got != tt.want {
				t.Errorf("formatDamageLine(%d, %d, %q) = %q, want %q", tt.min, tt.max, tt.damageType, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// findPartialAllResists
// ---------------------------------------------------------------------------

func TestFindPartialAllResists(t *testing.T) {
	tests := []struct {
		name             string
		fr, cr, lr, pr   int
		wantFound        bool
		wantValue        int
		wantOutlierLabel string
	}{
		{"all equal", 30, 30, 30, 30, false, 0, ""},
		{"all different", 10, 20, 30, 40, false, 0, ""},
		{"fire outlier", 15, 30, 30, 30, true, 30, "Fire Resist"},
		{"cold outlier", 30, 15, 30, 30, true, 30, "Cold Resist"},
		{"lightning outlier", 30, 30, 15, 30, true, 30, "Lightning Resist"},
		{"poison outlier", 30, 30, 30, 15, true, 30, "Poison Resist"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, val, id := findPartialAllResists(tt.fr, tt.cr, tt.lr, tt.pr)
			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}
			if found {
				if val != tt.wantValue {
					t.Errorf("value = %d, want %d", val, tt.wantValue)
				}
				if resistLabel(id) != tt.wantOutlierLabel {
					t.Errorf("outlier = %q, want %q", resistLabel(id), tt.wantOutlierLabel)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// resistLabel
// ---------------------------------------------------------------------------

func TestResistLabel(t *testing.T) {
	tests := []struct {
		id   stat.ID
		want string
	}{
		{stat.FireResist, "Fire Resist"},
		{stat.ColdResist, "Cold Resist"},
		{stat.LightningResist, "Lightning Resist"},
		{stat.PoisonResist, "Poison Resist"},
		{stat.ID(9999), "Resist"},
	}
	for _, tt := range tests {
		if got := resistLabel(tt.id); got != tt.want {
			t.Errorf("resistLabel(%d) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// resistValueForID
// ---------------------------------------------------------------------------

func TestResistValueForID(t *testing.T) {
	fr, cr, lr, pr := 10, 20, 30, 40
	tests := []struct {
		id   stat.ID
		want int
	}{
		{stat.FireResist, 10},
		{stat.ColdResist, 20},
		{stat.LightningResist, 30},
		{stat.PoisonResist, 40},
		{stat.ID(9999), 0},
	}
	for _, tt := range tests {
		if got := resistValueForID(tt.id, fr, cr, lr, pr); got != tt.want {
			t.Errorf("resistValueForID(%d) = %d, want %d", tt.id, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// formatRuleLocation
// ---------------------------------------------------------------------------

func TestFormatRuleLocation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"spaces only", "   ", ""},
		{"bare filename", "pickit.nip", "pickit.nip"},
		{"path with line", "/home/user/pickit.nip:42", "pickit.nip : line 42"},
		{"path no line", "/home/user/pickit.nip", "pickit.nip"},
		{"colon but non-digit tail", "/home/user/pickit.nip:abc", "pickit.nip:abc"},
		{"windows path with line", `C:\Users\foo\pickit.nip:10`, "pickit.nip : line 10"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatRuleLocation(tt.input)
			if got != tt.want {
				t.Errorf("formatRuleLocation(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// isAllDigits
// ---------------------------------------------------------------------------

func TestIsAllDigits(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"", false},
		{"0", true},
		{"42", true},
		{"12345", true},
		{"12a45", false},
		{"abc", false},
		{" 42", false},
	}
	for _, tt := range tests {
		if got := isAllDigits(tt.input); got != tt.want {
			t.Errorf("isAllDigits(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// buildItemStashEmbed — basic fields
// ---------------------------------------------------------------------------

func TestBuildItemStashEmbed_BasicFields(t *testing.T) {
	drop := makeTestDropWithName("SharkstoothArmor", item.QualityUnique)
	evt := event.ItemStashed(event.Text("Koza", "Item stashed"), drop)

	embed := buildItemStashEmbed(evt, false)

	if embed == nil {
		t.Fatal("embed should not be nil")
	}
	if embed.Color != qualityColor("Unique") {
		t.Errorf("color = 0x%06x, want 0x%06x", embed.Color, qualityColor("Unique"))
	}
	if !strings.Contains(embed.Description, "SharkstoothArmor") {
		t.Errorf("description should contain item name, got: %s", embed.Description)
	}
	if !strings.Contains(embed.Description, "Koza") {
		t.Errorf("description should contain supervisor name, got: %s", embed.Description)
	}
}

func TestBuildItemStashEmbed_UsesIdentifiedName(t *testing.T) {
	drop := makeTestDropIdentified("SharkstoothArmor", "Shaftstop", item.QualityUnique)
	evt := event.ItemStashed(event.Text("Koza", "Item stashed"), drop)

	embed := buildItemStashEmbed(evt, false)

	if !strings.Contains(embed.Description, "Shaftstop") {
		t.Errorf("description should use identified name, got: %s", embed.Description)
	}
}

func TestBuildItemStashEmbed_QualityBadge(t *testing.T) {
	drop := makeTestDropWithName("Ring", item.QualitySet)
	evt := event.ItemStashed(event.Text("Koza", "Item stashed"), drop)

	embed := buildItemStashEmbed(evt, false)

	if !strings.Contains(embed.Description, "[Set]") {
		t.Errorf("description should contain quality badge, got: %s", embed.Description)
	}
}

func TestBuildItemStashEmbed_ColorMatchesQuality(t *testing.T) {
	qualities := []struct {
		quality item.Quality
		name    string
		color   int
	}{
		{item.QualityUnique, "Unique", 0xbfa969},
		{item.QualitySet, "Set", 0x00ff00},
		{item.QualityRare, "Rare", 0xffff77},
		{item.QualityMagic, "Magic", 0x6969ff},
		{item.QualityNormal, "Normal", 0xffffff},
	}

	for _, tt := range qualities {
		t.Run(tt.name, func(t *testing.T) {
			drop := makeTestDropWithName("TestItem", tt.quality)
			evt := event.ItemStashed(event.Text("Koza", "Item stashed"), drop)
			embed := buildItemStashEmbed(evt, false)
			if embed.Color != tt.color {
				t.Errorf("color = 0x%06x, want 0x%06x", embed.Color, tt.color)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// buildItemStashEmbed — defense
// ---------------------------------------------------------------------------

func TestBuildItemStashEmbed_Defense(t *testing.T) {
	stats := stat.Stats{
		{ID: 31, Value: 500}, // defense
	}
	drop := makeTestDropWithStats("Armor", item.QualityUnique, stats)
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, false)

	if !strings.Contains(embed.Description, "Defense: 500") {
		t.Errorf("expected Defense: 500 in description, got: %s", embed.Description)
	}
}

func TestBuildItemStashEmbed_NoDefenseWhenZero(t *testing.T) {
	stats := stat.Stats{
		{ID: stat.Strength, Value: 10},
	}
	drop := makeTestDropWithStats("Ring", item.QualityUnique, stats)
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, false)

	if strings.Contains(embed.Description, "Defense:") {
		t.Errorf("should not contain Defense when zero, got: %s", embed.Description)
	}
}

// ---------------------------------------------------------------------------
// buildItemStashEmbed — All Attributes combined
// ---------------------------------------------------------------------------

func TestBuildItemStashEmbed_AllAttributes(t *testing.T) {
	stats := stat.Stats{
		{ID: stat.Strength, Value: 20},
		{ID: stat.Energy, Value: 20},
		{ID: stat.Dexterity, Value: 20},
		{ID: stat.Vitality, Value: 20},
	}
	drop := makeTestDropWithStats("Annihilus", item.QualityUnique, stats)
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, false)

	if !strings.Contains(embed.Description, "+20 to All Attributes") {
		t.Errorf("expected All Attributes line, got: %s", embed.Description)
	}
	// Individual stats should be suppressed
	if strings.Contains(embed.Description, "+20 to Strength") {
		t.Errorf("individual Strength should be suppressed when combined, got: %s", embed.Description)
	}
}

func TestBuildItemStashEmbed_NoAllAttributes_WhenUnequal(t *testing.T) {
	stats := stat.Stats{
		{ID: stat.Strength, Value: 20},
		{ID: stat.Energy, Value: 15},
		{ID: stat.Dexterity, Value: 20},
		{ID: stat.Vitality, Value: 20},
	}
	drop := makeTestDropWithStats("Ring", item.QualityUnique, stats)
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, false)

	if strings.Contains(embed.Description, "All Attributes") {
		t.Errorf("should not show All Attributes when values differ, got: %s", embed.Description)
	}
}

// ---------------------------------------------------------------------------
// buildItemStashEmbed — All Resistances combined
// ---------------------------------------------------------------------------

func TestBuildItemStashEmbed_AllResists(t *testing.T) {
	stats := stat.Stats{
		{ID: stat.FireResist, Value: 30},
		{ID: stat.ColdResist, Value: 30},
		{ID: stat.LightningResist, Value: 30},
		{ID: stat.PoisonResist, Value: 30},
	}
	drop := makeTestDropWithStats("Annihilus", item.QualityUnique, stats)
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, false)

	if !strings.Contains(embed.Description, "All Resistances +30") {
		t.Errorf("expected All Resistances +30, got: %s", embed.Description)
	}
}

func TestBuildItemStashEmbed_PartialResists(t *testing.T) {
	stats := stat.Stats{
		{ID: stat.FireResist, Value: 30},
		{ID: stat.ColdResist, Value: 30},
		{ID: stat.LightningResist, Value: 30},
		{ID: stat.PoisonResist, Value: 15},
	}
	drop := makeTestDropWithStats("Ring", item.QualityUnique, stats)
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, false)

	if !strings.Contains(embed.Description, "All Resistances +30") {
		t.Errorf("expected partial All Resistances +30, got: %s", embed.Description)
	}
	if !strings.Contains(embed.Description, "Poison Resist +15") {
		t.Errorf("expected outlier Poison Resist +15, got: %s", embed.Description)
	}
}

func TestBuildItemStashEmbed_NoAllResists_WhenAllDifferent(t *testing.T) {
	stats := stat.Stats{
		{ID: stat.FireResist, Value: 10},
		{ID: stat.ColdResist, Value: 20},
		{ID: stat.LightningResist, Value: 30},
		{ID: stat.PoisonResist, Value: 40},
	}
	drop := makeTestDropWithStats("Ring", item.QualityUnique, stats)
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, false)

	if strings.Contains(embed.Description, "All Resistances") {
		t.Errorf("should not show All Resistances when all different, got: %s", embed.Description)
	}
}

// ---------------------------------------------------------------------------
// buildItemStashEmbed — Enhanced Damage
// ---------------------------------------------------------------------------

func TestBuildItemStashEmbed_EnhancedDamage(t *testing.T) {
	tests := []struct {
		name    string
		eMin    int
		eMax    int
		wantStr string
	}{
		{"max only", 0, 200, "+200% Enhanced Damage"},
		{"min only", 150, 0, "+150% Enhanced Damage"},
		{"range", 100, 200, "+100-200% Enhanced Damage"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := stat.Stats{}
			if tt.eMin > 0 {
				stats = append(stats, stat.Data{ID: 17, Value: tt.eMin})
			}
			if tt.eMax > 0 {
				stats = append(stats, stat.Data{ID: 18, Value: tt.eMax})
			}
			drop := makeTestDropWithStats("Weapon", item.QualityUnique, stats)
			evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
			embed := buildItemStashEmbed(evt, false)

			if !strings.Contains(embed.Description, tt.wantStr) {
				t.Errorf("expected %q, got: %s", tt.wantStr, embed.Description)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// buildItemStashEmbed — Elemental Damage lines
// ---------------------------------------------------------------------------

func TestBuildItemStashEmbed_ElementalDamage(t *testing.T) {
	stats := stat.Stats{
		{ID: 48, Value: 5}, {ID: 49, Value: 10}, // fire
		{ID: 50, Value: 1}, {ID: 51, Value: 50}, // lightning
		{ID: 54, Value: 3}, {ID: 55, Value: 7}, // cold
		{ID: 52, Value: 2}, {ID: 53, Value: 8}, // magic
		{ID: 57, Value: 20}, {ID: 58, Value: 100}, // poison
	}
	drop := makeTestDropWithStats("Weapon", item.QualityUnique, stats)
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, false)

	expected := []string{
		"Adds 5-10 Fire Damage",
		"Adds 1-50 Lightning Damage",
		"Adds 3-7 Cold Damage",
		"Adds 2-8 Magic Damage",
		"Adds 20-100 Poison Damage",
	}
	for _, want := range expected {
		if !strings.Contains(embed.Description, want) {
			t.Errorf("expected %q in description, got: %s", want, embed.Description)
		}
	}
}

// ---------------------------------------------------------------------------
// buildItemStashEmbed — Ethereal and Sockets
// ---------------------------------------------------------------------------

func TestBuildItemStashEmbed_Ethereal(t *testing.T) {
	drop := makeTestDropEthereal("Armor", item.QualityUnique)
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, false)

	if !strings.Contains(embed.Description, "Ethereal") {
		t.Errorf("expected Ethereal flag, got: %s", embed.Description)
	}
}

func TestBuildItemStashEmbed_Sockets(t *testing.T) {
	drop := makeTestDropWithSockets("Armor", item.QualityNormal, 4)
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, false)

	if !strings.Contains(embed.Description, "Sockets: 4") {
		t.Errorf("expected Sockets: 4, got: %s", embed.Description)
	}
}

func TestBuildItemStashEmbed_NoSocketLine_WhenZero(t *testing.T) {
	drop := makeTestDropWithName("Ring", item.QualityUnique)
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, false)

	if strings.Contains(embed.Description, "Sockets:") {
		t.Errorf("should not contain Sockets line when zero, got: %s", embed.Description)
	}
}

// ---------------------------------------------------------------------------
// buildItemStashEmbed — Pickit info
// ---------------------------------------------------------------------------

func TestBuildItemStashEmbed_PickitInfo_Enabled(t *testing.T) {
	drop := makeTestDropWithPickit("Ring", item.QualityUnique, "/path/to/rules.nip:42", "[Type] == Ring")
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, true)

	if !strings.Contains(embed.Description, "rules.nip : line 42") {
		t.Errorf("expected rule location, got: %s", embed.Description)
	}
	if !strings.Contains(embed.Description, "[Type] == Ring") {
		t.Errorf("expected rule text, got: %s", embed.Description)
	}
}

func TestBuildItemStashEmbed_PickitInfo_Disabled(t *testing.T) {
	drop := makeTestDropWithPickit("Ring", item.QualityUnique, "/path/to/rules.nip:42", "[Type] == Ring")
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, false)

	if strings.Contains(embed.Description, "rules.nip") {
		t.Errorf("pickit info should be hidden when disabled, got: %s", embed.Description)
	}
}

func TestBuildItemStashEmbed_PickitInfo_NoRuleFile(t *testing.T) {
	drop := makeTestDropWithPickit("Ring", item.QualityUnique, "", "[Type] == Ring")
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, true)

	if !strings.Contains(embed.Description, "[Type] == Ring") {
		t.Errorf("expected rule text even without file, got: %s", embed.Description)
	}
}

// ---------------------------------------------------------------------------
// buildItemStashEmbed — excluded stat IDs are not shown individually
// ---------------------------------------------------------------------------

func TestBuildItemStashEmbed_ExcludedStats_NotShownIndividually(t *testing.T) {
	// Defense (31) should appear as "Defense: X", not as its stat.String() form
	stats := stat.Stats{
		{ID: 31, Value: 100}, // defense — shown via special handling
		{ID: 48, Value: 5},   // fire min — shown via elemental line
		{ID: 49, Value: 10},  // fire max — shown via elemental line
	}
	drop := makeTestDropWithStats("Armor", item.QualityUnique, stats)
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, false)

	// Defense should be formatted specially
	if !strings.Contains(embed.Description, "Defense: 100") {
		t.Errorf("expected Defense: 100, got: %s", embed.Description)
	}
	// Fire damage should be formatted as elemental line
	if !strings.Contains(embed.Description, "Adds 5-10 Fire Damage") {
		t.Errorf("expected fire damage line, got: %s", embed.Description)
	}
}

// ---------------------------------------------------------------------------
// buildItemStashEmbed — unidentified item has no stat lines
// ---------------------------------------------------------------------------

func TestBuildItemStashEmbed_UnidentifiedItem_NoStats(t *testing.T) {
	drop := makeTestDropWithName("MysteryItem", item.QualityUnique)
	// Not identified, no stats
	evt := event.ItemStashed(event.Text("Koza", "stashed"), drop)
	embed := buildItemStashEmbed(evt, false)

	if strings.Contains(embed.Description, "Defense:") {
		t.Errorf("unidentified item should not show Defense, got: %s", embed.Description)
	}
	if strings.Contains(embed.Description, "All Attributes") {
		t.Errorf("unidentified item should not show All Attributes, got: %s", embed.Description)
	}
}
