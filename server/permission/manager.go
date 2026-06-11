package permission

import (
	"strings"
	"sync/atomic"
)

// Manager, permission registry'si ile XUID tabanlı operatör store'unu tek API altında birleştirir.
type Manager struct {
	registry *Registry
	store    Store
	version  atomic.Uint64
}

// NewManager, verilen operatör store'u ile varsayılan permission ağacını taşıyan bir manager oluşturur.
func NewManager(store Store) *Manager {
	if store == nil {
		store = NewMemoryOperatorStore()
	}
	registry := NewRegistry()
	RegisterDefaults(registry)
	return &Manager{
		registry: registry,
		store:    store,
	}
}

// Registry, manager'ın permission registry'sini döndürür.
func (m *Manager) Registry() *Registry {
	return m.registry
}

// Register, yeni bir permission düğümünü registry'ye ekler.
func (m *Manager) Register(permission Permission) {
	m.registry.Register(permission)
	m.version.Add(1)
}

// CalculatePermission, subject için permission sonucunu hesaplar.
func (m *Manager) CalculatePermission(subject Subject, name string) State {
	return m.Snapshot(subject).Permission(name)
}

// Snapshot, subject için hesaplanmış permission görünümünü üretir.
func (m *Manager) Snapshot(subject Subject) Snapshot {
	if subject == nil {
		return EmptySnapshot()
	}
	xuid := subject.PermissionXUID()
	operator := m.IsOperator(xuid)
	explicit := m.store.Permissions(xuid)
	roots := make([]rootPermission, 0, len(explicit)+2)
	for permission, state := range explicit {
		if state != Undefined {
			roots = append(roots, rootPermission{name: permission, state: state})
		}
	}
	if operator {
		roots = append(roots, rootPermission{name: GroupOperator, state: Allow})
	}
	roots = append(roots, rootPermission{name: GroupUser, state: Allow})

	permissions := map[string]State{}
	m.registry.mu.RLock()
	for _, root := range roots {
		m.applyLocked(permissions, root.name, root.state, map[string]struct{}{})
	}
	m.registry.mu.RUnlock()
	return Snapshot{
		version:    m.PermissionVersion(),
		operator:   operator,
		permission: permissions,
	}
}

// IsOperator, XUID'nin operatör olup olmadığını döndürür.
func (m *Manager) IsOperator(xuid string) bool {
	return m.store.IsOperator(xuid)
}

// Operator, XUID için operatör kaydını döndürür.
func (m *Manager) Operator(xuid string) (Operator, bool) {
	return m.store.Operator(xuid)
}

// Operators, tüm operatör kayıtlarını döndürür.
func (m *Manager) Operators() []Operator {
	return m.store.Operators()
}

// SetOperator, XUID için operatör yetkisini kalıcı olarak ekler veya kaldırır.
func (m *Manager) SetOperator(xuid, lastKnownName string, value bool) error {
	if err := m.store.SetOperator(xuid, lastKnownName, value); err != nil {
		return err
	}
	m.version.Add(1)
	return nil
}

// Permission, XUID için açık permission kararını döndürür.
func (m *Manager) Permission(xuid, name string) (State, bool) {
	return m.store.Permission(xuid, name)
}

// Permissions, XUID için yazılmış açık permission kararlarını döndürür.
func (m *Manager) Permissions(xuid string) map[string]State {
	return m.store.Permissions(xuid)
}

// SetPermission, XUID için açık permission kararı yazar. Undefined verilirse kayıt silinir.
func (m *Manager) SetPermission(xuid, name string, state State) error {
	if err := m.store.SetPermission(xuid, name, state); err != nil {
		return err
	}
	m.version.Add(1)
	return nil
}

// RememberOperatorIdentity, operatör olan oyuncunun son bilinen adını günceller.
func (m *Manager) RememberOperatorIdentity(xuid, lastKnownName string) error {
	return m.store.RememberOperatorIdentity(xuid, lastKnownName)
}

// Close, permission manager'ın arkasındaki store'u kapatır.
func (m *Manager) Close() error {
	return m.store.Close()
}

// PermissionVersion, permission kararını etkileyen her değişiklikte artan sürümü döndürür.
func (m *Manager) PermissionVersion() uint64 {
	return m.version.Load()
}

type rootPermission struct {
	name  string
	state State
}

func (m *Manager) applyLocked(permissions map[string]State, current string, state State, seen map[string]struct{}) {
	if current == "" || state == Undefined {
		return
	}
	if _, ok := seen[current]; ok {
		return
	}
	seen[current] = struct{}{}
	defer delete(seen, current)

	applyState(permissions, current, state)
	permission, ok := m.registry.permissionUnsafe(current)
	if !ok {
		return
	}
	for child, childValue := range permission.Children {
		nextState := stateFromBool(state.Bool() == childValue)
		m.applyLocked(permissions, child, nextState, seen)
	}
}

func applyState(permissions map[string]State, name string, state State) {
	if state == Deny {
		permissions[name] = Deny
		return
	}
	if _, denied := permissions[name]; denied && permissions[name] == Deny {
		return
	}
	permissions[name] = Allow
}
