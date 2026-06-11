package cmd

import (
	"testing"
)

// TestGameModeValues, GameMode enum değerlerini test eder
func TestGameModeValues(t *testing.T) {
	tests := []struct {
		mode GameMode
		name string
	}{
		{GameModeSurvival, "survival"},
		{GameModeCreative, "creative"},
		{GameModeAdventure, "adventure"},
		{GameModeSpectator, "spectator"},
	}

	for _, tt := range tests {
		if tt.mode.String() != tt.name {
			t.Errorf("GameMode %d: expected %s, got %s", tt.mode, tt.name, tt.mode.String())
		}
	}
}

// TestDifficultyValues, Difficulty enum değerlerini test eder
func TestDifficultyValues(t *testing.T) {
	tests := []struct {
		diff Difficulty
		name string
	}{
		{DifficultyPeaceful, "peaceful"},
		{DifficultyEasy, "easy"},
		{DifficultyNormal, "normal"},
		{DifficultyHard, "hard"},
	}

	for _, tt := range tests {
		if tt.diff.String() != tt.name {
			t.Errorf("Difficulty %d: expected %s, got %s", tt.diff, tt.name, tt.diff.String())
		}
	}
}

// TestCoordinateParsing, Coordinate parsing fonksiyonunu test eder
func TestCoordinateParsing(t *testing.T) {
	tests := []struct {
		input string
		mode  CoordinateMode
		value float64
	}{
		{"100", CoordinateAbsolute, 100},
		{"~5", CoordinateRelative, 5},
		{"~-2", CoordinateRelative, -2},
		{"^1", CoordinateCaret, 1},
	}

	for _, tt := range tests {
		coord, err := ParseCoordinate(tt.input)
		if err != nil {
			t.Errorf("ParseCoordinate(%s): %v", tt.input, err)
			continue
		}
		if coord.Mode != tt.mode || coord.Value != tt.value {
			t.Errorf("ParseCoordinate(%s): expected Mode=%d Value=%f, got Mode=%d Value=%f",
				tt.input, tt.mode, tt.value, coord.Mode, coord.Value)
		}
	}
}

// TestPermissionFiltering, output filtering mekanizmasını test eder
func TestPermissionFiltering(t *testing.T) {
	output := &Output{}
	output.Print("Bu mesaj herkese görünür")
	output.SetBroadcastScope(BroadcastConsole)
	output.SetRequiredPermissions("admin.only")

	if output.BroadcastScope() != BroadcastConsole {
		t.Errorf("BroadcastScope: expected %s, got %s", BroadcastConsole, output.BroadcastScope())
	}

	perms := output.RequiredPermissions()
	if len(perms) != 1 || perms[0] != "admin.only" {
		t.Errorf("RequiredPermissions: expected [admin.only], got %v", perms)
	}
}
