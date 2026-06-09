package permission

import (
	"strings"
	"time"
)

// Operator, XUID ile kalıcı olarak yetkilendirilmiş operatör kaydıdır.
// LastKnownName yalnızca okunabilirlik içindir; yetki kararı hiçbir zaman isimden verilmez.
type Operator struct {
	XUID          string    `json:"xuid"`
	LastKnownName string    `json:"lastKnownName,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// OperatorStore, operatör kayıtlarının kalıcı depolama arayüzüdür.
type OperatorStore interface {
	IsOperator(xuid string) bool
	Operator(xuid string) (Operator, bool)
	Operators() []Operator
	SetOperator(xuid, lastKnownName string, value bool) error
	RememberOperatorIdentity(xuid, lastKnownName string) error
	Close() error
}

// PermissionStore, XUID'ye bağlı açık permission kararlarını saklar.
type PermissionStore interface {
	Permission(xuid, name string) (State, bool)
	Permissions(xuid string) map[string]State
	SetPermission(xuid, name string, state State) error
}

// Store, operatör ve açık permission kayıtlarını birlikte tutan kalıcı arayüzdür.
type Store interface {
	OperatorStore
	PermissionStore
}

func normalizeXUID(xuid string) string {
	return strings.TrimSpace(xuid)
}

func normalizeName(name string) string {
	return strings.TrimSpace(name)
}
