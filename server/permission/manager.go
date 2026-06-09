package permission

// Manager, permission registry'si ile XUID tabanlı operatör store'unu tek API altında birleştirir.
type Manager struct {
	registry *Registry
	store    Store
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
}

// CalculatePermission, subject için permission sonucunu hesaplar.
func (m *Manager) CalculatePermission(subject Subject, name string) State {
	if name == "" || subject == nil {
		return Undefined
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

	result := Undefined
	m.registry.mu.RLock()
	for _, root := range roots {
		state := m.resolveLocked(root.name, name, root.state, map[string]struct{}{})
		if state == Deny {
			m.registry.mu.RUnlock()
			return Deny
		}
		if state == Allow {
			result = Allow
		}
	}
	_, registered := m.registry.permissionUnsafe(name)
	m.registry.mu.RUnlock()
	if result == Undefined && !registered && operator {
		return Allow
	}
	return result
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
	return m.store.SetOperator(xuid, lastKnownName, value)
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
	return m.store.SetPermission(xuid, name, state)
}

// RememberOperatorIdentity, operatör olan oyuncunun son bilinen adını günceller.
func (m *Manager) RememberOperatorIdentity(xuid, lastKnownName string) error {
	return m.store.RememberOperatorIdentity(xuid, lastKnownName)
}

// Close, permission manager'ın arkasındaki store'u kapatır.
func (m *Manager) Close() error {
	return m.store.Close()
}

type rootPermission struct {
	name  string
	state State
}

func (m *Manager) resolveLocked(current, target string, state State, seen map[string]struct{}) State {
	if current == target {
		return state
	}
	if _, ok := seen[current]; ok {
		return Undefined
	}
	seen[current] = struct{}{}

	permission, ok := m.registry.permissionUnsafe(current)
	if !ok {
		return Undefined
	}

	result := Undefined
	for child, childValue := range permission.Children {
		nextState := stateFromBool(state.Bool() == childValue)
		resolved := m.resolveLocked(child, target, nextState, seen)
		if resolved == Deny {
			return Deny
		}
		if resolved == Allow {
			result = Allow
		}
	}
	return result
}
