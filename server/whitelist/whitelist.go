package whitelist

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

// Entry, XUID tabanlı whitelist girişidir.
// Kimlik doğrulama her zaman XUID üzerinden yapılır, isim sadece okunabilirlik içindir.
type Entry struct {
	XUID          string    `json:"xuid"`
	LastKnownName string    `json:"lastKnownName,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

// Whitelist, XUID tabanlı bağlantı filtreleme sistemi.
// RAM'deki map ile O(1) lookup yapar, Allow() içinde disk I/O yoktur.
type Whitelist struct {
	mu      sync.RWMutex
	enabled bool
	entries map[string]Entry
	path    string
}

// Config, whitelist'in ayarlarını taşır.
type Config struct {
	Enabled bool             `json:"enabled"`
	Entries map[string]Entry `json:"entries"`
}

// New, verilen dosya yolundan whitelist kayıtlarını yükler.
// Dosya yoksa boş whitelist oluşturulur.
func New(path string) (*Whitelist, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "whitelist.json"
	}
	w := &Whitelist{
		enabled: false,
		entries: make(map[string]Entry),
		path:    path,
	}
	if err := w.load(); err != nil {
		return nil, err
	}
	return w, nil
}

// Allow, server.Allower interface'ini uygular.
// Bağlantı sırasında çağrılır, sadece RLock + map lookup yapar.
// Disk I/O yoktur, tahsis yoktur — lag-free.
func (w *Whitelist) Allow(addr net.Addr, d login.IdentityData, _ login.ClientData) (string, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.enabled {
		return "", true
	}

	xuid := strings.TrimSpace(d.XUID)
	if xuid == "" {
		return "XUID bulunamadı, Xbox Live ile giriş yapmalısınız.", false
	}

	if _, ok := w.entries[xuid]; ok {
		return "", true
	}

	return "Whitelist'te değilsiniz.", false
}

// Enabled, whitelist'in etkin olup olmadığını döndürür.
func (w *Whitelist) Enabled() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.enabled
}

// SetEnabled, whitelist'i açar veya kapatır.
func (w *Whitelist) SetEnabled(v bool) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.enabled = v
	return w.saveLocked()
}

// Add, XUID'yi whitelist'e ekler.
func (w *Whitelist) Add(xuid, lastKnownName string) error {
	xuid = strings.TrimSpace(xuid)
	if xuid == "" {
		return errors.New("XUID boş olamaz")
	}
	lastKnownName = strings.TrimSpace(lastKnownName)

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.entries[xuid]; ok {
		return nil
	}
	w.entries[xuid] = Entry{
		XUID:          xuid,
		LastKnownName: lastKnownName,
		CreatedAt:     time.Now(),
	}
	return w.saveLocked()
}

// Remove, XUID'yi whitelist'ten çıkarır.
func (w *Whitelist) Remove(xuid string) error {
	xuid = strings.TrimSpace(xuid)
	if xuid == "" {
		return errors.New("XUID boş olamaz")
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.entries, xuid)
	return w.saveLocked()
}

// Has, XUID'nin whitelist'te olup olmadığını kontrol eder.
func (w *Whitelist) Has(xuid string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	_, ok := w.entries[xuid]
	return ok
}

// Entries, tüm whitelist kayıtlarının kopyasını döndürür.
func (w *Whitelist) Entries() []Entry {
	w.mu.RLock()
	defer w.mu.RUnlock()
	entries := make([]Entry, 0, len(w.entries))
	for _, e := range w.entries {
		entries = append(entries, e)
	}
	return entries
}

// Reload, whitelist'i diskten tekrar yükler.
func (w *Whitelist) Reload() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.entries = make(map[string]Entry)
	return w.load()
}

// ResolveByName, isimle XUID çözmeye çalışır.
// Whitelist kayıtları içinde lastKnownName eşleşmesi arar.
func (w *Whitelist) ResolveByName(name string) (string, bool) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return "", false
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	for xuid, entry := range w.entries {
		if strings.ToLower(entry.LastKnownName) == name {
			return xuid, true
		}
	}
	return "", false
}

func (w *Whitelist) load() error {
	b, err := os.ReadFile(w.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return nil
	}

	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}
	w.enabled = c.Enabled
	if c.Entries != nil {
		w.entries = c.Entries
	}
	return nil
}

func (w *Whitelist) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(w.path), 0777); err != nil {
		return err
	}
	c := Config{
		Enabled: w.enabled,
		Entries: w.entries,
	}
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}
	tmp := w.path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0666); err != nil {
		return err
	}
	return os.Rename(tmp, w.path)
}
