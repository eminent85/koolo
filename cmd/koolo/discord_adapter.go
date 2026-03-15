package main

import (
	"github.com/hectorgimenez/koolo/internal/bot"
	"github.com/hectorgimenez/koolo/internal/remote/discordv2"
)

// managerAdapter wraps *bot.SupervisorManager to satisfy the
// discordv2.SupervisorControl interface. This bridges the gap between the bot
// package (which returns bot.Stats) and discordv2 (which uses its own
// SupervisorStats type to avoid a transitive Windows-only dependency).
type managerAdapter struct {
	mgr *bot.SupervisorManager
}

func newManagerAdapter(mgr *bot.SupervisorManager) *managerAdapter {
	return &managerAdapter{mgr: mgr}
}

func (a *managerAdapter) AvailableSupervisors() []string {
	return a.mgr.AvailableSupervisors()
}

func (a *managerAdapter) Start(name string, attachToExisting bool, manualMode bool, pidHwnd ...uint32) error {
	return a.mgr.Start(name, attachToExisting, manualMode, pidHwnd...)
}

func (a *managerAdapter) Stop(name string) {
	a.mgr.Stop(name)
}

func (a *managerAdapter) Status(name string) discordv2.SupervisorStats {
	return convertStats(a.mgr.Status(name))
}

func (a *managerAdapter) GetSupervisorStats(name string) discordv2.SupervisorStats {
	return convertStats(a.mgr.GetSupervisorStats(name))
}

// convertStats maps bot.Stats to discordv2.SupervisorStats.
func convertStats(s bot.Stats) discordv2.SupervisorStats {
	return discordv2.SupervisorStats{
		SupervisorStatus: discordv2.SupervisorStatus(s.SupervisorStatus),
		StartedAt:        s.StartedAt,
		Drops:            s.Drops,
		TotalGames:       s.TotalGames(),
		TotalDeaths:      s.TotalDeaths(),
		TotalChickens:    s.TotalChickens(),
		TotalErrors:      s.TotalErrors(),
		Character: discordv2.CharacterInfo{
			Class:           s.UI.Class,
			Level:           s.UI.Level,
			Area:            s.UI.Area,
			Difficulty:      s.UI.Difficulty,
			Life:            s.UI.Life,
			MaxLife:         s.UI.MaxLife,
			Mana:            s.UI.Mana,
			MaxMana:         s.UI.MaxMana,
			MagicFind:       s.UI.MagicFind,
			GoldFind:        s.UI.GoldFind,
			FireResist:      s.UI.FireResist,
			ColdResist:      s.UI.ColdResist,
			LightningResist: s.UI.LightningResist,
			PoisonResist:    s.UI.PoisonResist,
			Ping:            s.UI.Ping,
		},
	}
}
