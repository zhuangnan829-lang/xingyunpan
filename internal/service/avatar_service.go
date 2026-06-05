package service

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

const avatarUploadRoot = "uploads"

// AvatarRuntimeSettings is the effective site_settings subset used by avatar runtime logic.
type AvatarRuntimeSettings struct {
	Path       string
	LimitBytes int64
	Dimension  int
}

// AvatarService applies site avatar settings to upload storage, image processing, and URLs.
type AvatarService interface {
	GetRuntimeSettings() (AvatarRuntimeSettings, error)
	SaveUploadedAvatar(userID uint, file *multipart.FileHeader) (string, error)
}

type avatarService struct {
	siteSettingRepo repository.SiteSettingRepository
}

// NewAvatarService creates an avatar runtime service.
func NewAvatarService(siteSettingRepo repository.SiteSettingRepository) AvatarService {
	return &avatarService{siteSettingRepo: siteSettingRepo}
}

func (s *avatarService) GetRuntimeSettings() (AvatarRuntimeSettings, error) {
	setting, err := s.resolveSiteSetting()
	if err != nil {
		return AvatarRuntimeSettings{}, err
	}
	return avatarRuntimeSettings(setting)
}

func (s *avatarService) SaveUploadedAvatar(userID uint, file *multipart.FileHeader) (string, error) {
	if userID == 0 {
		return "", fmt.Errorf("user ID cannot be empty")
	}
	if file == nil {
		return "", fmt.Errorf("please choose an avatar image")
	}

	settings, err := s.GetRuntimeSettings()
	if err != nil {
		return "", err
	}
	if file.Size <= 0 || file.Size > settings.LimitBytes {
		return "", fmt.Errorf("avatar image must be smaller than %s", formatAvatarLimit(settings.LimitBytes))
	}

	opened, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to read avatar image")
	}
	defer opened.Close()

	contentType, err := detectAvatarContentType(opened)
	if err != nil {
		return "", err
	}
	outputExt, err := avatarOutputExtension(contentType)
	if err != nil {
		return "", err
	}
	if seeker, ok := opened.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return "", fmt.Errorf("failed to read avatar image")
		}
	}

	src, _, err := image.Decode(opened)
	if err != nil {
		return "", fmt.Errorf("failed to decode avatar image")
	}
	resized := squareResizeImage(src, settings.Dimension)

	storageDir := filepath.Join(avatarUploadRoot, filepath.FromSlash(settings.Path))
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to prepare avatar storage")
	}

	filename := fmt.Sprintf("%d-%d-%s%s", userID, time.Now().UnixNano(), uuid.NewString(), outputExt)
	dst := filepath.Join(storageDir, filename)
	out, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("failed to save avatar image")
	}
	encodeErr := encodeAvatarImage(out, resized, outputExt)
	closeErr := out.Close()
	if encodeErr != nil {
		_ = os.Remove(dst)
		return "", encodeErr
	}
	if closeErr != nil {
		_ = os.Remove(dst)
		return "", fmt.Errorf("failed to save avatar image")
	}

	return "/api/v1/avatars/" + settings.Path + "/" + filename, nil
}

func (s *avatarService) resolveSiteSetting() (*model.SiteSetting, error) {
	if s.siteSettingRepo == nil {
		return defaultAvatarSiteSetting(), nil
	}
	setting, err := s.siteSettingRepo.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return defaultAvatarSiteSetting(), nil
	}
	return setting, nil
}

func avatarRuntimeSettings(setting *model.SiteSetting) (AvatarRuntimeSettings, error) {
	if setting == nil {
		setting = defaultAvatarSiteSetting()
	}
	avatarPath, err := NormalizeAvatarPath(setting.AvatarPath)
	if err != nil {
		return AvatarRuntimeSettings{}, err
	}
	sizeLimitMB := setting.AvatarSizeLimitMB
	if sizeLimitMB <= 0 {
		sizeLimitMB = 4
	}
	dimension := setting.AvatarDimension
	if dimension <= 0 {
		dimension = 200
	}
	return AvatarRuntimeSettings{
		Path:       avatarPath,
		LimitBytes: int64(sizeLimitMB) << 20,
		Dimension:  dimension,
	}, nil
}

