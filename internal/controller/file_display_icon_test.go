package controller

import (
	"testing"

	"xingyunpan-v2/internal/service"
)

func TestParseFileEmojiOptionsSupportsCategoryJSON(t *testing.T) {
	options := parseFileEmojiOptions(&service.FileSystemSettingPayload{
		EmojiOptions: `{"enabled":true,"showInList":true,"fallbackUnknown":true,"folderEmoji":"\ud83d\udcc1","unknownEmoji":"\ud83d\uddc2\ufe0f","categories":[{"icon":"\ud83d\ude00","label":"Happy","match":"jpg,png,image/*","emojis":["\ud83d\ude00","\ud83d\ude03"]}]}`,
	})

	if len(options.Categories) != 1 {
		t.Fatalf("len(categories) = %d, want 1", len(options.Categories))
	}
	category := options.Categories[0]
	if category.Icon != "😀" || category.Label != "Happy" || category.Match != "jpg,png,image/*" || len(category.Emojis) != 2 {
		t.Fatalf("category = %#v, want parsed emoji category", category)
	}
}

func TestEmojiCategoryMatchExtensionReturnsCategoryIcon(t *testing.T) {
	options := fileEmojiOptions{
		Enabled:         true,
		ShowInList:      true,
		FallbackUnknown: true,
		UnknownEmoji:    "X",
		Categories: []fileEmojiCategory{
			{Icon: "I", Label: "Images", Match: "jpg,png,image/*", Emojis: []string{"I", "P"}},
			{Icon: "N", Label: "No Match", Emojis: []string{"N"}},
		},
	}

	got := resolveDisplayIcon("photo.JPG", false, "application/octet-stream", nil, options)
	if got.Icon != "I" || got.Label != "Images" || got.Source != "emoji_category" {
		t.Fatalf("resolveDisplayIcon() = %#v, want category icon", got)
	}
}

func TestEmojiCategoryMatchMimeWildcardReturnsCategoryIcon(t *testing.T) {
	options := fileEmojiOptions{
		Enabled:    true,
		ShowInList: true,
		Categories: []fileEmojiCategory{
			{Icon: "A", Label: "Audio", Match: "audio/*"},
		},
	}

	got := resolveDisplayIcon("track.bin", false, "audio/mpeg", nil, options)
	if got.Icon != "A" || got.Label != "Audio" || got.Source != "emoji_category" {
		t.Fatalf("resolveDisplayIcon() = %#v, want category icon", got)
	}
}

func TestFileIconRulesTakePriorityOverEmojiCategory(t *testing.T) {
	options := fileEmojiOptions{
		Enabled:    true,
		ShowInList: true,
		Categories: []fileEmojiCategory{
			{Icon: "C", Label: "Category", Match: "pdf,application/pdf"},
		},
	}
	rules := []fileIconRule{
		{Icon: "R", Label: "Rule", Match: "pdf", Tint: "#ff0000"},
	}

	got := resolveDisplayIcon("manual.pdf", false, "application/pdf", rules, options)
	if got.Icon != "R" || got.Label != "Rule" || got.Source != "rule" || got.Tint != "#ff0000" {
		t.Fatalf("resolveDisplayIcon() = %#v, want rule icon", got)
	}
}

func TestEmojiFallbackUnknownStillWorks(t *testing.T) {
	options := fileEmojiOptions{
		Enabled:         true,
		ShowInList:      true,
		FallbackUnknown: true,
		UnknownEmoji:    "U",
		Categories: []fileEmojiCategory{
			{Icon: "C", Label: "Category", Emojis: []string{"C"}},
		},
	}

	got := resolveDisplayIcon("archive.unknown", false, "application/octet-stream", nil, options)
	if got.Icon != "U" || got.Label != "Unknown file" || got.Source != "emoji" {
		t.Fatalf("resolveDisplayIcon() = %#v, want fallback unknown icon", got)
	}
}
