package permission

import (
	"cmp"
	"slices"
	"sync"
	"time"
)

// MemoryOperatorStore, testler ve özel gömülü kullanımlar için RAM üstünde operatör kayıtları tutar.
type MemoryOperatorStore struct {
	mu         sync.RWMutex
	operator   map[string]Operator
	permission map[string]map[string]State
}

// NewMemoryOperatorStore boş bir bellek içi operatör store'u oluşturur.
func NewMemoryOperatorStore() *MemoryOperatorStore {
	return &MemoryOperatorStore{
		operator:   make(map[string]Operator),
		permission: make(map[string]map[string]State),
	}
}

// IsOperator, XUID'nin operatör olup olmadığını döndürür.
func (s *MemoryOperatorStore) IsOperator(xuid string) bool {
	xuid = normalizeXUID(xuid)
	if xuid == "" {
		return false
	}
	s.mu.RLock()
	_, ok := s.operator[xuid]
	s.mu.RUnlock()
	return ok
}

// Operator, XUID için operatör kaydını döndürür.
func (s *MemoryOperatorStore) Operator(xuid string) (Operator, bool) {
	xuid = normalizeXUID(xuid)
	if xuid == "" {
		return Operator{}, false
	}
	s.mu.RLock()
	operator, ok := s.operator[xuid]
	s.mu.RUnlock()
	return operator, ok
}

// Operators, tüm operatör kayıtlarını XUID'ye göre sıralı döndürür.
func (s *MemoryOperatorStore) Operators() []Operator {
	s.mu.RLock()
	operators := make([]Operator, 0, len(s.operator))
	for _, operator := range s.operator {
		operators = append(operators, operator)
	}
	s.mu.RUnlock()
	slices.SortFunc(operators, func(a, b Operator) int {
		return cmp.Compare(a.XUID, b.XUID)
	})
	return operators
}

// SetOperator, XUID için operatör kaydını ekler, günceller veya siler.
func (s *MemoryOperatorStore) SetOperator(xuid, lastKnownName string, value bool) error {
	xuid = normalizeXUID(xuid)
	if xuid == "" {
		return errEmptyXUID()
	}
	lastKnownName = normalizeName(lastKnownName)

	s.mu.Lock()
	defer s.mu.Unlock()
	if !value {
		delete(s.operator, xuid)
		return nil
	}

	now := time.Now()
	operator, ok := s.operator[xuid]
	if !ok {
		operator = Operator{XUID: xuid, CreatedAt: now}
	}
	operator.LastKnownName = lastKnownName
	operator.UpdatedAt = now
	s.operator[xuid] = operator
	return nil
}

// RememberOperatorIdentity, oyuncu zaten operatörse son bilinen adını günceller.
func (s *MemoryOperatorStore) RememberOperatorIdentity(xuid, lastKnownName string) error {
	xuid = normalizeXUID(xuid)
	if xuid == "" {
		return nil
	}
	lastKnownName = normalizeName(lastKnownName)

	s.mu.Lock()
	operator, ok := s.operator[xuid]
	if ok && operator.LastKnownName != lastKnownName {
		operator.LastKnownName = lastKnownName
		operator.UpdatedAt = time.Now()
		s.operator[xuid] = operator
	}
	s.mu.Unlock()
	return nil
}

// Permission, XUID için açık permission kararını döndürür.
func (s *MemoryOperatorStore) Permission(xuid, name string) (State, bool) {
	xuid = normalizeXUID(xuid)
	if xuid == "" || name == "" {
		return Undefined, false
	}
	s.mu.RLock()
	permissions := s.permission[xuid]
	state, ok := permissions[name]
	s.mu.RUnlock()
	return state, ok
}

// Permissions, XUID için yazılmış açık permission kararlarının kopyasını döndürür.
func (s *MemoryOperatorStore) Permissions(xuid string) map[string]State {
	xuid = normalizeXUID(xuid)
	result := map[string]State{}
	if xuid == "" {
		return result
	}
	s.mu.RLock()
	for name, state := range s.permission[xuid] {
		result[name] = state
	}
	s.mu.RUnlock()
	return result
}

// SetPermission, XUID için açık permission kararı yazar. Undefined verilirse kayıt silinir.
func (s *MemoryOperatorStore) SetPermission(xuid, name string, state State) error {
	xuid = normalizeXUID(xuid)
	if xuid == "" {
		return errEmptyXUID()
	}
	if name == "" {
		return errEmptyPermission()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if state == Undefined {
		delete(s.permission[xuid], name)
		if len(s.permission[xuid]) == 0 {
			delete(s.permission, xuid)
		}
		return nil
	}
	if s.permission[xuid] == nil {
		s.permission[xuid] = make(map[string]State)
	}
	s.permission[xuid][name] = state
	return nil
}

// Close, bellek içi store için işlem yapmaz.
func (s *MemoryOperatorStore) Close() error {
	return nil
}
