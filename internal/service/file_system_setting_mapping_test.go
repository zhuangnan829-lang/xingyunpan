package service

import (
	"encoding/json"
	"testing"
)

func TestDefaultMimeMapCoversCommonTypes(t *testing.T) {
	var mapping map[string]string
	if err := json.Unmarshal([]byte(defaultMimeMapJSON()), &mapping); err != nil {
		t.Fatalf("defaultMimeMapJSON is invalid: %v", err)
	}

	expected := map[string]string{
		".pdf":  "application/pdf",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".webp": "image/webp",
		".avif": "image/avif",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".mp3":  "audio/mpeg",
		".aac":  "audio/aac",
		".zip":  "application/zip",
	}
	for ext, mime := range expected {
		if mapping[ext] != mime {
			t.Fatalf("expected %s => %s, got %q", ext, mime, mapping[ext])
		}
	}
}

func TestLegacyDefaultMappingsUpgrade(t *testing.T) {
	if got := upgradeDefaultMimeMap(legacyDefaultMimeMapJSON); got == legacyDefaultMimeMapJSON {
		t.Fatal("legacy MIME map should upgrade to expanded default")
	}
	if got := upgradeDefaultCategoryQuery("audio", legacyDefaultAudioQuery); got != defaultAudioCategoryQuery() {
		t.Fatalf("legacy audio query should upgrade, got %q", got)
	}
	if got := upgradeDefaultCategoryQuery("audio", "type=file&case_folding&use_or&name=*.mp3&name=*.acc"); got != defaultAudioCategoryQuery() {
		t.Fatalf("audio query containing .acc without .aac should upgrade, got %q", got)
	}
}

func TestDefaultCategoryQueriesParseExpectedExtensions(t *testing.T) {
	cases := map[string][]string{
		defaultImageCategoryQuery():    {"jpg", "webp", "svg", "heic", "avif", "dng"},
		defaultVideoCategoryQuery():    {"mp4", "m3u8", "mov", "webm", "rmvb", "ts"},
		defaultAudioCategoryQuery():    {"mp3", "flac", "aac", "opus", "wma", "amr"},
		defaultDocumentCategoryQuery(): {"pdf", "docx", "xlsx", "md", "csv", "epub", "azw3"},
	}
	for query, expectedExtensions := range cases {
		extensions := categoryExtensionsFromQuery(query, nil)
		seen := map[string]struct{}{}
		for _, ext := range extensions {
			seen[ext] = struct{}{}
		}
		for _, ext := range expectedExtensions {
			if _, ok := seen[ext]; !ok {
				t.Fatalf("expected query %q to include extension %s; got %v", query, ext, extensions)
			}
		}
	}
}

func TestDefaultFileSystemSettingsUseReadableIconAndEmojiDefaults(t *testing.T) {
	payload := defaultFileSystemSettingPayload()

	var rules []fileIconRulePayload
	if err := json.Unmarshal([]byte(payload.FileIconRules), &rules); err != nil {
		t.Fatalf("default file icon rules are invalid: %v", err)
	}
	if len(rules) == 0 {
		t.Fatal("expected default file icon rules")
	}
	if rules[0].Label != "音频文件" || rules[0].Icon != "🎧" {
		t.Fatalf("first default rule = %#v, want readable Chinese label and emoji", rules[0])
	}

	var emoji emojiOptionsPayload
	if err := json.Unmarshal([]byte(payload.EmojiOptions), &emoji); err != nil {
		t.Fatalf("default emoji options are invalid: %v", err)
	}
	if emoji.FolderEmoji != "📁" || emoji.UnknownEmoji != "🗂️" {
		t.Fatalf("emoji defaults = folder %q unknown %q", emoji.FolderEmoji, emoji.UnknownEmoji)
	}
}

func TestLegacyMojibakeIconAndEmojiDefaultsUpgrade(t *testing.T) {
	payload := defaultFileSystemSettingPayload()
	payload.FileIconRules = "\u95ca\u62bd \u7459\u55db \u9365\u5267 PDF \u93c2 Word \u93c2 Excel \u741b Go \u5a67 mp3,flac,ape,wav,aac,ogg,m4a m3u8,mp4,flv,avi,wmv,mkv,rm,rmvb,mov,webm drawio,dwb,excalidraw"
	payload.EmojiOptions = "{\"enabled\":true,\"showInList\":true,\"fallbackUnknown\":true,\"folderEmoji\":\"\u9983\u6427\",\"unknownEmoji\":\"\u9983\u642b\",\"categories\":[]}"

	normalized, err := normalizeFileSystemSettingPayload(payload)
	if err != nil {
		t.Fatalf("normalizeFileSystemSettingPayload returned error: %v", err)
	}

	var rules []fileIconRulePayload
	if err := json.Unmarshal([]byte(normalized.FileIconRules), &rules); err != nil {
		t.Fatalf("normalized icon rules are invalid: %v", err)
	}
	if rules[0].Label != "音频文件" {
		t.Fatalf("legacy icon rules were not upgraded: %#v", rules[0])
	}

	var emoji emojiOptionsPayload
	if err := json.Unmarshal([]byte(normalized.EmojiOptions), &emoji); err != nil {
		t.Fatalf("normalized emoji options are invalid: %v", err)
	}
	if emoji.FolderEmoji != "📁" || emoji.UnknownEmoji != "🗂️" {
		t.Fatalf("legacy emoji options were not upgraded: %#v", emoji)
	}
}

