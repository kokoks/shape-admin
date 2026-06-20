package backup

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"math/rand"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type StdLogger struct{}

func (l *StdLogger) Info(msg string)  { fmt.Println("[INFO]", msg) }
func (l *StdLogger) Error(msg string) { fmt.Println("[ERROR]", msg) }

type BackupEngine struct {
	sm     *StorageManager
	logger Logger
}

func NewBackupEngine(sm *StorageManager, logger Logger) *BackupEngine {
	return &BackupEngine{sm: sm, logger: logger}
}

func (be *BackupEngine) CreateBackup(ctx context.Context, db *sql.DB, storageName string) (string, error) {
	storage, ok := be.sm.GetStorage(storageName)
	if !ok {
		storage, _ = be.sm.GetStorage("")
	}

	timestamp := time.Now().Format("20060102_150405")
	randNum := rand.Intn(10000)
	filename := fmt.Sprintf("backup_%s_%04d", timestamp, randNum)
	archivePath := filepath.Join(storage.Path, filename+".tar.gz")

	be.logger.Info(fmt.Sprintf("Starting backup to %s", archivePath))

	var dbName string
	db.QueryRow("SELECT DATABASE()").Scan(&dbName)

	tmpDir := os.TempDir()
	dumpFile := filepath.Join(tmpDir, filename+".sql")

	cmd := exec.CommandContext(ctx, "mysqldump",
		"--single-transaction",
		"--quick",
		"--lock-tables=false",
		dbName,
	)

	dumpOut, err := os.Create(dumpFile)
	if err != nil {
		be.logger.Error(fmt.Sprintf("Failed to create dump file: %v", err))
		return "", err
	}

	cmd.Stdout = dumpOut
	if err := cmd.Run(); err != nil {
		dumpOut.Close()
		os.Remove(dumpFile)
		be.logger.Error(fmt.Sprintf("mysqldump failed: %v", err))
		return "", err
	}
	dumpOut.Close()

	tarCmd := exec.CommandContext(ctx, "tar", "-czf", archivePath, "-C", tmpDir, filename+".sql")
	if err := tarCmd.Run(); err != nil {
		os.Remove(dumpFile)
		be.logger.Error(fmt.Sprintf("tar failed: %v", err))
		return "", err
	}

	os.Remove(dumpFile)
	be.logger.Info(fmt.Sprintf("Backup completed: %s", archivePath))
	return archivePath, nil
}

