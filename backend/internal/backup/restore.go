package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (be *BackupEngine) RestoreBackup(ctx context.Context, db *sql.DB, archivePath string) error {
	be.logger.Info(fmt.Sprintf("Starting restore from %s", archivePath))

	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var dumpFile string
	var dbName string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if strings.HasSuffix(header.Name, ".sql") {
			baseName := filepath.Base(header.Name)
			dbName = strings.TrimSuffix(baseName, ".sql")

			tmpFile := filepath.Join(os.TempDir(), baseName)
			out, _ := os.Create(tmpFile)
			io.Copy(out, tarReader)
			out.Close()
			dumpFile = tmpFile
			break
		}
	}

	if dumpFile == "" {
		return fmt.Errorf("no SQL dump found in archive")
	}
	defer os.Remove(dumpFile)

	var exists int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = ?", dbName).Scan(&exists)
	if err != nil {
		return err
	}

	if exists == 0 {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
		if err != nil {
			be.logger.Error(fmt.Sprintf("Failed to create database: %v", err))
			return fmt.Errorf("database %s does not exist and cannot be created: %w", dbName, err)
		}
		be.logger.Info(fmt.Sprintf("Created database: %s", dbName))
	}

	cmd := exec.CommandContext(ctx, "mysql", dbName)
	dumpIn, _ := os.Open(dumpFile)
	defer dumpIn.Close()
	cmd.Stdin = dumpIn

	if err := cmd.Run(); err != nil {
		be.logger.Error(fmt.Sprintf("Restore failed: %v", err))
		return err
	}

	be.logger.Info(fmt.Sprintf("Restore completed for database: %s", dbName))
	return nil
}