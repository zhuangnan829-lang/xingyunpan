$content = @'
package property

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/storage"
)

func setupRecyclePropertyTest(t *testing.T) (*gorm.DB, service.RecycleService, repository.UserFileRepository, storage.Storage, string) {
	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        ":memory:",
	}, &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(&model.User{}, &model.UserFile{}, &model.PhysicalFile{}, &model.RecycleBin{})
	require.NoError(t, err)

	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("recycle_test_%d", time.Now().UnixNano()))
	err = os.MkdirAll(tempDir, 0755)
	require.NoError(t, err)

	storageInstance := storage.NewLocalStorage(tempDir)

	recycleRepo := repository.NewRecycleRepository(db)
	userFileRepo := repository.NewUserFileRepository(db)

	recycleService := service.NewRecycleService(recycleRepo, userFileRepo, storageInstance, db)

	return db, recycleService, userFileRepo, storageInstance, tempDir
}

func cleanupRecyclePropertyTest(tempDir string) {
	os.RemoveAll(tempDir)
}

func createRecycleTestUser(db *gorm.DB, userID uint) error {
	user := &model.User{
		BaseModel: model.BaseModel{ID: userID},
		Username:  fmt.Sprintf("user%d", userID),
		Email:     fmt.Sprintf("user%d@test.com", userID),
		Password:  "hashed_password",
	}
	return db.Create(user).Error
}

func createRecycleTestFile(db *gorm.DB, storage storage.Storage, userID uint, fileName string) (*model.UserFile, error) {
	content := []byte(fmt.Sprintf("test content for %s", fileName))
	relativePath := filepath.Join(fmt.Sprintf("user_%d", userID), fileName)
	
	err := storage.Save(bytes.NewReader(content), relativePath)
	if err != nil {
		return nil, err
	}

	physicalFile := &model.PhysicalFile{
		StoragePath: relativePath,
		FileSize:    int64(len(content)),
		FileHash:    "test_hash_" + fileName,
		StorageType: "local",
	}
	if err := db.Create(physicalFile).Error; err != nil {
		return nil, err
	}

	userFile := &model.UserFile{
		UserID:         userID,
		FileName:       fileName,
		FileSize:       int64(len(content)),
		PhysicalFileID: &physicalFile.ID,
		IsFolder:       false,
	}
	if err := db.Create(userFile).Error; err != nil {
		return nil, err
	}

	return userFile, nil
}

// Property 16: Recycle Bin Record Integrity
// **Validates: Requirements 3.1, 3.2**
func TestProperty16_RecycleBinRecordIntegrity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("recycle bin records contain all required fields", prop.ForAll(
		func(userID uint, fileCount int) bool {
			if userID == 0 || fileCount <= 0 || fileCount > 10 {
				return true
			}

			db, recycleService, _, storageInstance, tempDir := setupRecyclePropertyTest(t)
			defer cleanupRecyclePropertyTest(tempDir)

			ctx := context.Background()

			err := createRecycleTestUser(db, userID)
			if err != nil {
				return false
			}

			fileIDs := make([]uint, fileCount)
			for i := 0; i < fileCount; i++ {
				fileName := fmt.Sprintf("file%d.txt", i)
				userFile, err := createRecycleTestFile(db, storageInstance, userID, fileName)
				if err != nil {
					return false
				}
				fileIDs[i] = userFile.ID
			}

			err = recycleService.MoveToRecycle(ctx, userID, fileIDs)
			if err != nil {
				return false
			}

			var recycleItems []model.RecycleBin
			err = db.Where("user_id = ?", userID).Find(&recycleItems).Error
			if err != nil || len(recycleItems) != fileCount {
				return false
			}

			for _, item := range recycleItems {
				if item.UserID != userID || item.FileID == 0 || item.FileName == "" {
					return false
				}
				if item.FileSize <= 0 || item.FileType == "" || item.OriginalPath == "" {
					return false
				}
				if item.FileDeletedAt.IsZero() || item.ExpiresAt.IsZero() {
					return false
				}

				expectedExpiry := item.FileDeletedAt.Add(30 * 24 * time.Hour)
				timeDiff := item.ExpiresAt.Sub(expectedExpiry).Abs()
				if timeDiff > time.Second {
					return false
				}
			}

			return true
		},
		gen.UIntRange(1, 100),
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}
'@

$content | Out-File -FilePath "test\property\recycle_property_test.go" -Encoding UTF8 -NoNewline
Write-Host "File created successfully"
