package handlers

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"shape-admin/internal/backup"
	"shape-admin/internal/models"
)

type Handler struct {
	DB     *gorm.DB
	Backup *backup.BackupEngine
	SM     *backup.StorageManager
}

func (h *Handler) GetColors(c *gin.Context) {
	var colors []models.Color
	h.DB.Preload("Shapes").Find(&colors)
	c.JSON(200, colors)
}

func (h *Handler) CreateColor(c *gin.Context) {
	var color models.Color
	if err := c.ShouldBindJSON(&color); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	h.DB.Create(&color)
	c.JSON(201, color)
}

func (h *Handler) GetShapes(c *gin.Context) {
	var shapes []models.Shape
	h.DB.Preload("Color").Find(&shapes)
	c.JSON(200, shapes)
}

func (h *Handler) CreateShape(c *gin.Context) {
	var shape models.Shape
	if err := c.ShouldBindJSON(&shape); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	h.DB.Create(&shape)
	c.JSON(201, shape)
}

func (h *Handler) ListStorages(c *gin.Context) {
	storages := h.SM.GetStorages()
	c.JSON(200, storages)
}

func (h *Handler) AddStorage(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.SM.AddStorage(req.Name, req.Path); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "added"})
}

func (h *Handler) ListBackups(c *gin.Context) {
	storageName := c.Query("storage")
	storage, ok := h.SM.GetStorage(storageName)
	if !ok {
		c.JSON(404, gin.H{"error": "storage not found"})
		return
	}

	entries, _ := os.ReadDir(storage.Path)
	var backups []map[string]string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tar.gz") {
			backups = append(backups, map[string]string{
				"name": entry.Name(),
				"path": filepath.Join(storage.Path, entry.Name()),
			})
		}
	}
	c.JSON(200, backups)
}

func (h *Handler) CreateBackup(c *gin.Context) {
	storageName := c.PostForm("storage")

	sqlDB, _ := h.DB.DB()
	path, err := h.Backup.CreateBackup(c.Request.Context(), sqlDB, storageName)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"path": path, "status": "success"})
}

func (h *Handler) RestoreBackup(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	sqlDB, _ := h.DB.DB()
	if err := h.Backup.RestoreBackup(c.Request.Context(), sqlDB, req.Path); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "restored"})
}

func (h *Handler) DeleteBackup(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.Backup.DeleteBackup(req.Path); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "deleted"})
}