func defaultAvatarSiteSetting() *model.SiteSetting {
	return &model.SiteSetting{
		AvatarPath:        "avatar",
		AvatarSizeLimitMB: 4,
		AvatarDimension:   200,
		GravatarServer:    "https://www.gravatar.com/",
	}
}

// NormalizeAvatarPath keeps avatar storage paths relative to uploads and URL-safe.
func NormalizeAvatarPath(value string) (string, error) {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	value = strings.Trim(value, "/")
	if value == "" {
		return "", fmt.Errorf("avatar storage path cannot be empty")
	}
	if filepath.IsAbs(value) || strings.Contains(value, ":") {
		return "", fmt.Errorf("avatar storage path must be relative")
	}
	cleaned := path.Clean(value)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", fmt.Errorf("avatar storage path cannot escape uploads")
	}
	return cleaned, nil
}

// AvatarFilePathFromURLPath maps /api/v1/avatars/* URL suffixes safely into uploads.
func AvatarFilePathFromURLPath(urlPath string) (string, error) {
	rel := strings.Trim(strings.ReplaceAll(urlPath, "\\", "/"), "/")
	if rel == "" {
		return "", fmt.Errorf("avatar path cannot be empty")
	}
	cleaned := path.Clean(rel)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", fmt.Errorf("avatar path cannot escape uploads")
	}
	return filepath.Join(avatarUploadRoot, filepath.FromSlash(cleaned)), nil
}

func detectAvatarContentType(file multipart.File) (string, error) {
	buffer := make([]byte, 512)
	n, _ := file.Read(buffer)
	contentType := http.DetectContentType(buffer[:n])
	if _, err := avatarOutputExtension(contentType); err != nil {
		return "", err
	}
	return contentType, nil
}

func avatarOutputExtension(contentType string) (string, error) {
	switch strings.ToLower(contentType) {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	case "image/webp", "image/gif":
		return ".png", nil
	default:
		return "", fmt.Errorf("only JPEG, PNG, WebP and GIF images are supported")
	}
}

func squareResizeImage(src image.Image, dimension int) image.Image {
	if dimension <= 0 {
		dimension = 200
	}
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	side := width
	if height < side {
		side = height
	}
	crop := image.Rect(
		bounds.Min.X+(width-side)/2,
		bounds.Min.Y+(height-side)/2,
		bounds.Min.X+(width-side)/2+side,
		bounds.Min.Y+(height-side)/2+side,
	)
	dst := image.NewRGBA(image.Rect(0, 0, dimension, dimension))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, crop, draw.Over, nil)
	return dst
}

func encodeAvatarImage(writer io.Writer, img image.Image, ext string) error {
	switch ext {
	case ".jpg":
		if err := jpeg.Encode(writer, img, &jpeg.Options{Quality: 90}); err != nil {
			return fmt.Errorf("failed to encode avatar image")
		}
	case ".png":
		if err := png.Encode(writer, img); err != nil {
			return fmt.Errorf("failed to encode avatar image")
		}
	case ".gif":
		if err := gif.Encode(writer, img, nil); err != nil {
			return fmt.Errorf("failed to encode avatar image")
		}
	default:
		return fmt.Errorf("unsupported avatar output format")
	}
	return nil
}

func formatAvatarLimit(bytes int64) string {
	if bytes%(1<<20) == 0 {
		return fmt.Sprintf("%dMB", bytes/(1<<20))
	}
	return fmt.Sprintf("%.2fMB", float64(bytes)/float64(1<<20))
}

func GravatarURL(server, email string, size int) string {
	server = strings.TrimSpace(server)
	if server == "" {
		server = "https://www.gravatar.com/"
	}
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	sum := md5.Sum([]byte(normalizedEmail))
	hash := hex.EncodeToString(sum[:])
	if size <= 0 {
		size = 200
	}
	if strings.Contains(server, "{hash}") {
		return strings.ReplaceAll(server, "{hash}", hash)
	}

	parsed, err := url.Parse(server)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		server = "https://www.gravatar.com/"
		parsed, _ = url.Parse(server)
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/avatar/" + hash
	query := parsed.Query()
	if query.Get("d") == "" {
		query.Set("d", "identicon")
	}
	if query.Get("s") == "" {
		query.Set("s", fmt.Sprintf("%d", size))
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
