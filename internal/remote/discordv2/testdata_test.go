package discordv2

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
)

// makeTestDrop creates a minimal Drop suitable for testing.
func makeTestDrop() data.Drop {
	return data.Drop{
		Item: data.Item{
			Name:    item.Name("SomeItem"),
			Quality: item.QualityUnique,
		},
	}
}

// makeTestDropWithName creates a Drop with a specific item name and quality.
func makeTestDropWithName(name string, quality item.Quality) data.Drop {
	return data.Drop{
		Item: data.Item{
			Name:    item.Name(name),
			Quality: quality,
		},
	}
}

// makeTestDropIdentified creates a Drop with an IdentifiedName set.
func makeTestDropIdentified(baseName, identifiedName string, quality item.Quality) data.Drop {
	return data.Drop{
		Item: data.Item{
			Name:           item.Name(baseName),
			IdentifiedName: identifiedName,
			Quality:        quality,
			Identified:     true,
		},
	}
}

// makeTestDropWithStats creates an identified Drop with the given stats.
func makeTestDropWithStats(name string, quality item.Quality, stats stat.Stats) data.Drop {
	return data.Drop{
		Item: data.Item{
			Name:       item.Name(name),
			Quality:    quality,
			Identified: true,
			Stats:      stats,
		},
	}
}

// makeTestDropEthereal creates an ethereal Drop.
func makeTestDropEthereal(name string, quality item.Quality) data.Drop {
	return data.Drop{
		Item: data.Item{
			Name:     item.Name(name),
			Quality:  quality,
			Ethereal: true,
		},
	}
}

// makeTestDropWithSockets creates a Drop with the given number of sockets.
func makeTestDropWithSockets(name string, quality item.Quality, numSockets int) data.Drop {
	sockets := make([]data.Item, numSockets)
	for i := range sockets {
		sockets[i] = data.Item{Name: item.Name("Empty")}
	}
	return data.Drop{
		Item: data.Item{
			Name:    item.Name(name),
			Quality: quality,
			Sockets: sockets,
		},
	}
}

// makeTestDropWithPickit creates a Drop with pickit rule info.
func makeTestDropWithPickit(name string, quality item.Quality, ruleFile, rule string) data.Drop {
	return data.Drop{
		Item: data.Item{
			Name:    item.Name(name),
			Quality: quality,
		},
		RuleFile: ruleFile,
		Rule:     rule,
	}
}
