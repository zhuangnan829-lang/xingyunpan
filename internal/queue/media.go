package queue

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"xingyunpan-v2/pkg/mimetype"
	"xingyunpan-v2/pkg/storage"
)

const thumbnailMaxEdge = 360

type mediaMetadata struct {
	ContentType string `json:"content_type"`
	Extension   string `json:"extension"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	IsImage     bool   `json:"is_image"`
}

func extractMediaMetadata(stor storage.Storage, storagePath, fileName string) (mediaMetadata, error) {
	reader, err := stor.Read(storagePath)
	if err != nil {
		return mediaMetadata{}, fmt.Errorf("read source file failed: %w", err)
	}
	defer reader.Close()

	header := make([]byte, 512)
	n, err := io.ReadFull(reader, header)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return mediaMetadata{}, fmt.Errorf("read source header failed: %w", err)
	}
	header = header[:n]

	extension := strings.ToLower(filepath.Ext(fileName))
	contentType := http.DetectContentType(header)
	if mimetype.IsGeneric(contentType) {
		contentType = mimetype.FromFileName(fileName)
	}

	result := mediaMetadata{
		ContentType: contentType,
		Extension:   extension,
	}

	configReader, err := stor.Read(storagePath)
	if err != nil {
		return result, nil
	}
	defer configReader.Close()

	cfg, format, err := image.DecodeConfig(configReader)
	if err == nil {
		result.Width = cfg.Width
		result.Height = cfg.Height
		result.IsImage = isSupportedImageFormat(format)
		if result.ContentType == "application/octet-stream" {
			result.ContentType = imageFormatToContentType(format)
		}
	}

	return result, nil
}

func generateThumbnail(stor storage.MultipartStorage, sourcePath, targetPath string) (map[string]interface{}, error) {
	reader, err := stor.Read(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("read thumbnail source failed: %w", err)
	}
	defer reader.Close()

	img, format, err := decodeSupportedImage(reader)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid source image dimensions")
	}

	dstWidth, dstHeight := fitIntoBox(width, height, thumbnailMaxEdge)
	resized := resizeImage(img, dstWidth, dstHeight)

	buffer := bytes.NewBuffer(nil)
	if err := jpeg.Encode(buffer, resized, &jpeg.Options{Quality: 82}); err != nil {
		return nil, fmt.Errorf("encode thumbnail failed: %w", err)
	}

	if err := stor.Save(bytes.NewReader(buffer.Bytes()), targetPath); err != nil {
		return nil, fmt.Errorf("save thumbnail failed: %w", err)
	}

	return map[string]interface{}{
		"status":            "completed",
		"source_format":     format,
		"thumbnail_path":    targetPath,
		"thumbnail_width":   dstWidth,
		"thumbnail_height":  dstHeight,
		"thumbnail_quality": 82,
	}, nil
}

func decodeSupportedImage(reader io.Reader) (image.Image, string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", fmt.Errorf("read image bytes failed: %w", err)
	}

	if img, err := jpeg.Decode(bytes.NewReader(data)); err == nil {
		return img, "jpeg", nil
	}
	if img, err := png.Decode(bytes.NewReader(data)); err == nil {
		return img, "png", nil
	}
	if img, err := gif.Decode(bytes.NewReader(data)); err == nil {
		return img, "gif", nil
	}

	return nil, "", fmt.Errorf("unsupported thumbnail source format")
}

func resizeImage(src image.Image, dstWidth, dstHeight int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	srcBounds := src.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	for y := 0; y < dstHeight; y++ {
		srcY := srcBounds.Min.Y + y*srcHeight/dstHeight
		for x := 0; x < dstWidth; x++ {
			srcX := srcBounds.Min.X + x*srcWidth/dstWidth
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}

	return dst
}

func fitIntoBox(width, height, maxEdge int) (int, int) {
	if width <= maxEdge && height <= maxEdge {
		return width, height
	}

	if width >= height {
		return maxEdge, max(1, height*maxEdge/width)
	}

	return max(1, width*maxEdge/height), maxEdge
}

func isSupportedImageFormat(format string) bool {
	switch strings.ToLower(format) {
	case "jpeg", "png", "gif":
		return true
	default:
		return false
	}
}

func imageFormatToContentType(format string) string {
	switch strings.ToLower(format) {
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}
