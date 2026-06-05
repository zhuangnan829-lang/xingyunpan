package controller

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/storage"
)

func TestFileThumbnailEndpointReadsGeneratedThumbnail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, stor := newFileThumbnailTestEnv(t)
	physical := model.PhysicalFile{
		FileHash:    "hash-photo",
		FileSize:    123,
		StoragePath: "files/hash-photo",
		RefCount:    1,
		StorageType: "local",
		ContentType: "image/jpeg",
	}
	if err := db.Create(&physical).Error; err != nil {
		t.Fatalf("seed physical file: %v", err)
	}
	userFile := model.UserFile{
		UserID:         7,
		PhysicalFileID: &physical.ID,
		FileName:       "photo.jpg",
		FileSize:       123,
	}
	if err := db.Create(&userFile).Error; err != nil {
		t.Fatalf("seed user file: %v", err)
	}

	thumbnailBytes := encodeControllerTestJPEG(t)
	if err := stor.Save(bytes.NewReader(thumbnailBytes), "thumbnails/"+strconv.FormatUint(uint64(physical.ID), 10)+".jpg"); err != nil {
		t.Fatalf("save thumbnail: %v", err)
	}

	fileService := service.NewFileService(
		db,
		repository.NewPhysicalFileRepository(db),
		repository.NewUserFileRepository(db),
		nil,
		nil,
		nil,
		stor,
		nil,
		nil,
		nil,
	)
	controller := NewFileController(fileService, nil, nil)
	router := gin.New()
	router.GET("/api/v1/file/:id/thumbnail", func(ctx *gin.Context) {
		ctx.Set("user_id", uint(7))
		controller.Thumbnail(ctx)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/file/"+strconv.FormatUint(uint64(userFile.ID), 10)+"/thumbnail", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d body=%s", rec.Code, rec.Body.String())
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "image/jpeg" {
		t.Fatalf("content type = %q, want image/jpeg", contentType)
	}
	if !bytes.Equal(rec.Body.Bytes(), thumbnailBytes) {
		t.Fatalf("thumbnail bytes mismatch")
	}
}

func newFileThumbnailTestEnv(t *testing.T) (*gorm.DB, *storage.LocalStorage) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "file-thumbnail.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.PhysicalFile{}, &model.UserFile{}); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("open sqlite sql db: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Fatalf("close sqlite db: %v", err)
		}
	})

	return db, storage.NewLocalStorage(t.TempDir())
}

func encodeControllerTestJPEG(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 4, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: uint8(80 + y*30), B: uint8(30 + x*20), A: 255})
		}
	}
	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, img, nil); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return buffer.Bytes()
}
