package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// GameMode, Minecraft oyun modları temsil eder.
// 0: Survival, 1: Creative, 2: Adventure, 3: Spectator
type GameMode uint8

const (
	GameModeSurvival GameMode = iota
	GameModeCreative
	GameModeAdventure
	GameModeSpectator
)

// String, oyun modunu string'e çevirir.
func (g GameMode) String() string {
	return []string{"survival", "creative", "adventure", "spectator"}[g]
}

// ToWorldGameMode, cmd.GameMode'u world.GameMode'a dönüştürür.
// Dragonfly'nin SetGameMode API'si world.GameMode interface'i beklediği için bu dönüşüm gereklidir.
func (g GameMode) ToWorldGameMode() world.GameMode {
	switch g {
	case GameModeSurvival:
		return world.GameModeSurvival
	case GameModeCreative:
		return world.GameModeCreative
	case GameModeAdventure:
		return world.GameModeAdventure
	case GameModeSpectator:
		return world.GameModeSpectator
	default:
		return world.GameModeSurvival
	}
}

// Type, Enum interface'i için oyun modu türünü döndürür.
func (g GameMode) Type() string {
	return "oyun_modu"
}

// Options, Enum interface'i için mevcut oyun modalarını döndürür.
func (g GameMode) Options(source Source) []string {
	return []string{"0", "1", "2", "3"}
}

// ParseGameMode, string'ten GameMode'a dönüştürür.
func ParseGameMode(src Source, input string) (GameMode, error) {
	switch strings.ToLower(input) {
	case "0", "survival", "s":
		return GameModeSurvival, nil
	case "1", "creative", "c":
		return GameModeCreative, nil
	case "2", "adventure", "a":
		return GameModeAdventure, nil
	case "3", "spectator", "spc", "sp":
		return GameModeSpectator, nil
	}
	return 0, fmt.Errorf("geçersiz oyun modu: %s", input)
}

// Difficulty, Minecraft zorluk seviyeleri
type Difficulty uint8

const (
	DifficultyPeaceful Difficulty = iota
	DifficultyEasy
	DifficultyNormal
	DifficultyHard
)

// String, zorluk seviyesini string'e çevirir.
func (d Difficulty) String() string {
	return []string{"peaceful", "easy", "normal", "hard"}[d]
}

// ToWorldDifficulty, cmd.Difficulty'yi world.Difficulty'ye dönüştürür.
// Dragonfly'nin SetDifficulty API'si world.Difficulty interface'i beklediği için bu dönüşüm gereklidir.
func (d Difficulty) ToWorldDifficulty() world.Difficulty {
	switch d {
	case DifficultyPeaceful:
		return world.DifficultyPeaceful
	case DifficultyEasy:
		return world.DifficultyEasy
	case DifficultyNormal:
		return world.DifficultyNormal
	case DifficultyHard:
		return world.DifficultyHard
	default:
		return world.DifficultyNormal
	}
}

// Type, Enum interface'i için zorluk seviyesi türünü döndürür.
func (d Difficulty) Type() string {
	return "zorluk_seviyesi"
}

// Options, Enum interface'i için mevcut zorluk seviyelerini döndürür.
func (d Difficulty) Options(source Source) []string {
	return []string{"0", "1", "2", "3"}
}

// ParseDifficulty, string'ten Difficulty'ye dönüştürür.
func ParseDifficulty(src Source, input string) (Difficulty, error) {
	switch strings.ToLower(input) {
	case "0", "peaceful", "p":
		return DifficultyPeaceful, nil
	case "1", "easy", "e":
		return DifficultyEasy, nil
	case "2", "normal", "n":
		return DifficultyNormal, nil
	case "3", "hard", "h":
		return DifficultyHard, nil
	}
	return 0, fmt.Errorf("geçersiz zorluk seviyesi: %s", input)
}

// CoordinateMode, koordinat modunun türü
type CoordinateMode uint8

const (
	CoordinateAbsolute CoordinateMode = iota // Mutlak konum: 100 64 200
	CoordinateRelative                       // Göreceli konum: ~5 ~-2 ~10
	CoordinateCaret                          // Caret (ok) konumu: ^1 ^0 ^-2
)

// Coordinate, X/Y/Z koordinatını temsil eder (kayan nokta)
type Coordinate struct {
	Mode  CoordinateMode
	Value float64
}

