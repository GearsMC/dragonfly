package world

import (
	"sync/atomic"
	"testing"

	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/df-mc/goleveldb/leveldb"
)

type closeCountingProvider struct {
	set        *Settings
	closeCount atomic.Int32
}

func (p *closeCountingProvider) Settings() *Settings    { return p.set }
func (p *closeCountingProvider) SaveSettings(*Settings) {}
func (p *closeCountingProvider) LoadColumn(ChunkPos, Dimension) (*chunk.Column, error) {
	return nil, leveldb.ErrNotFound
}
func (p *closeCountingProvider) StoreColumn(ChunkPos, Dimension, *chunk.Column) error { return nil }
func (p *closeCountingProvider) Close() error {
	p.closeCount.Add(1)
	return nil
}

func TestLevelSharedProviderClosesAfterAllDimensions(t *testing.T) {
	provider := &closeCountingProvider{set: defaultSettings()}
	l, err := NewLevel(LevelConfig{Name: "world", Provider: provider, ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}

	_ = l.Overworld.Close()
	if got := provider.closeCount.Load(); got != 0 {
		t.Fatalf("close count after overworld close = %v, want 0", got)
	}
	_ = l.Nether.Close()
	if got := provider.closeCount.Load(); got != 0 {
		t.Fatalf("close count after nether close = %v, want 0", got)
	}
	_ = l.End.Close()
	if got := provider.closeCount.Load(); got != 1 {
		t.Fatalf("close count after end close = %v, want 1", got)
	}
}

func TestLevelCloseClosesSharedProviderOnce(t *testing.T) {
	provider := &closeCountingProvider{set: defaultSettings()}
	l, err := NewLevel(LevelConfig{Name: "world", Provider: provider, ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}

	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	if got := provider.closeCount.Load(); got != 1 {
		t.Fatalf("close count = %v, want 1", got)
	}
}

func TestWorldActiveTicksWithoutViewers(t *testing.T) {
	activeSet := defaultSettings()
	activeSet.CurrentTick = 1
	activeProvider := &closeCountingProvider{set: activeSet}
	w := Config{Provider: activeProvider, Active: func() bool { return true }, ReadOnly: true}.New()
	if got := activeSet.CurrentTick; got != 2 {
		t.Fatalf("active world tick = %v, want 2", got)
	}
	_ = w.Close()

	inactiveSet := defaultSettings()
	inactiveSet.CurrentTick = 1
	inactiveProvider := &closeCountingProvider{set: inactiveSet}
	w = Config{Provider: inactiveProvider, ReadOnly: true}.New()
	if got := inactiveSet.CurrentTick; got != 1 {
		t.Fatalf("inactive world tick = %v, want 1", got)
	}
	_ = w.Close()
}
