package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Storage struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type StorageManager struct {
	mu             sync.RWMutex
	storages       map[string]Storage
	configPath     string
	defaultStorage string
}

func NewStorageManager(configPath string) (*StorageManager, error) {
	sm := &StorageManager{
		storages:       make(map[string]Storage),
		configPath:     configPath,
		defaultStorage: "default",
	}

	defaultPath := "/app/dbBackup"
	os.MkdirAll(defaultPath, 0755)
	sm.storages["default"] = Storage{Name: "default", Path: defaultPath}

	if _, err := os.Stat(configPath); err == nil {
		data, _ := os.ReadFile(configPath)
		json.Unmarshal(data, &sm.storages)
	}

	return sm, nil
}

func (sm *StorageManager) AddStorage(name, path string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	absPath, _ := filepath.Abs(path)
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("cannot create directory: %w", err)
	}

	sm.storages[name] = Storage{Name: name, Path: absPath}
	return sm.saveConfig()
}

func (sm *StorageManager) RemoveStorage(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if name == sm.defaultStorage {
		return fmt.Errorf("cannot remove default storage")
	}

	delete(sm.storages, name)
	return sm.saveConfig()
}

func (sm *StorageManager) GetStorages() []Storage {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]Storage, 0, len(sm.storages))
	for _, s := range sm.storages {
		result = append(result, s)
	}
	return result
}

func (sm *StorageManager) GetStorage(name string) (Storage, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if name == "" {
		s, ok := sm.storages[sm.defaultStorage]
		return s, ok
	}

	s, ok := sm.storages[name]
	return s, ok
}

func (sm *StorageManager) saveConfig() error {
	data, _ := json.MarshalIndent(sm.storages, "", "  ")
	return os.WriteFile(sm.configPath, data, 0644)
}