// ToAbsolute, koordinatı mutlak konuma dönüştürür.
func (c Coordinate) ToAbsolute(senderPos float64) float64 {
	if c.Mode == CoordinateAbsolute {
		return c.Value
	}
	return senderPos + c.Value
}

// String, koordinatı string'e çevirir.
func (c Coordinate) String() string {
	switch c.Mode {
	case CoordinateAbsolute:
		return fmt.Sprintf("%.0f", c.Value)
	case CoordinateRelative:
		if c.Value >= 0 {
			return fmt.Sprintf("~%.0f", c.Value)
		}
		return fmt.Sprintf("~%.0f", c.Value) // "~-5" outputs as expected
	case CoordinateCaret:
		if c.Value >= 0 {
			return fmt.Sprintf("^%.0f", c.Value)
		}
		return fmt.Sprintf("^%.0f", c.Value)
	}
	return ""
}

// ParseCoordinate, string'ten Coordinate'ye dönüştürür.
func ParseCoordinate(input string) (Coordinate, error) {
	if strings.HasPrefix(input, "~") {
		// Göreceli konum
		rest := strings.TrimPrefix(input, "~")
		if rest == "" {
			return Coordinate{Mode: CoordinateRelative, Value: 0}, nil
		}
		val, err := strconv.ParseFloat(rest, 64)
		if err != nil {
			return Coordinate{}, fmt.Errorf("geçersiz göreceli koordinat: %s", input)
		}
		return Coordinate{Mode: CoordinateRelative, Value: val}, nil
	} else if strings.HasPrefix(input, "^") {
		// Caret konumu
		rest := strings.TrimPrefix(input, "^")
		if rest == "" {
			return Coordinate{Mode: CoordinateCaret, Value: 0}, nil
		}
		val, err := strconv.ParseFloat(rest, 64)
		if err != nil {
			return Coordinate{}, fmt.Errorf("geçersiz caret koordinat: %s", input)
		}
		return Coordinate{Mode: CoordinateCaret, Value: val}, nil
	} else {
		// Mutlak konum
		val, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return Coordinate{}, fmt.Errorf("geçersiz mutlak koordinat: %s", input)
		}
		return Coordinate{Mode: CoordinateAbsolute, Value: val}, nil
	}
}

// Position3D, 3D konumu temsil eder
// X, Y, Z koordinatlarından oluşur ve göreceli/mutlak/caret modlarını destekler
type Position3D struct {
	X Coordinate
	Y Coordinate
	Z Coordinate
}

// ToVec3, Position3D'yi mutlak Vec3'e dönüştürür.
func (p Position3D) ToVec3(senderPos mgl64.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{
		p.X.ToAbsolute(senderPos.X()),
		p.Y.ToAbsolute(senderPos.Y()),
		p.Z.ToAbsolute(senderPos.Z()),
	}
}

// String, konum bilgisini string'e çevirir.
func (p Position3D) String() string {
	return fmt.Sprintf("%s %s %s", p.X.String(), p.Y.String(), p.Z.String())
}

// IntRange, Min..Max tarzında sayı aralığı temsil eder
type IntRange struct {
	Min int32
	Max int32
}

// InRange, sayının aralık içinde olup olmadığını kontrol eder.
func (r IntRange) InRange(value int32) bool {
	return value >= r.Min && value <= r.Max
}

// String, aralığı string'e çevirir.
func (r IntRange) String() string {
	if r.Min == r.Max {
		return fmt.Sprintf("%d", r.Min)
	}
	return fmt.Sprintf("%d..%d", r.Min, r.Max)
}

// ParseIntRange, string'ten IntRange'e dönüştürür.
func ParseIntRange(input string) (IntRange, error) {
	if strings.Contains(input, "..") {
		parts := strings.Split(input, "..")
		if len(parts) != 2 {
			return IntRange{}, fmt.Errorf("geçersiz aralık formatı: %s", input)
		}
		min, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 32)
		if err != nil {
			return IntRange{}, fmt.Errorf("minimum sayı geçersiz: %s", parts[0])
		}
		max, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 32)
		if err != nil {
			return IntRange{}, fmt.Errorf("maksimum sayı geçersiz: %s", parts[1])
		}
		return IntRange{Min: int32(min), Max: int32(max)}, nil
	}

	val, err := strconv.ParseInt(input, 10, 32)
	if err != nil {
		return IntRange{}, fmt.Errorf("geçersiz sayı: %s", input)
	}
	intVal := int32(val)
	return IntRange{Min: intVal, Max: intVal}, nil
}
