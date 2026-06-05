package controller

import (
	"testing"

	"xingyunpan-v2/internal/service"
)

func TestAdminFileDisplayIconUsesSharedFileIconRules(t *testing.T) {
	rules := []fileIconRule{
		{Icon: "PDFX", Label: "PDF Custom", Match: "pdf,application/pdf", Tint: "#123456"},
	}
	emojiOptions := fileEmojiOptions{
		Enabled:         true,
		ShowInList:      true,
		FallbackUnknown: true,
		UnknownEmoji:    "U",
	}

	ordinary := resolveDisplayIcon("report.pdf", false, "application/pdf", rules, emojiOptions)
	admin := service.AdminFilePayload{
		FileName:    "report.pdf",
		ContentType: "application/pdf",
	}
	applyAdminFileDisplayIcon(&admin, rules, emojiOptions, nil)

	if admin.DisplayIcon != ordinary.Icon ||
		admin.DisplayIconTint != ordinary.Tint ||
		admin.DisplayIconLabel != ordinary.Label ||
		admin.DisplayIconSource != ordinary.Source {
		t.Fatalf("admin display icon = %#v, ordinary = %#v", admin, ordinary)
	}
}
