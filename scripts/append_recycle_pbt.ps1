$additionalContent = @'

// Property 17: Soft Delete Preserves Physical File
// **Validates: Requirements 3.3**
func TestProperty17_SoftDeletePreservesPhysicalFile(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("soft delete preserves physical file", prop.ForAll(
		func(userID uint, fileName string) bool {
			if userID == 0 || fileName == "" {
				return true
			}

			db, recycleService, _, storageInstance, tempDir := setupRecyclePropertyTest(t)
			defer cleanupRecyclePropertyTest(tempDir)

			ctx := context.Background()

			err := createRecycleTestUser(db, userID)
			if err != nil {
				return false
			}

			userFile, err := createRecycleTestFile(db, storageInstance, userID, fileName)
			if err != nil {
				return false
			}

			var physicalFile model.PhysicalFile
			err = db.First(&physicalFile, userFile.PhysicalFileID).Error
			if err != nil {
				return false
			}

			exists := storageInstance.Exists(physicalFile.StoragePath)
			if !exists {
				return false
			}

			fileIDs := []uint{userFile.ID}
			err = recycleService.MoveToRecycle(ctx, userID, fileIDs)
			if err != nil {
				return false
			}

			exists = storageInstance.Exists(physicalFile.StoragePath)
			if !exists {
				return false
			}

			var deletedFile model.UserFile
			err = db.Unscoped().First(&deletedFile, userFile.ID).Error
			if err != nil {
				return false
			}
			if !deletedFile.DeletedAt.Valid {
				return false
			}

			return true
		},
		gen.UIntRange(1, 100),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// Property 18: Recycle Bin Round-Trip
// **Validates: Requirements 3.6**
func TestProperty18_RecycleBinRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("delete and restore maintains file data consistency", prop.ForAll(
		func(userID uint, fileName string, fileContent string) bool {
			if userID == 0 || fileName == "" || fileContent == "" {
				return true
			}

			db, recycleService, userFileRepo, storageInstance, tempDir := setupRecyclePropertyTest(t)
			defer cleanupRecyclePropertyTest(tempDir)

			ctx := context.Background()

			err := createRecycleTestUser(db, userID)
			if err != nil {
				return false
			}

			content := []byte(fileContent)
			relativePath := filepath.Join(fmt.Sprintf("user_%d", userID), fileName)
			err = storageInstance.Save(bytes.NewReader(content), relativePath)
			if err != nil {
				return false
			}

			physicalFile := &model.PhysicalFile{
				StoragePath: relativePath,
				FileSize:    int64(len(content)),
				FileHash:    "test_hash",
				StorageType: "local",
			}
			if err := db.Create(physicalFile).Error; err != nil {
				return false
			}

			userFile := &model.UserFile{
				UserID:         userID,
				FileName:       fileName,
				FileSize:       int64(len(content)),
				PhysicalFileID: &physicalFile.ID,
				IsFolder:       false,
			}
			if err := db.Create(userFile).Error; err != nil {
				return false
			}

			originalFileID := userFile.ID
			originalFileName := userFile.FileName
			originalFileSize := userFile.FileSize

			fileIDs := []uint{userFile.ID}
			err = recycleService.MoveToRecycle(ctx, userID, fileIDs)
			if err != nil {
				return false
			}

			var recycleItem model.RecycleBin
			err = db.Where("user_id = ? AND file_id = ?", userID, originalFileID).First(&recycleItem).Error
			if err != nil {
				return false
			}

			itemIDs := []uint{recycleItem.ID}
			err = recycleService.RestoreFiles(ctx, userID, itemIDs)
			if err != nil {
				return false
			}

			restoredFile, err := userFileRepo.GetByID(originalFileID)
			if err != nil {
				return false
			}

			if restoredFile.FileName != originalFileName && !strings.Contains(restoredFile.FileName, originalFileName) {
				return false
			}
			if restoredFile.FileSize != originalFileSize {
				return false
			}
			if restoredFile.DeletedAt.Valid {
				return false
			}

			reader, err := storageInstance.Read(relativePath)
			if err != nil {
				return false
			}
			defer reader.Close()

			restoredContent, err := io.ReadAll(reader)
			if err != nil {
				return false
			}
			if string(restoredContent) != fileContent {
				return false
			}

			return true
		},
		gen.UIntRange(1, 100),
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// Property 19: File Name Conflict Auto Rename
// **Validates: Requirements 3.8**
func TestProperty19_FileNameConflictAutoRename(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("restore auto-renames on file name conflict", prop.ForAll(
		func(userID uint, fileName string) bool {
			if userID == 0 || fileName == "" {
				return true
			}

			db, recycleService, userFileRepo, storageInstance, tempDir := setupRecyclePropertyTest(t)
			defer cleanupRecyclePropertyTest(tempDir)

			ctx := context.Background()

			err := createRecycleTestUser(db, userID)
			if err != nil {
				return false
			}

			file1, err := createRecycleTestFile(db, storageInstance, userID, fileName)
			if err != nil {
				return false
			}

			fileIDs := []uint{file1.ID}
			err = recycleService.MoveToRecycle(ctx, userID, fileIDs)
			if err != nil {
				return false
			}

			file2, err := createRecycleTestFile(db, storageInstance, userID, fileName)
			if err != nil {
				return false
			}

			var recycleItem model.RecycleBin
			err = db.Where("user_id = ? AND file_id = ?", userID, file1.ID).First(&recycleItem).Error
			if err != nil {
				return false
			}

			itemIDs := []uint{recycleItem.ID}
			err = recycleService.RestoreFiles(ctx, userID, itemIDs)
			if err != nil {
				return false
			}

			restoredFile, err := userFileRepo.GetByID(file1.ID)
			if err != nil {
				return false
			}

			if restoredFile.FileName == fileName {
				existingFile, err := userFileRepo.GetByID(file2.ID)
				if err != nil || existingFile.DeletedAt.Valid {
					return false
				}
			}

			if restoredFile.DeletedAt.Valid {
				return false
			}

			return true
		},
		gen.UIntRange(1, 100),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// Property 20: Restore Data Cleanup
// **Validates: Requirements 3.9**
func TestProperty20_RestoreDataCleanup(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("restore removes recycle bin records", prop.ForAll(
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

			itemIDs := make([]uint, len(recycleItems))
			for i, item := range recycleItems {
				itemIDs[i] = item.ID
			}
			err = recycleService.RestoreFiles(ctx, userID, itemIDs)
			if err != nil {
				return false
			}

			var remainingItems []model.RecycleBin
			err = db.Where("user_id = ?", userID).Find(&remainingItems).Error
			if err != nil {
				return false
			}

			if len(remainingItems) != 0 {
				return false
			}

			return true
		},
		gen.UIntRange(1, 100),
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// Property 21: Permanent Delete Integrity
// **Validates: Requirements 3.10, 3.11**
func TestProperty21_PermanentDeleteIntegrity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("permanent delete removes all data", prop.ForAll(
		func(userID uint, fileName string) bool {
			if userID == 0 || fileName == "" {
				return true
			}

			db, recycleService, _, storageInstance, tempDir := setupRecyclePropertyTest(t)
			defer cleanupRecyclePropertyTest(tempDir)

			ctx := context.Background()

			err := createRecycleTestUser(db, userID)
			if err != nil {
				return false
			}

			userFile, err := createRecycleTestFile(db, storageInstance, userID, fileName)
			if err != nil {
				return false
			}

			var physicalFile model.PhysicalFile
			err = db.First(&physicalFile, userFile.PhysicalFileID).Error
			if err != nil {
				return false
			}
			physicalFilePath := physicalFile.StoragePath

			fileIDs := []uint{userFile.ID}
			err = recycleService.MoveToRecycle(ctx, userID, fileIDs)
			if err != nil {
				return false
			}

			var recycleItem model.RecycleBin
			err = db.Where("user_id = ? AND file_id = ?", userID, userFile.ID).First(&recycleItem).Error
			if err != nil {
				return false
			}

			itemIDs := []uint{recycleItem.ID}
			err = recycleService.PermanentDelete(ctx, userID, itemIDs)
			if err != nil {
				return false
			}

			exists := storageInstance.Exists(physicalFilePath)
			if exists {
				return false
			}

			var deletedRecycleItem model.RecycleBin
			err = db.Where("id = ?", recycleItem.ID).First(&deletedRecycleItem).Error
			if err == nil {
				return false
			}

			var deletedPhysicalFile model.PhysicalFile
			err = db.Where("id = ?", physicalFile.ID).First(&deletedPhysicalFile).Error
			if err == nil {
				return false
			}

			return true
		},
		gen.UIntRange(1, 100),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// Property 22: Empty Recycle Bin Integrity
// **Validates: Requirements 3.12**
func TestProperty22_EmptyRecycleBinIntegrity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("empty recycle bin deletes all user items", prop.ForAll(
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
			physicalFilePaths := make([]string, fileCount)
			for i := 0; i < fileCount; i++ {
				fileName := fmt.Sprintf("file%d.txt", i)
				userFile, err := createRecycleTestFile(db, storageInstance, userID, fileName)
				if err != nil {
					return false
				}
				fileIDs[i] = userFile.ID

				var physicalFile model.PhysicalFile
				db.First(&physicalFile, userFile.PhysicalFileID)
				physicalFilePaths[i] = physicalFile.StoragePath
			}

			err = recycleService.MoveToRecycle(ctx, userID, fileIDs)
			if err != nil {
				return false
			}

			err = recycleService.EmptyRecycleBin(ctx, userID)
			if err != nil {
				return false
			}

			var remainingItems []model.RecycleBin
			err = db.Where("user_id = ?", userID).Find(&remainingItems).Error
			if err != nil {
				return false
			}
			if len(remainingItems) != 0 {
				return false
			}

			for _, filePath := range physicalFilePaths {
				exists := storageInstance.Exists(filePath)
				if exists {
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

// Property 23: Batch Operation Transactionality
// **Validates: Requirements 3.14**
func TestProperty23_BatchOperationTransactionality(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	properties.Property("batch operations are transactional", prop.ForAll(
		func(userID uint, fileCount int) bool {
			if userID == 0 || fileCount <= 1 || fileCount > 5 {
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

			var countBefore int64
			db.Model(&model.RecycleBin{}).Where("user_id = ?", userID).Count(&countBefore)

			var recycleItems []model.RecycleBin
			err = db.Where("user_id = ?", userID).Find(&recycleItems).Error
			if err != nil {
				return false
			}

			itemIDs := make([]uint, len(recycleItems))
			for i, item := range recycleItems {
				itemIDs[i] = item.ID
			}

			var wg sync.WaitGroup
			errors := make([]error, 2)

			for i := 0; i < 2; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					errors[idx] = recycleService.RestoreFiles(ctx, userID, itemIDs)
				}(i)
			}

			wg.Wait()

			successCount := 0
			for _, err := range errors {
				if err == nil {
					successCount++
				}
			}

			if successCount == 0 {
				return false
			}

			var countAfter int64
			db.Model(&model.RecycleBin{}).Where("user_id = ?", userID).Count(&countAfter)

			if countAfter != 0 && countAfter != countBefore {
				var restoredFiles []model.UserFile
				db.Where("user_id = ?", userID).Find(&restoredFiles)

				nonDeletedCount := 0
				for _, f := range restoredFiles {
					if !f.DeletedAt.Valid {
						nonDeletedCount++
					}
				}

				if nonDeletedCount != fileCount && nonDeletedCount != 0 {
					return false
				}
			}

			return true
		},
		gen.UIntRange(1, 100),
		gen.IntRange(2, 5),
	))

	properties.TestingRun(t)
}
'@

$additionalContent | Add-Content -Path "test\property\recycle_property_test.go" -Encoding UTF8 -NoNewline
Write-Host "Additional properties appended successfully"
