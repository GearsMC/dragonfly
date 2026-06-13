package world

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"
)

// Level, Allay tarzı bir dünya container'ıdır. İçinde Overworld, Nether ve End
// dimension'ları barındırır. Üç dimension aynı Provider'ı ve Settings'ı paylaşır;
// böylece zaman, hava durumu ve game rules seviye bazında senkronize olur.
type Level struct {
	// Name, level'in diskteki ve yönetimdeki adıdır (örn. "world", "island-42").
	Name string

	// Overworld, level'in overworld dimension'unu temsil eden World'tür.
	Overworld *World

	// Nether, level'in nether dimension'unu temsil eden World'tür.
	Nether *World

	// End, level'in end dimension'unu temsil eden World'tür.
	End *World
}

// World, verilen dimension'a ait World'ü döndürür. Tanınmayan dimension için nil döner.
func (l *Level) World(dim Dimension) *World {
	switch dim {
	case Overworld:
		return l.Overworld
	case Nether:
		return l.Nether
	case End:
		return l.End
	default:
		return nil
	}
}

// Close, level içindeki tüm dimension World'lerini kapatır. Provider/Settings
// ortak kullanıldığı için son kapanan World provider'ı da kapatır.
func (l *Level) Close() error {
	var firstErr error
	for _, w := range []*World{l.Overworld, l.Nether, l.End} {
		if w == nil {
			continue
		}
		if err := w.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (l *Level) hasViewers() bool {
	for _, w := range []*World{l.Overworld, l.Nether, l.End} {
		if w == nil {
			continue
		}
		viewers, _ := w.allViewers()
		if len(viewers) > 0 {
			return true
		}
	}
	return false
}

// LevelConfig, yeni bir Level oluşturmak için gerekli ayarları taşır.
type LevelConfig struct {
	// Name, level'in adıdır.
	Name string

	// Log, level içindeki World'lerin kullanacağı logger. Boş bırakılırsa slog.Default() kullanılır.
	Log *slog.Logger

	// Provider, level'in tüm dimension'ları tarafından paylaşılacak veri sağlayıcısıdır.
	// Boş bırakılırsa NopProvider kullanılır (test/runtime-only senaryolar için).
	Provider Provider

	// Generator, dimension'a göre generator seçen fonksiyondur.
	Generator func(dim Dimension) Generator

	// Blocks, level'deki World'ler tarafından kullanılacak block registry'sidir.
	Blocks BlockRegistry

	// Entities, level'deki World'ler tarafından kullanılacak entity registry'sidir.
	Entities EntityRegistry

	// RandomTickSpeed, her dimension'daki random tick hızıdır. 0 bırakılırsa varsayılan 3 kullanılır.
	RandomTickSpeed int

	// SaveInterval, otomatik kayıt aralığıdır. 0 bırakılırsa varsayılan 10 dakika kullanılır.
	SaveInterval time.Duration

	// ChunkUnloadInterval, kullanılmayan chunk'ların bellekten atılma aralığıdır.
	// 0 veya negatif bırakılırsa varsayılan 2 dakika kullanılır.
	ChunkUnloadInterval time.Duration

	// ReadOnly, level'in diskine yazı yapılmayacağını belirtir.
	ReadOnly bool
}

// NewLevel, verilen config ile yeni bir Level oluşturur. Overworld zorunludur;
// Nether ve End opsiyoneldir. Eğer config'te Provider nil ise, tüm dimension'lar
// için paylaşılan bir NopProvider oluşturulur.
func NewLevel(conf LevelConfig) (*Level, error) {
	if conf.Name == "" {
		return nil, errors.New("level name cannot be empty")
	}
	if conf.Log == nil {
		conf.Log = slog.Default()
	}
	if conf.Generator == nil {
		conf.Generator = func(dim Dimension) Generator { return NopGenerator{} }
	}
	if conf.Provider == nil {
		conf.Provider = NopProvider{Set: defaultSettings()}
	}

	l := &Level{Name: conf.Name}

	overworld, err := l.createDimension(Overworld, conf)
	if err != nil {
		return nil, fmt.Errorf("create overworld for level %q: %w", conf.Name, err)
	}
	// Overworld için spawn pozisyonunun geçerli aralıkta olduğundan emin ol.
	// Diskten gelen level.dat içinde SpawnY math.MaxInt16 olabilir; bu durumda
	// generator'ın varsayılan spawn pozisyonunu kullan. Aksi halde oyuncular
	// geçersiz bir Y değeri nedeniyle siyah ekranda takılır.
	overworld.set.Lock()
	if overworld.set.Spawn[1] > Overworld.Range().Max() || overworld.set.Spawn[1] < math.MinInt32 {
		overworld.set.Spawn = conf.Generator(Overworld).DefaultSpawn(Overworld)
	}
	if conf.Name != "" {
		overworld.set.Name = conf.Name
	}
	overworld.set.Unlock()
	l.Overworld = overworld

	nether, err := l.createDimension(Nether, conf)
	if err != nil {
		l.Close()
		return nil, fmt.Errorf("create nether for level %q: %w", conf.Name, err)
	}
	l.Nether = nether

	end, err := l.createDimension(End, conf)
	if err != nil {
		l.Close()
		return nil, fmt.Errorf("create end for level %q: %w", conf.Name, err)
	}
	l.End = end

	return l, nil
}

// createDimension, tek bir dimension World'ü oluşturur. PortalDestination closure'ı
// level içindeki diğer dimension'lara çözümleme yapar.
func (l *Level) createDimension(dim Dimension, conf LevelConfig) (*World, error) {
	portalDest := func(target Dimension) *World {
		return l.World(target)
	}

	wconf := Config{
		Log:                 conf.Log.With("dimension", fmt.Sprint(dim)),
		Dim:                 dim,
		Provider:            conf.Provider,
		Generator:           conf.Generator(dim),
		Active:              l.hasViewers,
		RandomTickSpeed:     conf.RandomTickSpeed,
		ReadOnly:            conf.ReadOnly,
		SaveInterval:        conf.SaveInterval,
		ChunkUnloadInterval: conf.ChunkUnloadInterval,
		Entities:            conf.Entities,
		Blocks:              conf.Blocks,
		PortalDestination:   portalDest,
	}
	return wconf.New(), nil
}
