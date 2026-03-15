package discordv2

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	d2stat "github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/event"
)

// excludedStatIDs are stat IDs that are handled specially (defense, enhanced
// damage, elemental damage, etc.) and should not appear in the general stat
// listing. Matches v1 exactly.
var excludedStatIDs = map[int]bool{
	17: true, 18: true, // enhanced damage min/max
	21: true, 22: true, 23: true, 24: true, // min/max damage
	31: true,           // defense
	48: true, 49: true, // fire damage
	50: true, 51: true, // lightning damage
	52: true, 53: true, // magic damage
	54: true, 55: true, // cold damage
	57: true, 58: true, // poison damage
	67: true, 68: true, // velocity
	72: true, 73: true, // unused display
	92:  true, // unused display
	118: true, // unused display
	134: true, // unused display
	326: true, // unused display
}

// buildItemStashEmbed constructs a Discord embed for an item stash event with
// full stat formatting matching v1 feature parity.
func buildItemStashEmbed(evt event.ItemStashedEvent, includePickitInfo bool) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Description: buildItemStashDescription(evt, includePickitInfo),
		Color:       qualityColor(evt.Item.Item.Quality.ToString()),
	}
}

// buildItemStashDescription produces the full embed description text for an
// item stash event, including stats, sockets, ethereal, and pickit info.
func buildItemStashDescription(evt event.ItemStashedEvent, includePickitInfo bool) string {
	item := evt.Item.Item
	quality := item.Quality.ToString()
	itemType := item.Desc().Name
	isEthereal := item.Ethereal
	socketCount := len(item.Sockets)
	hasSocketStat := false

	itemName := string(item.Name)
	if item.IdentifiedName != "" {
		itemName = item.IdentifiedName
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("## **%s**\n", itemName))

	switch {
	case itemType != "" && quality != "":
		b.WriteString(fmt.Sprintf("▪️ %s [%s]\n", itemType, quality))
	case itemType != "":
		b.WriteString(fmt.Sprintf("▪️ %s\n", itemType))
	case quality != "":
		b.WriteString(fmt.Sprintf("▪️ [%s]\n", quality))
	}

	if item.Identified && len(item.Stats) > 0 {
		var defense int
		var eMin, eMax int
		var fMin, fMax, lMin, lMax int
		var cMin, cMax, mMin, mMax int
		var pMin, pMax int
		var strVal, energyVal, dexVal, vitVal int
		var hasStr, hasEnergy, hasDex, hasVit bool
		var frVal, crVal, lrVal, prVal int
		var hasFr, hasCr, hasLr, hasPr bool

		for _, s := range item.Stats {
			switch s.ID {
			case d2stat.Strength:
				strVal = s.Value
				hasStr = true
			case d2stat.Energy:
				energyVal = s.Value
				hasEnergy = true
			case d2stat.Dexterity:
				dexVal = s.Value
				hasDex = true
			case d2stat.Vitality:
				vitVal = s.Value
				hasVit = true
			case d2stat.FireResist:
				frVal = s.Value
				hasFr = true
			case d2stat.ColdResist:
				crVal = s.Value
				hasCr = true
			case d2stat.LightningResist:
				lrVal = s.Value
				hasLr = true
			case d2stat.PoisonResist:
				prVal = s.Value
				hasPr = true
			case 31: // defense
				defense = s.Value
			case 17:
				eMin = s.Value
			case 18:
				eMax = s.Value
			case 48:
				fMin = s.Value
			case 49:
				fMax = s.Value
			case 50:
				lMin = s.Value
			case 51:
				lMax = s.Value
			case 52:
				mMin = s.Value
			case 53:
				mMax = s.Value
			case 54:
				cMin = s.Value
			case 55:
				cMax = s.Value
			case 57:
				pMin = s.Value
			case 58:
				pMax = s.Value
			}
		}

		allStatsCombined := hasStr && hasEnergy && hasDex && hasVit &&
			strVal == energyVal && strVal == dexVal && strVal == vitVal
		allResistsPresent := hasFr && hasCr && hasLr && hasPr
		allResistsCombined := allResistsPresent && frVal == crVal && frVal == lrVal && frVal == prVal
		partialResistsCombined, partialResistValue, partialResistID := false, 0, d2stat.ID(0)
		if !allResistsCombined && allResistsPresent {
			partialResistsCombined, partialResistValue, partialResistID = findPartialAllResists(frVal, crVal, lrVal, prVal)
		}

		if defense > 0 {
			b.WriteString(fmt.Sprintf("Defense: %d\n", defense))
		}
		if allStatsCombined && strVal != 0 {
			b.WriteString(fmt.Sprintf("+%d to All Attributes\n", strVal))
		}
		if allResistsCombined && frVal != 0 {
			b.WriteString(fmt.Sprintf("All Resistances %+d\n", frVal))
		} else if partialResistsCombined && partialResistValue != 0 {
			b.WriteString(fmt.Sprintf("All Resistances %+d\n", partialResistValue))
			b.WriteString(fmt.Sprintf("%s %+d\n", resistLabel(partialResistID), resistValueForID(partialResistID, frVal, crVal, lrVal, prVal)))
		}
		if eMin > 0 || eMax > 0 {
			if eMin > 0 && eMax > 0 {
				b.WriteString(fmt.Sprintf("+%d-%d%% Enhanced Damage\n", eMin, eMax))
			} else if eMax > 0 {
				b.WriteString(fmt.Sprintf("+%d%% Enhanced Damage\n", eMax))
			} else {
				b.WriteString(fmt.Sprintf("+%d%% Enhanced Damage\n", eMin))
			}
		}

		b.WriteString(formatDamageLine(fMin, fMax, "Fire"))
		b.WriteString(formatDamageLine(lMin, lMax, "Lightning"))
		b.WriteString(formatDamageLine(cMin, cMax, "Cold"))
		b.WriteString(formatDamageLine(mMin, mMax, "Magic"))
		b.WriteString(formatDamageLine(pMin, pMax, "Poison"))

		for _, s := range item.Stats {
			if allStatsCombined && (s.ID == d2stat.Strength || s.ID == d2stat.Energy || s.ID == d2stat.Dexterity || s.ID == d2stat.Vitality) {
				continue
			}
			if (allResistsCombined || partialResistsCombined) &&
				(s.ID == d2stat.FireResist || s.ID == d2stat.ColdResist || s.ID == d2stat.LightningResist || s.ID == d2stat.PoisonResist) {
				continue
			}
			if excludedStatIDs[int(s.ID)] {
				continue
			}
			statText := s.String()
			if statText != "" {
				if strings.Contains(statText, "Socketed") || strings.Contains(statText, "Sockets") {
					hasSocketStat = true
				}
				b.WriteString(fmt.Sprintf("%s\n", statText))
			}
		}
	}

	if isEthereal {
		b.WriteString("Ethereal\n")
	}
	if socketCount > 0 && !hasSocketStat {
		b.WriteString(fmt.Sprintf("Sockets: %d\n", socketCount))
	}

	if includePickitInfo && (evt.Item.RuleFile != "" || evt.Item.Rule != "") {
		location := formatRuleLocation(evt.Item.RuleFile)
		ruleLine := strings.TrimSpace(evt.Item.Rule)
		b.WriteString("\n")
		if location != "" {
			b.WriteString(fmt.Sprintf("> *%s*\n", location))
		}
		if ruleLine != "" {
			b.WriteString(fmt.Sprintf("> %s\n", ruleLine))
		}
	}

	b.WriteString(fmt.Sprintf("\n`%s | %s`", evt.Supervisor(), time.Now().Format("2006-01-02 15:04:05")))
	return strings.TrimSpace(b.String())
}

