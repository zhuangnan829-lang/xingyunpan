package service

import (
	"strings"
	"testing"

	"xingyunpan-v2/internal/model"
)

func TestValidateStoragePolicyFileSize(t *testing.T) {
	policy := &model.StoragePolicy{MaxFileSize: 1, MaxFileSizeUnit: "MB"}

	if err := validateStoragePolicyFileSize(policy, 1024*1024); err != nil {
		t.Fatalf("expected size at limit to pass, got %v", err)
	}
	if err := validateStoragePolicyFileSize(policy, 1024*1024+1); err == nil {
		t.Fatal("expected oversize file to fail")
	}
}

func TestValidateStoragePolicyExtensionAllow(t *testing.T) {
	policy := &model.StoragePolicy{ExtensionMode: "allow", Extensions: "jpg, .png"}

	if err := validateStoragePolicyExtension(policy, "photo.JPG"); err != nil {
		t.Fatalf("expected allowed extension to pass, got %v", err)
	}
	if err := validateStoragePolicyExtension(policy, "archive.zip"); err == nil {
		t.Fatal("expected extension outside allow list to fail")
	}
}

func TestValidateStoragePolicyExtensionDeny(t *testing.T) {
	policy := &model.StoragePolicy{ExtensionMode: "deny", Extensions: "exe, bat"}

	if err := validateStoragePolicyExtension(policy, "document.pdf"); err != nil {
		t.Fatalf("expected extension outside deny list to pass, got %v", err)
	}
	if err := validateStoragePolicyExtension(policy, "setup.exe"); err == nil {
		t.Fatal("expected denied extension to fail")
	}
}

func TestValidateStoragePolicyNameRegex(t *testing.T) {
	policy := &model.StoragePolicy{NameRuleMode: "allow", NameRegex: `^[a-z0-9_-]+\.txt$`}

	if err := validateStoragePolicyNameRegex(policy, "report_01.txt"); err != nil {
		t.Fatalf("expected matching file name to pass, got %v", err)
	}
	if err := validateStoragePolicyNameRegex(policy, "报告.txt"); err == nil {
		t.Fatal("expected non-matching file name to fail")
	}
}

func TestValidateStoragePolicyInvalidRegex(t *testing.T) {
	policy := &model.StoragePolicy{NameRuleMode: "allow", NameRegex: `[`}

	err := validateStoragePolicyNameRegex(policy, "file.txt")
	if err == nil {
		t.Fatal("expected invalid regex to fail")
	}
	if !strings.Contains(err.Error(), "正则无效") {
		t.Fatalf("expected invalid regex message, got %v", err)
	}
}

func TestStoragePolicySizeBytesForChunkSize(t *testing.T) {
	tests := []struct {
		name string
		size int
		unit string
		want int64
	}{
		{name: "kb", size: 512, unit: "KB", want: 512 * 1024},
		{name: "mb", size: 25, unit: "MB", want: 25 * 1024 * 1024},
		{name: "gb", size: 2, unit: "GB", want: 2 * 1024 * 1024 * 1024},
		{name: "zero", size: 0, unit: "MB", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := storagePolicySizeBytes(tt.size, tt.unit); got != tt.want {
				t.Fatalf("storagePolicySizeBytes(%d, %q) = %d, want %d", tt.size, tt.unit, got, tt.want)
			}
		})
	}
}

func TestApplyStoragePolicyTemplateAndCleanPath(t *testing.T) {
	values := map[string]string{
		"uid":        "7",
		"path":       "docs/reports",
		"originname": "file.txt",
		"randomkey8": "abcd1234",
		"hash":       "hash-value",
	}

	rendered := applyStoragePolicyTemplate("/cloudreve/data/{uid}/{path}/{hash}", values)
	if got := cleanRelativeStoragePath(rendered); got != "cloudreve/data/7/docs/reports/hash-value" {
		t.Fatalf("unexpected rendered path: %s", got)
	}
}

func TestSanitizeStorageFileName(t *testing.T) {
	if got := sanitizeStorageFileName(`..\evil/name.txt`); got != "name.txt" {
		t.Fatalf("sanitizeStorageFileName returned %q", got)
	}
	if got := sanitizeStorageFileName(".."); got != "" {
		t.Fatalf("expected unsafe empty name, got %q", got)
	}
}

func TestEnsureUniqueStoragePath(t *testing.T) {
	existing := map[string]bool{"uploads/file.txt": true}
	got := ensureUniqueStoragePath("uploads/file.txt", func(path string) bool {
		return existing[path]
	})
	if got == "uploads/file.txt" || !strings.HasPrefix(got, "uploads/file-") || !strings.HasSuffix(got, ".txt") {
		t.Fatalf("expected unique suffixed path, got %q", got)
	}
}

func TestEscapeStoragePathForURL(t *testing.T) {
	got := escapeStoragePathForURL(`/uploads/7/报告 01.txt`)
	if got != "uploads/7/%E6%8A%A5%E5%91%8A%2001.txt" {
		t.Fatalf("unexpected escaped path: %s", got)
	}
}
