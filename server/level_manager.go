package server

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/mcdb"
)

// LevelManager, sunucudaki çoklu Level (Overworld+Nether+End) yapısını yönetir.
type LevelManager interface {
	// LoadLevel, worlds klasöründen belirtilen isimdeki Level'i yükler.
	LoadLevel(name string) (*world.Level, error)
	// UnloadLevel, belirtilen Level'i kapatır ve yönetimden çıkarır.
	// Varsayılan level kapatılamaz; level'de oyuncu varsa hata döner.
	UnloadLevel(name string) error
	// Level, isme göre yüklü Level'i döndürür. Bulunamazsa nil döner.
	Level(name string) *world.Level
	// Levels, yüklü tüm Level'leri alfabetik sıralı isim listesi olarak döndürür.
	Levels() []*world.Level
	// DefaultLevel, varsayılan Level'i döndürür.
	DefaultLevel() *world.Level
	// SetDefaultLevel, varsayılan Level'i ayarlar. Level yüklenmiş olmalıdır.
	SetDefaultLevel(name string) error
	// LookupWorld, isim + dimension ile bir dimension World'ü çözer.
	LookupWorld(name string, dim world.Dimension) *world.World
	// LoadLevelsFromDisk, worlds klasöründeki tüm Level'leri yükler.
	LoadLevelsFromDisk() error
	// Close, yönetici altındaki tüm Level'leri kapatır.
	Close() error
}

// levelManagerConfig, defaultLevelManager'ın çalışması için gereken ayarlardır.
type levelManagerConfig struct {
	log                 *slog.Logger
	worldsFolder        string
	defaultLevel        string
	generator           func(levelName string, dim world.Dimension) world.Generator
	openProvider        func(folder string) (world.Provider, error)
	blocks              world.BlockRegistry
	entities            world.EntityRegistry
	randomTickSpeed     int
	saveInterval        time.Duration
	chunkUnloadInterval time.Duration
	readOnly            bool
}

// defaultLevelManager, LevelManager'ın varsayılan implementasyonudur.
type defaultLevelManager struct {
	conf levelManagerConfig

	mu           sync.RWMutex
	levels       map[string]*world.Level
	defaultLevel string
}

// newLevelManager, yeni bir defaultLevelManager oluşturur.
func newLevelManager(conf levelManagerConfig) *defaultLevelManager {
	if conf.log == nil {
		conf.log = slog.Default()
	}
	if conf.worldsFolder == "" {
		conf.worldsFolder = "worlds"
	}
	if conf.openProvider == nil {
		conf.openProvider = defaultProviderOpener(conf.log)
	}
	return &defaultLevelManager{
		conf:         conf,
		levels:       make(map[string]*world.Level),
		defaultLevel: conf.defaultLevel,
	}
}

// defaultProviderOpener, mcdb bazlı varsayılan provider açıcısıdır.
func defaultProviderOpener(log *slog.Logger) func(folder string) (world.Provider, error) {
	return func(folder string) (world.Provider, error) {
		return mcdb.Config{Log: log}.Open(folder)
	}
}

func (m *defaultLevelManager) levelRoot(name string) string {
	root := filepath.Join(m.conf.worldsFolder, name)
	legacyRoot := filepath.Join(root, "db")
	if _, err := os.Stat(filepath.Join(root, "level.dat")); os.IsNotExist(err) {
		if _, err := os.Stat(filepath.Join(legacyRoot, "level.dat")); err == nil {
			if info, err := os.Stat(filepath.Join(legacyRoot, "db")); err == nil && info.IsDir() {
				return legacyRoot
			}
		}
	}
	return root
}

// LoadLevelsFromDisk, worlds klasöründeki tüm alt klasörleri Level olarak yükler.
func (m *defaultLevelManager) LoadLevelsFromDisk() error {
	entries, err := os.ReadDir(m.conf.worldsFolder)
	if os.IsNotExist(err) {
		// Klasör yoksa sadece varsayılan level'i oluştur.
		return m.createDefaultLevel()
	}
	if err != nil {
		return fmt.Errorf("read worlds folder: %w", err)
	}

	var loaded []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if _, err := m.LoadLevel(name); err != nil {
			m.conf.log.Error("failed to load level", "name", name, "err", err)
			continue
		}
		loaded = append(loaded, name)
	}

	if len(loaded) == 0 {
		return m.createDefaultLevel()
	}

	// Varsayılan level seçimi.
	if m.defaultLevel == "" {
		// "world" isimli level varsa onu, yoksa alfabetik ilkini varsayılan yap.
		if _, ok := m.levels["world"]; ok {
			m.defaultLevel = "world"
		} else {
			sort.Strings(loaded)
			m.defaultLevel = loaded[0]
		}
	}
	if _, ok := m.levels[m.defaultLevel]; !ok {
		return fmt.Errorf("default level %q not found", m.defaultLevel)
	}
	return nil
}