// qualityColor returns a Discord embed color for the given item quality string.
func qualityColor(quality string) int {
	switch quality {
	case "LowQuality":
		return 0x666666
	case "Normal":
		return 0xffffff
	case "Superior":
		return 0xc0c0c0
	case "Magic":
		return 0x6969ff
	case "Set":
		return 0x00ff00
	case "Rare":
		return 0xffff77
	case "Unique":
		return 0xbfa969
	case "Crafted":
		return 0xff8000
	default:
		return 0x999999
	}
}

// formatDamageLine formats an elemental damage line. Returns empty string when
// both min and max are zero.
func formatDamageLine(min, max int, damageType string) string {
	if min > 0 || max > 0 {
		return fmt.Sprintf("Adds %d-%d %s Damage\n", min, max, damageType)
	}
	return ""
}

// findPartialAllResists detects when exactly 3 of 4 resist values are equal,
// returning the common value and the ID of the outlier.
func findPartialAllResists(frVal, crVal, lrVal, prVal int) (bool, int, d2stat.ID) {
	values := []int{frVal, crVal, lrVal, prVal}
	counts := map[int]int{}
	for _, v := range values {
		counts[v]++
	}
	for val, count := range counts {
		if count == 3 {
			switch {
			case frVal != val:
				return true, val, d2stat.FireResist
			case crVal != val:
				return true, val, d2stat.ColdResist
			case lrVal != val:
				return true, val, d2stat.LightningResist
			case prVal != val:
				return true, val, d2stat.PoisonResist
			}
		}
	}
	return false, 0, 0
}

// resistLabel returns a human-readable name for a resist stat ID.
func resistLabel(id d2stat.ID) string {
	switch id {
	case d2stat.FireResist:
		return "Fire Resist"
	case d2stat.ColdResist:
		return "Cold Resist"
	case d2stat.LightningResist:
		return "Lightning Resist"
	case d2stat.PoisonResist:
		return "Poison Resist"
	default:
		return "Resist"
	}
}

// resistValueForID returns the resist value for a specific stat ID.
func resistValueForID(id d2stat.ID, frVal, crVal, lrVal, prVal int) int {
	switch id {
	case d2stat.FireResist:
		return frVal
	case d2stat.ColdResist:
		return crVal
	case d2stat.LightningResist:
		return lrVal
	case d2stat.PoisonResist:
		return prVal
	default:
		return 0
	}
}

// formatRuleLocation extracts the base filename and optional line number from
// a pickit rule file path like "/path/to/rules.nip:42".
func formatRuleLocation(ruleFile string) string {
	trimmed := strings.TrimSpace(ruleFile)
	if trimmed == "" {
		return ""
	}

	pathPart := trimmed
	line := ""
	if idx := strings.LastIndex(trimmed, ":"); idx != -1 && idx+1 < len(trimmed) {
		tail := strings.TrimSpace(trimmed[idx+1:])
		if isAllDigits(tail) {
			line = tail
			pathPart = strings.TrimSpace(trimmed[:idx])
		}
	}

	base := filepath.Base(pathPart)
	if base == "." || base == string(filepath.Separator) {
		base = pathPart
	}

	if line != "" {
		return fmt.Sprintf("%s : line %s", base, line)
	}
	return base
}

// isAllDigits returns true if val is non-empty and contains only digits.
func isAllDigits(val string) bool {
	if val == "" {
		return false
	}
	for i := 0; i < len(val); i++ {
		if val[i] < '0' || val[i] > '9' {
			return false
		}
	}
	return true
}
