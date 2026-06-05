package service

import (
	"encoding/json"
	"strings"
)

type emojiOptionsPayload struct {
	Enabled         bool                   `json:"enabled"`
	ShowInList      bool                   `json:"showInList"`
	FallbackUnknown bool                   `json:"fallbackUnknown"`
	FolderEmoji     string                 `json:"folderEmoji"`
	UnknownEmoji    string                 `json:"unknownEmoji"`
	Categories      []emojiCategoryPayload `json:"categories"`
}

type emojiCategoryPayload struct {
	Icon   string   `json:"icon"`
	Label  string   `json:"label,omitempty"`
	Match  string   `json:"match,omitempty"`
	Emojis []string `json:"emojis"`
}

func defaultEmojiOptionsJSON() string {
	payload := emojiOptionsPayload{
		Enabled:         true,
		ShowInList:      true,
		FallbackUnknown: true,
		FolderEmoji:     "📁",
		UnknownEmoji:    "🗂️",
		Categories: []emojiCategoryPayload{
			{Icon: "🖼️", Label: "图片文件", Match: "jpg,jpeg,png,gif,webp,svg,image/*", Emojis: []string{"🖼️", "📷", "🌄"}},
			{Icon: "🎵", Label: "音频文件", Match: "mp3,flac,wav,aac,ogg,audio/*", Emojis: []string{"🎵", "🎧", "🎙️"}},
			{Icon: "🎬", Label: "视频文件", Match: "mp4,mkv,mov,webm,video/*", Emojis: []string{"🎬", "📹", "🎞️"}},
			{Icon: "📄", Label: "文档文件", Match: "pdf,doc,docx,xls,xlsx,ppt,pptx", Emojis: []string{"📄", "📕", "📊"}},
			{Icon: "😀", Label: "常用 Emoji", Emojis: []string{"😀", "😃", "😄", "😁", "😊", "👍", "✨"}},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return `{"enabled":true,"showInList":true,"fallbackUnknown":true,"folderEmoji":"📁","unknownEmoji":"🗂️","categories":[]}`
	}
	return string(data)
}

func shouldUpgradeEmojiOptions(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	return trimmed == "" || !strings.Contains(trimmed, `"categories"`) || isLegacyMojibakeEmojiOptions(trimmed)
}

func isLegacyMojibakeEmojiOptions(raw string) bool {
	return strings.Contains(raw, "\"folderEmoji\":\"\u9983\u6427\"") ||
		strings.Contains(raw, "\"unknownEmoji\":\"\u9983\u642b\"") ||
		strings.Contains(raw, "\"unknownEmoji\":\"\u9983\u68bb") ||
		strings.Contains(raw, "\"icon\":\"\u9983")
}
