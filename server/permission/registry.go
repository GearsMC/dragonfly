package permission

import "sync"

// Registry, permission ağacını tutar. Registry okunurken ve yazılırken eşzamanlı kullanıma uygundur.
type Registry struct {
	mu         sync.RWMutex
	permission map[string]Permission
}

// NewRegistry boş bir permission registry'si oluşturur.
func NewRegistry() *Registry {
	return &Registry{permission: make(map[string]Permission)}
}

// Register, verilen permission düğümünü registry'ye ekler veya mevcut kaydı günceller.
func (r *Registry) Register(permission Permission) {
	if permission.Name == "" {
		return
	}
	if permission.Children == nil {
		permission.Children = map[string]bool{}
	} else {
		children := make(map[string]bool, len(permission.Children))
		for child, value := range permission.Children {
			children[child] = value
		}
		permission.Children = children
	}

	r.mu.Lock()
	r.permission[permission.Name] = permission
	r.mu.Unlock()
}

// Permission, verilen isimdeki kayıtlı permission düğümünü döndürür.
func (r *Registry) Permission(name string) (Permission, bool) {
	r.mu.RLock()
	permission, ok := r.permission[name]
	r.mu.RUnlock()
	if !ok {
		return Permission{}, false
	}
	permission.Children = cloneChildren(permission.Children)
	return permission, true
}

// Permissions, kayıtlı permission düğümlerinin kopyasını döndürür.
func (r *Registry) Permissions() []Permission {
	r.mu.RLock()
	permissions := make([]Permission, 0, len(r.permission))
	for _, permission := range r.permission {
		permission.Children = cloneChildren(permission.Children)
		permissions = append(permissions, permission)
	}
	r.mu.RUnlock()
	return permissions
}

func (r *Registry) permissionUnsafe(name string) (Permission, bool) {
	permission, ok := r.permission[name]
	return permission, ok
}

func cloneChildren(children map[string]bool) map[string]bool {
	if len(children) == 0 {
		return map[string]bool{}
	}
	clone := make(map[string]bool, len(children))
	for child, value := range children {
		clone[child] = value
	}
	return clone
}
