package permission

import (
	"os"
	"path/filepath"
	"testing"
)

type testSubject struct {
	xuid string
	name string
}

func (s testSubject) PermissionXUID() string { return s.xuid }
func (s testSubject) PermissionName() string { return s.name }

func TestManagerUserAndOperatorPermissions(t *testing.T) {
	manager := NewManager(NewMemoryOperatorStore())
	user := testSubject{xuid: "1", name: "lexa"}

	if manager.CalculatePermission(user, CommandList) != Allow {
		t.Fatal("varsayılan oyuncu list iznine sahip olmalı")
	}
	if manager.CalculatePermission(user, CommandStop) != Undefined {
		t.Fatal("varsayılan oyuncu stop iznine sahip olmamalı")
	}
	if err := manager.SetOperator(user.xuid, user.name, true); err != nil {
		t.Fatalf("operatör kaydı yazılamadı: %v", err)
	}
	if manager.CalculatePermission(user, CommandStop) != Allow {
		t.Fatal("operatör stop iznine sahip olmalı")
	}
	if manager.CalculatePermission(user, "dfmc.command.custom") != Allow {
		t.Fatal("operatör kayıtlı olmayan özel permission için varsayılan izin almalı")
	}
}

func TestManagerExplicitPermissionOverridesRoots(t *testing.T) {
	manager := NewManager(NewMemoryOperatorStore())
	user := testSubject{xuid: "1", name: "lexa"}
	version := manager.PermissionVersion()
	if err := manager.SetOperator(user.xuid, user.name, true); err != nil {
		t.Fatalf("operatör kaydı yazılamadı: %v", err)
	}
	if manager.PermissionVersion() == version {
		t.Fatal("operatör değişikliği permission sürümünü artırmalı")
	}
	if err := manager.SetPermission(user.xuid, CommandStop, Deny); err != nil {
		t.Fatalf("permission yazılamadı: %v", err)
	}
	if manager.CalculatePermission(user, CommandStop) != Deny {
		t.Fatal("açık deny operatör kök izninden önce gelmeli")
	}
	if err := manager.SetPermission(user.xuid, CommandStop, Undefined); err != nil {
		t.Fatalf("permission silinemedi: %v", err)
	}
	if manager.CalculatePermission(user, CommandStop) != Allow {
		t.Fatal("açık kayıt silinince operatör izni tekrar geçerli olmalı")
	}

	plainUser := testSubject{xuid: "2", name: "oyuncu"}
	if err := manager.SetPermission(plainUser.xuid, GroupOperator, Allow); err != nil {
		t.Fatalf("permission grubu yazılamadı: %v", err)
	}
	if manager.CalculatePermission(plainUser, CommandStop) != Allow {
		t.Fatal("açık grup permission çocuk izinlere yayılmalı")
	}
}

func TestManagerSnapshot(t *testing.T) {
	manager := NewManager(NewMemoryOperatorStore())
	user := testSubject{xuid: "1", name: "lexa"}
	if err := manager.SetOperator(user.xuid, user.name, true); err != nil {
		t.Fatalf("operatör kaydı yazılamadı: %v", err)
	}
	if err := manager.SetPermission(user.xuid, CommandStop, Deny); err != nil {
		t.Fatalf("permission yazılamadı: %v", err)
	}

	snapshot := manager.Snapshot(user)
	if snapshot.Version() != manager.PermissionVersion() {
		t.Fatal("snapshot güncel permission sürümünü taşımalı")
	}
	if !snapshot.Operator() {
		t.Fatal("snapshot operatör durumunu taşımalı")
	}
	if snapshot.Permission(CommandStop) != Deny {
		t.Fatal("snapshot açık deny kararını öncelemeli")
	}
	if snapshot.Permission("dfmc.command.custom") != Allow {
		t.Fatal("operatör snapshot bilinmeyen permission için izin vermeli")
	}
}

func TestManagerDenyChildOverridesAllow(t *testing.T) {
	manager := NewManager(NewMemoryOperatorStore())
	manager.Register(New("dfmc.test.root", "test").WithChildren(map[string]bool{
		"dfmc.test.child": false,
	}))
	manager.Register(New("dfmc.test.child", "test"))
	manager.Register(New(GroupUser, "Varsayılan oyuncu izinleri.").WithChildren(map[string]bool{
		"dfmc.test.root": true,
	}))

	user := testSubject{xuid: "1", name: "lexa"}
	if manager.CalculatePermission(user, "dfmc.test.child") != Deny {
		t.Fatal("false child açık reddetme olarak hesaplanmalı")
	}
}

func TestFileOperatorStorePersistsOperators(t *testing.T) {
	path := filepath.Join(t.TempDir(), "operators.json")
	store, err := NewFileOperatorStore(path)
	if err != nil {
		t.Fatalf("store açılamadı: %v", err)
	}
	if err := store.SetOperator("12345", "lexa", true); err != nil {
		t.Fatalf("operatör yazılamadı: %v", err)
	}
	if err := store.SetPermission("12345", CommandStop, Deny); err != nil {
		t.Fatalf("permission yazılamadı: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("store kapatılamadı: %v", err)
	}

	store, err = NewFileOperatorStore(path)
	if err != nil {
		t.Fatalf("store tekrar açılamadı: %v", err)
	}
	operator, ok := store.Operator("12345")
	if !ok {
		t.Fatal("operatör kaydı kalıcı olmalı")
	}
	if operator.LastKnownName != "lexa" {
		t.Fatalf("son bilinen ad korunmalı, got %q", operator.LastKnownName)
	}
	state, ok := store.Permission("12345", CommandStop)
	if !ok || state != Deny {
		t.Fatalf("permission kalıcı olmalı, got %v %v", state, ok)
	}
	if err := store.RememberOperatorIdentity("12345", "lexaYeni"); err != nil {
		t.Fatalf("operatör adı güncellenemedi: %v", err)
	}
	operator, _ = store.Operator("12345")
	if operator.LastKnownName != "lexaYeni" {
		t.Fatalf("son bilinen ad güncellenmeli, got %q", operator.LastKnownName)
	}
	if err := store.SetOperator("12345", "", false); err != nil {
		t.Fatalf("operatör silinemedi: %v", err)
	}
	if store.IsOperator("12345") {
		t.Fatal("silinen operatör kalmamalı")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("operatör dosyası var olmalı: %v", err)
	}
}
