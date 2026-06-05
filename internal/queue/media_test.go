package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"testing"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/pkg/storage"
)

func TestGenerateThumbnailForSupportedImages(t *testing.T) {
	cases := []struct {
		name     string
		fileName string
		encode   func(*testing.T) []byte
		format   string
	}{
		{name: "jpg", fileName: "photo.jpg", encode: encodeTestJPEG, format: "jpeg"},
		{name: "png", fileName: "photo.png", encode: encodeTestPNG, format: "png"},
		{name: "gif", fileName: "photo.gif", encode: encodeTestGIF, format: "gif"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			stor := storage.NewLocalStorage(t.TempDir())
			sourcePath := "files/" + tt.fileName
			targetPath := "thumbnails/" + tt.name + ".jpg"
			if err := stor.Save(bytes.NewReader(tt.encode(t)), sourcePath); err != nil {
				t.Fatalf("save source image: %v", err)
			}

			result, err := generateThumbnail(stor, sourcePath, targetPath)
			if err != nil {
				t.Fatalf("generate thumbnail: %v", err)
			}
			if result["status"] != "completed" || result["source_format"] != tt.format {
				t.Fatalf("thumbnail result = %#v", result)
			}

			reader, err := stor.Read(targetPath)
			if err != nil {
				t.Fatalf("read generated thumbnail: %v", err)
			}
			defer reader.Close()
			if _, err := jpeg.Decode(reader); err != nil {
				t.Fatalf("generated thumbnail is not jpeg: %v", err)
			}
		})
	}
}

func TestThumbnailUnsupportedFormatIsSkipped(t *testing.T) {
	if SupportsThumbnail("vector.svg") {
		t.Fatalf("svg should not be a thumbnail candidate")
	}
	if SupportsThumbnail("modern.webp") {
		t.Fatalf("webp should not be a thumbnail candidate")
	}

	payload := FileTaskPayload{
		UserFileID:     1,
		PhysicalFileID: 2,
		FileName:       "vector.svg",
		StoragePath:    "files/vector.svg",
		StorageType:    "local",
	}
	raw, err := EncodePayload(payload)
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	executor := NewExecutor(nil, storage.NewLocalStorage(t.TempDir()), nil, nil, nil, nil)
	result, err := executor.Execute(context.Background(), model.QueueJob{
		JobType: JobTypeFileThumbnail,
		Payload: raw,
	})
	if err != nil {
		t.Fatalf("unsupported thumbnail should be skipped, not failed: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(result), &decoded); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if decoded["status"] != "skipped" || decoded["reason"] != "unsupported_thumbnail_format" || decoded["extension"] != ".svg" {
		t.Fatalf("skip result = %#v", decoded)
	}
}

func encodeTestJPEG(t *testing.T) []byte {
	t.Helper()
	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, testRGBImage(), nil); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return buffer.Bytes()
}

func encodeTestPNG(t *testing.T) []byte {
	t.Helper()
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, testRGBImage()); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buffer.Bytes()
}

func encodeTestGIF(t *testing.T) []byte {
	t.Helper()
	img := image.NewPaletted(image.Rect(0, 0, 8, 6), []color.Color{color.White, color.RGBA{R: 100, G: 120, B: 220, A: 255}})
	for y := 0; y < 6; y++ {
		for x := 0; x < 8; x++ {
			if (x+y)%2 == 0 {
				img.SetColorIndex(x, y, 1)
			}
		}
	}

	var buffer bytes.Buffer
	if err := gif.Encode(&buffer, img, nil); err != nil {
		t.Fatalf("encode gif: %v", err)
	}
	return buffer.Bytes()
}

func testRGBImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 8, 6))
	for y := 0; y < 6; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{R: uint8(20 + x*20), G: uint8(40 + y*20), B: 180, A: 255})
		}
	}
	return img
}
