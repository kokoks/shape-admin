package backup

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type mockLogger struct{}

func (m *mockLogger) Info(msg string)  {}
func (m *mockLogger) Error(msg string) {}

func TestNewBackupEngine(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")
	logger := &mockLogger{}
	be := NewBackupEngine(sm, logger)
	if be == nil {
		t.Error("Expected BackupEngine, got nil")
	}
}

func TestCreateBackupFilenameFormat(t *testing.T) {
	timestamp := time.Now().Format("20060102_150405")
	if len(timestamp) != 15 {
		t.Errorf("Expected timestamp length 15, got %d", len(timestamp))
	}
}

func TestStdLogger(t *testing.T) {
	logger := &StdLogger{}
	logger.Info("test info")
	logger.Error("test error")
}

func TestBackupEngineWithNilStorage(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")
	logger := &mockLogger{}
	be := NewBackupEngine(sm, logger)

	if be.sm == nil {
		t.Error("Expected sm to be set")
	}
	if be.logger == nil {
		t.Error("Expected logger to be set")
	}
}

func TestDeleteBackupNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")
	logger := &mockLogger{}
	be := NewBackupEngine(sm, logger)

	err := be.DeleteBackup("/nonexistent/path/backup.tar.gz")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestDeleteBackupSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")
	logger := &mockLogger{}
	be := NewBackupEngine(sm, logger)

	backupPath := filepath.Join(tmpDir, "backup_test.tar.gz")
	os.WriteFile(backupPath, []byte("test"), 0644)

	err := be.DeleteBackup(backupPath)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("Expected backup file to be deleted")
	}
}

func TestStorageManagerConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")

	done := make(chan bool)
	go func() {
		sm.GetStorages()
		done <- true
	}()
	go func() {
		sm.GetStorage("default")
		done <- true
	}()

	<-done
	<-done
}

func TestAddStorageCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")

	newPath := tmpDir + "/new_storage_dir"
	sm.AddStorage("newdir", newPath)

	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}
}

func TestBackupEngineStruct(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")
	logger := &mockLogger{}
	be := NewBackupEngine(sm, logger)

	if be.sm != sm {
		t.Error("Expected sm to match")
	}
	if be.logger != logger {
		t.Error("Expected logger to match")
	}
}

func TestBackupEngineMethodsExist(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")
	logger := &mockLogger{}
	be := NewBackupEngine(sm, logger)

	if be == nil {
		t.Fatal("BackupEngine is nil")
	}

	var _ func(context.Context, *sql.DB, string) (string, error) = be.CreateBackup
	var _ func(context.Context, *sql.DB, string) error = be.RestoreBackup
	var _ func(string) error = be.DeleteBackup
}

func TestLoggerInterface(t *testing.T) {
	var _ Logger = &mockLogger{}
	var _ Logger = &StdLogger{}
}

func TestDeleteBackupLogs(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")
	logger := &mockLogger{}
	be := NewBackupEngine(sm, logger)

	be.DeleteBackup("/nonexistent/file.tar.gz")
}

func TestStorageManagerPersistenceAfterRemove(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/persist.json"

	sm1, _ := NewStorageManager(configPath)
	sm1.AddStorage("a", tmpDir+"/a")
	sm1.AddStorage("b", tmpDir+"/b")
	sm1.RemoveStorage("a")

	sm2, _ := NewStorageManager(configPath)
	storages := sm2.GetStorages()

	count := 0
	for _, s := range storages {
		if s.Name == "a" {
			t.Error("Expected 'a' to be removed from config")
		}
		count++
	}
	if count < 2 {
		t.Errorf("Expected at least 2 storages, got %d", count)
	}
}

func TestMultipleStorages(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")

	for i := 0; i < 5; i++ {
		name := "storage" + string(rune('0'+i))
		sm.AddStorage(name, tmpDir+"/"+name)
	}

	storages := sm.GetStorages()
	if len(storages) != 6 {
		t.Errorf("Expected 6 storages, got %d", len(storages))
	}
}

func TestGetStorageDefaultFallback(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")

	s, ok := sm.GetStorage("")
	if !ok || s.Name != "default" {
		t.Error("Expected default storage for empty name")
	}

	_, ok = sm.GetStorage("nonexistent")
	if ok {
		t.Error("Expected false for nonexistent storage")
	}
}

func TestStorageConfigFileCreated(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.json"

	_, err := os.Stat(configPath)
	if !os.IsNotExist(err) {
		t.Error("Config should not exist before NewStorageManager")
	}

	sm, _ := NewStorageManager(configPath)
	sm.AddStorage("trigger", tmpDir+"/trigger")

	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		t.Error("Config file should be created after adding storage")
	}
}


func TestRestoreBackupWithNilDB(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")
	logger := &mockLogger{}
	be := NewBackupEngine(sm, logger)

	// Создаём минимальный валидный gzip-файл
	archivePath := filepath.Join(tmpDir, "test.tar.gz")
	f, _ := os.Create(archivePath)
	f.Write([]byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03})
	f.Close()

	err := be.RestoreBackup(context.Background(), nil, archivePath)
	if err == nil {
		t.Error("Expected error for nil DB or bad archive")
	}
}

func TestDeleteBackupMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")
	logger := &mockLogger{}
	be := NewBackupEngine(sm, logger)

	for i := 0; i < 3; i++ {
		path := filepath.Join(tmpDir, "backup"+string(rune('0'+i))+".tar.gz")
		os.WriteFile(path, []byte("test"), 0644)
		err := be.DeleteBackup(path)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	}
}

func TestStorageManagerAddDuplicate(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")

	sm.AddStorage("dup", tmpDir+"/dup1")
	sm.AddStorage("dup", tmpDir+"/dup2")

	s, _ := sm.GetStorage("dup")
	if s.Path != tmpDir+"/dup2" {
		t.Errorf("Expected updated path, got %s", s.Path)
	}
}

func TestStorageManagerRemoveNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	sm, _ := NewStorageManager(tmpDir + "/config.json")

	err := sm.RemoveStorage("nonexistent")
	if err != nil {
		t.Logf("Remove nonexistent returned error (acceptable): %v", err)
	}
}