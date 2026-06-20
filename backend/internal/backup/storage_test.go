package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStorageManager(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.json"

	sm, err := NewStorageManager(configPath)
	if err != nil {
		t.Fatal(err)
	}

	storages := sm.GetStorages()
	if len(storages) != 1 {
		t.Errorf("Expected 1 storage, got %d", len(storages))
	}
}

func TestAddRemoveStorage(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")

	err := sm.AddStorage("manual", tmpDir+"/manual")
	if err != nil {
		t.Fatal(err)
	}

	storages := sm.GetStorages()
	if len(storages) != 2 {
		t.Errorf("Expected 2 storages, got %d", len(storages))
	}

	err = sm.RemoveStorage("manual")
	if err != nil {
		t.Fatal(err)
	}

	storages = sm.GetStorages()
	if len(storages) != 1 {
		t.Errorf("Expected 1 storage after removal, got %d", len(storages))
	}
}

func TestGetStorage(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")

	s, ok := sm.GetStorage("")
	if !ok || s.Name != "default" {
		t.Error("Expected default storage for empty name")
	}
}

func TestStorageManagerSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/storages.json"

	sm1, _ := NewStorageManager(configPath)
	sm1.AddStorage("test", tmpDir+"/test")

	sm2, _ := NewStorageManager(configPath)
	storages := sm2.GetStorages()

	found := false
	for _, s := range storages {
		if s.Name == "test" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to load 'test' storage from config")
	}
}

func TestRemoveDefaultStorage(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")

	err := sm.RemoveStorage("default")
	if err == nil {
		t.Error("Expected error when removing default storage")
	}
}

func TestGetStorageNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")

	_, ok := sm.GetStorage("nonexistent")
	if ok {
		t.Error("Expected false for nonexistent storage")
	}
}

func TestAddStorageInvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")

	err := sm.AddStorage("bad", "/root/badpath/test")
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestStorageManagerPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/persist.json"

	sm, _ := NewStorageManager(configPath)
	sm.AddStorage("persist", tmpDir+"/persist")

	// Проверяем, что файл конфига создан
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file should exist after adding storage")
	}

	// Создаём новый менеджер с тем же конфигом
	smNew, _ := NewStorageManager(configPath)
	s, ok := smNew.GetStorage("persist")
	if !ok {
		t.Error("Expected to find 'persist' storage after reload")
	}
	if s.Path != tmpDir+"/persist" {
		t.Errorf("Expected path %s, got %s", tmpDir+"/persist", s.Path)
	}
}

func TestGetStoragesOrder(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")

	sm.AddStorage("a", tmpDir+"/a")
	sm.AddStorage("b", tmpDir+"/b")

	storages := sm.GetStorages()
	if len(storages) < 3 {
		t.Errorf("Expected at least 3 storages, got %d", len(storages))
	}
}

func TestAddStorageDuplicate(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")

	err := sm.AddStorage("dup", tmpDir+"/dup1")
	if err != nil {
		t.Fatal(err)
	}

	err = sm.AddStorage("dup", tmpDir+"/dup2")
	if err != nil {
		t.Fatal(err)
	}

	s, _ := sm.GetStorage("dup")
	if s.Path != tmpDir+"/dup2" {
		t.Errorf("Expected updated path, got %s", s.Path)
	}
}

func TestRemoveStorageThenGet(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")

	sm.AddStorage("temp", tmpDir+"/temp")
	sm.RemoveStorage("temp")

	_, ok := sm.GetStorage("temp")
	if ok {
		t.Error("Expected storage to be removed")
	}
}

func TestStoragePathAbs(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")

	sm.AddStorage("rel", "relative/path")
	s, _ := sm.GetStorage("rel")

	if !filepath.IsAbs(s.Path) {
		t.Error("Expected absolute path")
	}
}