func TestLegacyVisualDefaultFileIconRulesUpgrade(t *testing.T) {
	payload := defaultFileSystemSettingPayload()
	legacyRules := []fileIconRulePayload{
		{Label: "音频文件", Icon: "🎵", Match: "mp3,flac,ape,wav,aac,ogg,m4a", Tint: "#7c3aed"},
		{Label: "视频文件", Icon: "🎬", Match: "m3u8,mp4,flv,avi,wmv,mkv,rm,rmvb,mov,webm", Tint: "#dc2626"},
		{Label: "图片文件", Icon: "🖼️", Match: "bmp,iff,png,gif,jpg,jpeg,psd,svg,webp,heif,heic,tiff,avif", Tint: "#ea580c"},
		{Label: "RAW 图片", Icon: "RAW", Match: "3fr,ari,arw,bay,braw,crw,cr2,cr3,cap,dcs,dcr,dng,drf,eip,erf,fff,gpr,iiq,k25,kdc,mdc,mef,mos,mrw,nef,nrw,obm,orf,pef,ptx,pxn,r3d,raf,raw,rwl,rw2,rwz,sr2,srf,srw,tif,x3f", Tint: "#ef4444"},
		{Label: "PDF 文档", Icon: "📕", Match: "pdf", Tint: "#ef4444"},
		{Label: "Word 文档", Icon: "W", Match: "doc,docx", Tint: "#3b82f6"},
		{Label: "PowerPoint 演示", Icon: "P", Match: "ppt,pptx", Tint: "#f97316"},
		{Label: "Excel 表格", Icon: "X", Match: "xls,xlsx,csv", Tint: "#22c55e"},
		{Label: "文本与网页", Icon: "📝", Match: "txt,md,markdown,html,htm,xml,json,yaml,yml", Tint: "#64748b"},
		{Label: "压缩包", Icon: "🗜️", Match: "zip,gz,xz,tar,rar,7z,bz2", Tint: "#f59e0b"},
		{Label: "程序安装包", Icon: "⚙️", Match: "exe,msi,bat,cmd", Tint: "#1e3a8a"},
		{Label: "Android 安装包", Icon: "🤖", Match: "apk", Tint: "#84cc16"},
		{Label: "Go 源码", Icon: "Go", Match: "go", Tint: "#06b6d4"},
		{Label: "Python 源码", Icon: "Py", Match: "py", Tint: "#3776ab"},
		{Label: "C 源码", Icon: "C", Match: "c,h", Tint: "#84cc16"},
		{Label: "C++ 源码", Icon: "C++", Match: "cpp,cxx,hpp,hxx", Tint: "#ec4899"},
		{Label: "JavaScript / TypeScript", Icon: "JS", Match: "js,jsx,ts,tsx", Tint: "#facc15"},
		{Label: "Rust 源码", Icon: "Rs", Match: "rs", Tint: "#111827"},
		{Label: "电子书", Icon: "📗", Match: "epub,mobi,azw3", Tint: "#65a30d"},
		{Label: "磁力 / 种子", Icon: "🧲", Match: "torrent", Tint: "#6366f1"},
		{Label: "流程图 / 白板", Icon: "✏️", Match: "drawio,dwb,excalidraw", Tint: "#f97316"},
	}
	data, err := json.Marshal(legacyRules)
	if err != nil {
		t.Fatalf("marshal legacy rules: %v", err)
	}
	payload.FileIconRules = string(data)

	normalized, err := normalizeFileSystemSettingPayload(payload)
	if err != nil {
		t.Fatalf("normalizeFileSystemSettingPayload returned error: %v", err)
	}

	var rules []fileIconRulePayload
	if err := json.Unmarshal([]byte(normalized.FileIconRules), &rules); err != nil {
		t.Fatalf("normalized icon rules are invalid: %v", err)
	}
	iconsByLabel := map[string]string{}
	for _, rule := range rules {
		iconsByLabel[rule.Label] = rule.Icon
	}
	if iconsByLabel["音频文件"] != "🎧" || iconsByLabel["视频文件"] != "🎞️" || iconsByLabel["Word 文档"] != "📘" || iconsByLabel["Excel 表格"] != "📗" {
		t.Fatalf("legacy visual defaults were not upgraded: %#v", iconsByLabel)
	}
}

func TestCustomIconAndEmojiSettingsAreNotOverwrittenByDefaultUpgrade(t *testing.T) {
	payload := defaultFileSystemSettingPayload()
	payload.FileIconRules = `[{"label":"自定义 PDF","icon":"P","match":"pdf","tint":"#123456"}]`
	payload.EmojiOptions = `{"enabled":true,"showInList":true,"fallbackUnknown":false,"folderEmoji":"F","unknownEmoji":"U","categories":[{"icon":"C","label":"自定义","match":"custom","emojis":["C"]}]}`

	normalized, err := normalizeFileSystemSettingPayload(payload)
	if err != nil {
		t.Fatalf("normalizeFileSystemSettingPayload returned error: %v", err)
	}

	var rules []fileIconRulePayload
	if err := json.Unmarshal([]byte(normalized.FileIconRules), &rules); err != nil {
		t.Fatalf("normalized icon rules are invalid: %v", err)
	}
	if len(rules) != 1 || rules[0].Label != "自定义 PDF" || rules[0].Icon != "P" {
		t.Fatalf("custom icon rules were overwritten: %#v", rules)
	}

	var emoji emojiOptionsPayload
	if err := json.Unmarshal([]byte(normalized.EmojiOptions), &emoji); err != nil {
		t.Fatalf("normalized emoji options are invalid: %v", err)
	}
	if emoji.FolderEmoji != "F" || emoji.UnknownEmoji != "U" || len(emoji.Categories) != 1 || emoji.Categories[0].Label != "自定义" {
		t.Fatalf("custom emoji options were overwritten: %#v", emoji)
	}
}
