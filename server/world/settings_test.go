package world

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block/cube"
)

func TestSettingsSnapshotCopiesFields(t *testing.T) {
	s := &Settings{
		Name:               "snapshot",
		Spawn:              cube.Pos{1, 2, 3},
		Time:               4,
		TimeCycle:          true,
		RainTime:           5,
		Raining:            true,
		ThunderTime:        6,
		Thundering:         true,
		WeatherCycle:       true,
		RequiredSleepTicks: 7,
		CurrentTick:        8,
		DefaultGameMode:    GameModeCreative,
		Difficulty:         DifficultyHard,
		TickRange:          9,
	}

	snapshot := s.Snapshot()
	if snapshot.Name != s.Name || snapshot.Spawn != s.Spawn || snapshot.Time != s.Time || snapshot.TimeCycle != s.TimeCycle ||
		snapshot.RainTime != s.RainTime || snapshot.Raining != s.Raining || snapshot.ThunderTime != s.ThunderTime ||
		snapshot.Thundering != s.Thundering || snapshot.WeatherCycle != s.WeatherCycle ||
		snapshot.RequiredSleepTicks != s.RequiredSleepTicks || snapshot.CurrentTick != s.CurrentTick ||
		snapshot.DefaultGameMode != s.DefaultGameMode || snapshot.Difficulty != s.Difficulty || snapshot.TickRange != s.TickRange {
		t.Fatalf("snapshot does not match settings")
	}
}
