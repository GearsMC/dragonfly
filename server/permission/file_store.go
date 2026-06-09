package permission

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FileOperatorStore, operatör kayıtlarını okunabilir bir JSON dosyasında saklar.
// Store küçük ve seyrek değişen bir veri tuttuğu için her değişiklikte atomik dosya yazımı yeterlidir.
type FileOperatorStore struct {
	mu         sync.RWMutex
	path       string
	operator   map[string]Operator
	permission map[string]map[string]State
}

type operatorFile struct {
	Operators   map[string]Operator         `json:"operators"`
	Permissions map[string]map[string]State `json:"permissions,omitempty"`
}

// NewFileOperatorStore, verilen dosyadan operatör kayıtlarını yükler. Dosya yoksa boş dosya oluşturulur.
func NewFileOperatorStore(path string) (*FileOperatorStore, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "operators.json"
	}
	store := &FileOperatorStore{
		path:       path,
		operator:   make(map[string]Operator),
		permission: make(map[string]map[string]State),
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	if err := store.saveLocked(); err != nil {
		return nil, err
	}
	return store, nil
}

// IsOperator, XUID'nin operatör olup olmadığını döndürür.
func (s *FileOperatorStore) IsOperator(xuid string) bool {
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
func (s *FileOperatorStore) Operator(xuid string) (Operator, bool) {
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
func (s *FileOperatorStore) Operators() []Operator {
	store := NewMemoryOperatorStore()
	s.mu.RLock()
	store.operator = make(map[string]Operator, len(s.operator))
	for xuid, operator := range s.operator {
		store.operator[xuid] = operator
	}
	s.mu.RUnlock()
	return store.Operators()
}

// SetOperator, XUID için operatör kaydını ekler, günceller veya siler ve dosyaya yazar.
func (s *FileOperatorStore) SetOperator(xuid, lastKnownName string, value bool) error {
	xuid = normalizeXUID(xuid)
	if xuid == "" {
		return errEmptyXUID()
	}
	lastKnownName = normalizeName(lastKnownName)

	s.mu.Lock()
	defer s.mu.Unlock()
	if !value {
		delete(s.operator, xuid)
		return s.saveLocked()
	}

	now := time.Now()
	operator, ok := s.operator[xuid]
	if !ok {
		operator = Operator{XUID: xuid, CreatedAt: now}
	}
	operator.LastKnownName = lastKnownName
	operator.UpdatedAt = now
	s.operator[xuid] = operator
	return s.saveLocked()
}

// RememberOperatorIdentity, oyuncu zaten operatörse son bilinen adını günceller ve dosyaya yazar.
func (s *FileOperatorStore) RememberOperatorIdentity(xuid, lastKnownName string) error {
	xuid = normalizeXUID(xuid)
	if xuid == "" {
		return nil
	}
	lastKnownName = normalizeName(lastKnownName)

	s.mu.Lock()
	operator, ok := s.operator[xuid]
	if !ok || operator.LastKnownName == lastKnownName {
		s.mu.Unlock()
		return nil
	}
	operator.LastKnownName = lastKnownName
	operator.UpdatedAt = time.Now()
	s.operator[xuid] = operator
	err := s.saveLocked()
	s.mu.Unlock()
	return err
}

// Permission, XUID için açık permission kararını döndürür.
func (s *FileOperatorStore) Permission(xuid, name string) (State, bool) {
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
func (s *FileOperatorStore) Permissions(xuid string) map[string]State {
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
func (s *FileOperatorStore) SetPermission(xuid, name string, state State) error {
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
		return s.saveLocked()
	}
	if s.permission[xuid] == nil {
		s.permission[xuid] = make(map[string]State)
	}
	s.permission[xuid][name] = state
	return s.saveLocked()
}

// Close, dosya store için işlem yapmaz. Her değişiklik anında diske yazılır.
func (s *FileOperatorStore) Close() error {
	return nil
}

func (s *FileOperatorStore) load() error {
	b, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return nil
	}

	var data operatorFile
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	for xuid, operator := range data.Operators {
		xuid = normalizeXUID(xuid)
		if xuid == "" {
			continue
		}
		operator.XUID = xuid
		s.operator[xuid] = operator
	}
	for xuid, permissions := range data.Permissions {
		xuid = normalizeXUID(xuid)
		if xuid == "" {
			continue
		}
		for name, state := range permissions {
			if name == "" || state == Undefined {
				continue
			}
			if s.permission[xuid] == nil {
				s.permission[xuid] = make(map[string]State)
			}
			s.permission[xuid][name] = state
		}
	}
	return nil
}

func (s *FileOperatorStore) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0777); err != nil {
		return err
	}
	data := operatorFile{
		Operators:   make(map[string]Operator, len(s.operator)),
		Permissions: make(map[string]map[string]State, len(s.permission)),
	}
	for xuid, operator := range s.operator {
		data.Operators[xuid] = operator
	}
	for xuid, permissions := range s.permission {
		if len(permissions) == 0 {
			continue
		}
		data.Permissions[xuid] = make(map[string]State, len(permissions))
		for name, state := range permissions {
			if state != Undefined {
				data.Permissions[xuid][name] = state
			}
		}
	}
	if len(data.Permissions) == 0 {
		data.Permissions = nil
	}
	b, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0666); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
