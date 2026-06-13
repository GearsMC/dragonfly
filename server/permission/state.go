package permission

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/i18n"
)

// State, bir permission sorgusunun üç durumlu sonucudur.
// Undefined, permission için açık karar verilmediğini belirtir.
type State uint8

const (
	Undefined State = iota
	Deny
	Allow
)

// Bool, yalnızca Allow durumunda true döndürür. Undefined ve Deny güvenli şekilde false kabul edilir.
func (s State) Bool() bool {
	return s == Allow
}

// String, State değerinin dosyada ve logda okunabilir karşılığını döndürür.
func (s State) String() string {
	switch s {
	case Allow:
		return "allow"
	case Deny:
		return "deny"
	default:
		return "undefined"
	}
}

// MarshalText, State değerini JSON içinde okunabilir string olarak yazar.
func (s State) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText, JSON içinden okunan State değerini çözer.
func (s *State) UnmarshalText(text []byte) error {
	switch string(text) {
	case "allow":
		*s = Allow
	case "deny":
		*s = Deny
	case "undefined", "":
		*s = Undefined
	default:
		return fmt.Errorf("%s", i18n.R("%df.internal.permission.unknown_state", fmt.Sprintf("%q", string(text))))
	}
	return nil
}

func stateFromBool(value bool) State {
	if value {
		return Allow
	}
	return Deny
}