// createDefaultLevel, diskte hiç level yoksa "world" adında boş bir level oluşturur.
func (m *defaultLevelManager) createDefaultLevel() error {
	m.conf.log.Info("no worlds found, creating default level", "name", "world")
	if _, err := m.LoadLevel("world"); err != nil {
		return fmt.Errorf("create default level: %w", err)
	}
	if m.defaultLevel == "" {
		m.defaultLevel = "world"
	}
	return nil
}

// LoadLevel, belirtilen isimdeki Level'i yükler. Zaten yüklüyse mevcut olanı döndürür.
func (m *defaultLevelManager) LoadLevel(name string) (*world.Level, error) {
	if name == "" {
		return nil, errors.New("level name cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if l, ok := m.levels[name]; ok {
		return l, nil
	}

	folder := m.levelRoot(name)
	provider, err := m.conf.openProvider(folder)
	if err != nil {
		return nil, fmt.Errorf("open provider for level %q: %w", name, err)
	}

	l, err := world.NewLevel(world.LevelConfig{
		Name:                name,
		Log:                 m.conf.log,
		Provider:            provider,
		Generator:           func(dim world.Dimension) world.Generator { return m.conf.generator(name, dim) },
		Blocks:              m.conf.blocks,
		Entities:            m.conf.entities,
		RandomTickSpeed:     m.conf.randomTickSpeed,
		SaveInterval:        m.conf.saveInterval,
		ChunkUnloadInterval: m.conf.chunkUnloadInterval,
		ReadOnly:            m.conf.readOnly,
	})
	if err != nil {
		_ = provider.Close()
		return nil, fmt.Errorf("create level %q: %w", name, err)
	}

	m.levels[name] = l
	m.conf.log.Info("level loaded", "name", name)
	return l, nil
}

// UnloadLevel, belirtilen Level'i kapatır ve yönetimden çıkarır.
func (m *defaultLevelManager) UnloadLevel(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if name == m.defaultLevel {
		return fmt.Errorf("cannot unload default level %q", name)
	}

	l, ok := m.levels[name]
	if !ok {
		return fmt.Errorf("level %q not loaded", name)
	}

	if err := l.Close(); err != nil {
		return fmt.Errorf("close level %q: %w", name, err)
	}
	delete(m.levels, name)
	m.conf.log.Info("level unloaded", "name", name)
	return nil
}

// Level, isme göre yüklü Level'i döndürür.
func (m *defaultLevelManager) Level(name string) *world.Level {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.levels[name]
}

// Levels, yüklü tüm Level'leri alfabetik sırayla döndürür.
func (m *defaultLevelManager) Levels() []*world.Level {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.levels))
	for name := range m.levels {
		names = append(names, name)
	}
	sort.Strings(names)

	levels := make([]*world.Level, len(names))
	for i, name := range names {
		levels[i] = m.levels[name]
	}
	return levels
}

// DefaultLevel, varsayılan Level'i döndürür.
func (m *defaultLevelManager) DefaultLevel() *world.Level {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.levels[m.defaultLevel]
}

// SetDefaultLevel, varsayılan Level'i ayarlar.
func (m *defaultLevelManager) SetDefaultLevel(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.levels[name]; !ok {
		return fmt.Errorf("level %q is not loaded", name)
	}
	m.defaultLevel = name
	return nil
}

// LookupWorld, isim + dimension ile dimension World'ü çözer. Eski player
// JSON verisindeki level adı silinmiş veya yeniden adlandırılmış olabileceğinden
// çözümleme dayanıklı yapılır: tam eşleşme, ardından büyük/küçük harf
// duyarsız eşleşme denenir; hiçbiri tutmazsa varsayılan level'in istenen
// dimension'ı döner (böylece spawn fallback'i oyuncunun kayıtlı pozisyonunu
// silmez).
func (m *defaultLevelManager) LookupWorld(name string, dim world.Dimension) *world.World {
	m.mu.RLock()
	defer m.mu.RUnlock()

	defaultLevel := func() *world.World {
		l := m.levels[m.defaultLevel]
		if l == nil {
			return nil
		}
		return l.World(dim)
	}

	// name boşsa varsayılan level kullanılır (legacy destek).
	if name == "" {
		return defaultLevel()
	}

	if l, ok := m.levels[name]; ok {
		if w := l.World(dim); w != nil {
			return w
		}
		return defaultLevel()
	}

	// Tam eşleşme yoksa büyük/küçük harf duyarsız eşleşme dene.
	lower := strings.ToLower(name)
	for k, l := range m.levels {
		if strings.ToLower(k) == lower {
			if w := l.World(dim); w != nil {
				return w
			}
			return defaultLevel()
		}
	}

	// Hâlâ bulunamadıysa varsayılan level'in dimension'ını döndür; eski
	// player JSON'u silinmiş/rename edilmiş bir level'a aitse oyuncunun
	// kayıtlı pozisyonu korunur.
	return defaultLevel()
}

// Close, yönetici altındaki tüm Level'leri kapatır.
func (m *defaultLevelManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	for _, l := range m.levels {
		if err := l.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	m.levels = make(map[string]*world.Level)
	return firstErr
}
