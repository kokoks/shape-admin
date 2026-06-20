package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"shape-admin/internal/backup"
	"shape-admin/internal/handlers"
	"shape-admin/internal/models"
)

func main() {
	godotenv.Load()

	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME") + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	db.AutoMigrate(&models.Color{}, &models.Shape{}, &models.Admin{})

	var count int64
	db.Model(&models.Admin{}).Where("email = ?", "admin@admin.ru").Count(&count)
	if count == 0 {
		db.Create(&models.Admin{Email: "admin@admin.ru", Password: "111111"})
	}

	sm, _ := backup.NewStorageManager("/app/config/storages.json")
	logger := &backup.StdLogger{}
	engine := backup.NewBackupEngine(sm, logger)

	h := &handlers.Handler{
		DB:     db,
		Backup: engine,
		SM:     sm,
	}

	r := gin.Default()
	r.Static("/static", "./web")
	r.LoadHTMLGlob("web/*")

	api := r.Group("/api")
	{
		api.GET("/colors", h.GetColors)
		api.POST("/colors", h.CreateColor)
		api.GET("/shapes", h.GetShapes)
		api.POST("/shapes", h.CreateShape)
		api.GET("/storages", h.ListStorages)
		api.POST("/storages", h.AddStorage)
		api.GET("/backups", h.ListBackups)
		api.POST("/backup", h.CreateBackup)
		api.POST("/restore", h.RestoreBackup)
		api.DELETE("/backups", h.DeleteBackup)
	}

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.Run(":8080")
}