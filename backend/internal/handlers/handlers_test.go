package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"shape-admin/internal/backup"
	"shape-admin/internal/models"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Color{}, &models.Shape{}, &models.Admin{})
	return db
}

func setupHandler() *Handler {
	db := setupTestDB()
	sm, _ := backup.NewStorageManager("/tmp/test_config.json")
	logger := &backup.StdLogger{}
	engine := backup.NewBackupEngine(sm, logger)
	return &Handler{
		DB:     db,
		Backup: engine,
		SM:     sm,
	}
}

func TestGetColors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupHandler()
	h.DB.Create(&models.Color{Name: "Red", Hex: "#FF0000"})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	h.GetColors(c)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestCreateColor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/colors", nil)
	h.CreateColor(c)

	if w.Code != 400 {
		t.Errorf("Expected 400 for empty body, got %d", w.Code)
	}
}

func TestCreateColorValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupHandler()

	color := models.Color{Name: "Blue", Hex: "#0000FF"}
	body, _ := json.Marshal(color)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/colors", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.CreateColor(c)

	if w.Code != 201 {
		t.Errorf("Expected 201, got %d", w.Code)
	}
}

func TestGetShapes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	h.GetShapes(c)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestCreateShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupHandler()
	h.DB.Create(&models.Color{Name: "Green", Hex: "#00FF00"})

	shape := models.Shape{Name: "Square", Type: "square", ColorID: 1}
	body, _ := json.Marshal(shape)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/shapes", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.CreateShape(c)

	if w.Code != 201 {
		t.Errorf("Expected 201, got %d", w.Code)
	}
}

func TestListStorages(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	h.ListStorages(c)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestAddStorage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupHandler()

	req := map[string]string{"name": "test", "path": "/tmp/test_storage"}
	body, _ := json.Marshal(req)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/storages", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.AddStorage(c)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestAddStorageInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/storages", strings.NewReader("bad json"))
	c.Request.Header.Set("Content-Type", "application/json")
	h.AddStorage(c)

	if w.Code != 400 {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestListBackups(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/backups?storage=", nil)
	h.ListBackups(c)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestListBackupsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/backups?storage=nonexistent", nil)
	h.ListBackups(c)

	if w.Code != 404 {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}

func TestDeleteBackup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupHandler()

	req := map[string]string{"path": "/tmp/nonexistent.tar.gz"}
	body, _ := json.Marshal(req)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("DELETE", "/backups", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.DeleteBackup(c)

	// Ожидаем ошибку, т.к. файла нет
	if w.Code != 500 {
		t.Errorf("Expected 500 for non-existent file, got %d", w.Code)
	}
}

func TestRestoreBackupInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/restore", strings.NewReader("bad"))
	c.Request.Header.Set("Content-Type", "application/json")
	h.RestoreBackup(c)

	if w.Code != 400 {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestCreateBackupHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupHandler()

	form := url.Values{}
	form.Add("storage", "")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/backup", strings.NewReader(form.Encode()))
	c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.CreateBackup(c)
}