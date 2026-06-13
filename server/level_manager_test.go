package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/df-mc/goleveldb/leveldb"
)

type testProvider struct {
	set *world.Settings
}

func (p *testProvider) Settings() *world.Settings    { return p.set }
func (p *testProvider) SaveSettings(*world.Settings) {}
func (p *testProvider) LoadColumn(world.ChunkPos, world.Dimension) (*chunk.Column, error) {
	return nil, leveldb.ErrNotFound
}
func (p *testProvider) StoreColumn(world.ChunkPos, world.Dimension, *chunk.Column) error { return nil }
func (p *testProvider) Close() error                                                     { return nil }

func TestLoadLevelOpensProviderAtLevelRoot(t *testing.T) {
	worldsFolder := t.TempDir()
	var opened string
	m := newLevelManager(levelManagerConfig{
		worldsFolder: worldsFolder,
		generator:    func(string, world.Dimension) world.Generator { return world.NopGenerator{} },
		openProvider: func(folder string) (world.Provider, error) {
			opened = folder
			return &testProvider{set: &world.Settings{}}, nil
		},
		readOnly: true,
	})

	if _, err := m.LoadLevel("level"); err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	want := filepath.Join(worldsFolder, "level")
	if opened != want {
		t.Fatalf("opened path = %q, want %q", opened, want)
	}
}

func TestLoadLevelUsesLegacyNestedRoot(t *testing.T) {
	worldsFolder := t.TempDir()
	legacyRoot := filepath.Join(worldsFolder, "level", "db")
	if err := os.MkdirAll(filepath.Join(legacyRoot, "db"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyRoot, "level.dat"), []byte{1}, 0644); err != nil {
		t.Fatal(err)
	}

	var opened string
	m := newLevelManager(levelManagerConfig{
		worldsFolder: worldsFolder,
		generator:    func(string, world.Dimension) world.Generator { return world.NopGenerator{} },
		openProvider: func(folder string) (world.Provider, error) {
			opened = folder
			return &testProvider{set: &world.Settings{}}, nil
		},
		readOnly: true,
	})

	if _, err := m.LoadLevel("level"); err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	if opened != legacyRoot {
		t.Fatalf("opened path = %q, want %q", opened, legacyRoot)
	}
}

func TestLoadLevelPrefersRootOverLegacyNestedRoot(t *testing.T) {
	worldsFolder := t.TempDir()
	root := filepath.Join(worldsFolder, "level")
	legacyRoot := filepath.Join(root, "db")
	if err := os.MkdirAll(filepath.Join(legacyRoot, "db"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "level.dat"), []byte{1}, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyRoot, "level.dat"), []byte{1}, 0644); err != nil {
		t.Fatal(err)
	}

	var opened string
	m := newLevelManager(levelManagerConfig{
		worldsFolder: worldsFolder,
		generator:    func(string, world.Dimension) world.Generator { return world.NopGenerator{} },
		openProvider: func(folder string) (world.Provider, error) {
			opened = folder
			return &testProvider{set: &world.Settings{}}, nil
		},
		readOnly: true,
	})

	if _, err := m.LoadLevel("level"); err != nil {
		t.Fatal(err)
	}
	defer m.Close()

	if opened != root {
		t.Fatalf("opened path = %q, want %q", opened, root)
	}
}
