package discordv2

import (
	"fmt"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
)

// mockManager implements SupervisorControl for testing.
type mockManager struct {
	supervisors map[string]SupervisorStats
	started     []string
	stopped     []string
}

func newMockManager() *mockManager {
	return &mockManager{
		supervisors: make(map[string]SupervisorStats),
	}
}

func (m *mockManager) addSupervisor(name string, status SupervisorStatus, startedAt time.Time, drops []data.Drop, games, deaths, chickens, errors int) {
	m.supervisors[name] = SupervisorStats{
		SupervisorStatus: status,
		StartedAt:        startedAt,
		Drops:            drops,
		TotalGames:       games,
		TotalDeaths:      deaths,
		TotalChickens:    chickens,
		TotalErrors:      errors,
	}
}

func (m *mockManager) AvailableSupervisors() []string {
	names := make([]string, 0, len(m.supervisors))
	for name := range m.supervisors {
		names = append(names, name)
	}
	return names
}

func (m *mockManager) Start(name string, _ bool, _ bool, _ ...uint32) error {
	if _, ok := m.supervisors[name]; !ok {
		return fmt.Errorf("supervisor %s not found", name)
	}
	m.started = append(m.started, name)
	return nil
}

func (m *mockManager) Stop(name string) {
	m.stopped = append(m.stopped, name)
}

func (m *mockManager) Status(name string) SupervisorStats {
	if s, ok := m.supervisors[name]; ok {
		return s
	}
	return SupervisorStats{}
}

func (m *mockManager) GetSupervisorStats(name string) SupervisorStats {
	return m.Status(name)
}
