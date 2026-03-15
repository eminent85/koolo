package main

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data/difficulty"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/bot"
	"github.com/hectorgimenez/koolo/internal/config"
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
	stats := convertStats(a.mgr.Status(name))
	stats.Character = enrichCharacterInfo(a.mgr, name)
	return stats
}

func (a *managerAdapter) GetSupervisorStats(name string) discordv2.SupervisorStats {
	return convertStats(a.mgr.GetSupervisorStats(name))
}

// convertStats maps bot.Stats to discordv2.SupervisorStats.
// Character data is not mapped here — it is enriched separately by Status()
// via enrichCharacterInfo, which reads live game data.
func convertStats(s bot.Stats) discordv2.SupervisorStats {
	return discordv2.SupervisorStats{
		SupervisorStatus: discordv2.SupervisorStatus(s.SupervisorStatus),
		StartedAt:        s.StartedAt,
		Drops:            s.Drops,
		TotalGames:       s.TotalGames(),
		TotalDeaths:      s.TotalDeaths(),
		TotalChickens:    s.TotalChickens(),
		TotalErrors:      s.TotalErrors(),
	}
}

// enrichCharacterInfo reads live game data for the named supervisor and builds
// a CharacterInfo. Returns a zero value if no game data is available (e.g.
// the supervisor is offline or between games).
func enrichCharacterInfo(mgr *bot.SupervisorManager, name string) discordv2.CharacterInfo {
	d := mgr.GetData(name)
	if d == nil {
		return discordv2.CharacterInfo{}
	}

	var lvl, life, maxLife, mana, maxMana, mf, gf int
	var fr, cr, lr, pr int
	var mfr, mcr, mlr, mpr int

	if v, ok := d.PlayerUnit.FindStat(stat.Level, 0); ok {
		lvl = v.Value
	}
	if v, ok := d.PlayerUnit.FindStat(stat.Life, 0); ok {
		life = v.Value
	}
	if v, ok := d.PlayerUnit.FindStat(stat.MaxLife, 0); ok {
		maxLife = v.Value
	}
	if v, ok := d.PlayerUnit.FindStat(stat.Mana, 0); ok {
		mana = v.Value
	}
	if v, ok := d.PlayerUnit.FindStat(stat.MaxMana, 0); ok {
		maxMana = v.Value
	}
	if v, ok := d.PlayerUnit.FindStat(stat.MagicFind, 0); ok {
		mf = v.Value
	}
	if v, ok := d.PlayerUnit.FindStat(stat.GoldFind, 0); ok {
		gf = v.Value
	}
	if v, ok := d.PlayerUnit.FindStat(stat.FireResist, 0); ok {
		fr = v.Value
	}
	if v, ok := d.PlayerUnit.FindStat(stat.ColdResist, 0); ok {
		cr = v.Value
	}
	if v, ok := d.PlayerUnit.FindStat(stat.LightningResist, 0); ok {
		lr = v.Value
	}
	if v, ok := d.PlayerUnit.FindStat(stat.PoisonResist, 0); ok {
		pr = v.Value
	}
	if v, ok := d.PlayerUnit.FindStat(stat.MaxFireResist, 0); ok {
		mfr = v.Value
	}
	if v, ok := d.PlayerUnit.FindStat(stat.MaxColdResist, 0); ok {
		mcr = v.Value
	}
	if v, ok := d.PlayerUnit.FindStat(stat.MaxLightningResist, 0); ok {
		mlr = v.Value
	}
	if v, ok := d.PlayerUnit.FindStat(stat.MaxPoisonResist, 0); ok {
		mpr = v.Value
	}

	// Apply difficulty resistance penalty and cap.
	penalty := 0
	switch d.CharacterCfg.Game.Difficulty {
	case difficulty.Nightmare:
		penalty = 40
	case difficulty.Hell:
		penalty = 100
	}
	capFR, capCR, capLR, capPR := 75+mfr, 75+mcr, 75+mlr, 75+mpr
	if fr-penalty > capFR {
		fr = capFR
	} else {
		fr = fr - penalty
	}
	if cr-penalty > capCR {
		cr = capCR
	} else {
		cr = cr - penalty
	}
	if lr-penalty > capLR {
		lr = capLR
	} else {
		lr = lr - penalty
	}
	if pr-penalty > capPR {
		pr = capPR
	} else {
		pr = pr - penalty
	}

	diffStr := fmt.Sprint(d.CharacterCfg.Game.Difficulty)
	areaStr := ""
	if areaInfo := d.PlayerUnit.Area.Area(); areaInfo.Name != "" {
		areaStr = areaInfo.Name
	} else {
		areaStr = fmt.Sprint(d.PlayerUnit.Area)
	}

	class := d.CharacterCfg.Character.Class
	if class == "" {
		if cfg, found := config.GetCharacter(name); found {
			class = cfg.Character.Class
		}
	}

	return discordv2.CharacterInfo{
		Class:           class,
		Level:           lvl,
		Area:            areaStr,
		Difficulty:      diffStr,
		Life:            life,
		MaxLife:         maxLife,
		Mana:            mana,
		MaxMana:         maxMana,
		MagicFind:       mf,
		GoldFind:        gf,
		FireResist:      fr,
		ColdResist:      cr,
		LightningResist: lr,
		PoisonResist:    pr,
		Ping:            d.Game.Ping,
	}
}
