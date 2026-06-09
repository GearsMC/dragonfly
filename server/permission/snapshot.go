package permission

// Snapshot, bir subject için hesaplanmış immutable permission görünümüdür.
// Hot path permission kontrollerinde store veya registry gezmeden bu görünüm okunur.
type Snapshot struct {
	version    uint64
	operator   bool
	permission map[string]State
}

// EmptySnapshot, hiç izin vermeyen boş snapshot döndürür.
func EmptySnapshot() Snapshot {
	return Snapshot{permission: map[string]State{}}
}

// Version, snapshot'ın üretildiği permission manager sürümünü döndürür.
func (s Snapshot) Version() uint64 {
	return s.version
}

// Operator, snapshot sahibinin operatör olup olmadığını döndürür.
func (s Snapshot) Operator() bool {
	return s.operator
}

// Permission, snapshot içinde verilen permission sonucunu döndürür.
// Operatörler açık deny bulunmadığı sürece bilinmeyen permissionları da alır.
func (s Snapshot) Permission(name string) State {
	if name == "" {
		return Undefined
	}
	if state, ok := s.permission[name]; ok {
		return state
	}
	if s.operator {
		return Allow
	}
	return Undefined
}

// Permissions, snapshot içindeki açık hesaplanmış permissionların kopyasını döndürür.
func (s Snapshot) Permissions() map[string]State {
	permissions := make(map[string]State, len(s.permission))
	for name, state := range s.permission {
		permissions[name] = state
	}
	return permissions
}

// SnapshotCalculator, permission sonucunu tek tek hesaplamanın yanında oyuncu başı snapshot üretebilen calculator'dır.
type SnapshotCalculator interface {
	Calculator
	Snapshot(subject Subject) Snapshot
}

// VersionedCalculator, ürettiği snapshotların güncelliğini kontrol etmek için sürüm taşıyan calculator'dır.
type VersionedCalculator interface {
	PermissionVersion() uint64
}
