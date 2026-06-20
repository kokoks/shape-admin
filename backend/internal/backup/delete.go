package backup

import (
	"fmt"
	"os"
)

func (be *BackupEngine) DeleteBackup(archivePath string) error {
	be.logger.Info(fmt.Sprintf("Deleting backup: %s", archivePath))

	if err := os.Remove(archivePath); err != nil {
		be.logger.Error(fmt.Sprintf("Failed to delete backup: %v", err))
		return err
	}

	be.logger.Info("Backup deleted successfully")
	return nil
}