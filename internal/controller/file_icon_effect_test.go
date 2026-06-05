package controller

import (
	"encoding/json"
	"net/http"
	"testing"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

func TestFileIconSettingsAffectFileListDisplayIcons(t *testing.T) {
	env := newAdminUserRealParamsEnv(t)
	adminToken := loginToken(t, env.router, "root", "rootpass1", response.CodeSuccess)
	admin := loadTestUserByUsername(t, env, "root")

	seedIconTestFile(t, env, admin, "report.pdf", "application/pdf")
	seedIconTestFile(t, env, admin, "mystery.zzz", "application/octet-stream")

	folder := model.UserFile{
		UserID:   admin.ID,
		FileName: "icon-folder",
		FilePath: "/icon-folder",
		IsFolder: true,
	}
	if err := env.db.Create(&folder).Error; err != nil {
		t.Fatalf("seed folder: %v", err)
	}

	updateTestFileIconSettings(t, env, `[{"label":"PDF One","icon":"P1","match":"pdf","tint":"#123456"}]`, `{"enabled":true,"showInList":true,"fallbackUnknown":true,"folderEmoji":"F1","unknownEmoji":"U1","categories":[]}`)
	firstList := listFilesForIconTest(t, env, adminToken)
	report := findListItemByName(t, firstList, "report.pdf")
	if report["display_icon"] != "P1" || report["display_icon_tint"] != "#123456" || report["display_icon_label"] != "PDF One" {
		t.Fatalf("report icon = %#v, want PDF One rule", report)
	}

	updateTestFileIconSettings(t, env, `[{"label":"PDF Two","icon":"P2","match":"txt pdf","tint":"#abcdef"}]`, `{"enabled":true,"showInList":true,"fallbackUnknown":true,"folderEmoji":"F1","unknownEmoji":"U1","categories":[]}`)
	secondList := listFilesForIconTest(t, env, adminToken)
	report = findListItemByName(t, secondList, "report.pdf")
	if report["display_icon"] != "P2" || report["display_icon_tint"] != "#abcdef" || report["display_icon_label"] != "PDF Two" {
		t.Fatalf("report icon after match change = %#v, want PDF Two rule", report)
	}

	unknown := findListItemByName(t, secondList, "mystery.zzz")
	if unknown["display_icon"] != "U1" {
		t.Fatalf("unknown fallback icon = %#v, want U1", unknown)
	}

	updateTestFileIconSettings(t, env, `[{"label":"PDF Two","icon":"P2","match":"txt pdf","tint":"#abcdef"}]`, `{"enabled":false,"showInList":true,"fallbackUnknown":true,"folderEmoji":"F2","unknownEmoji":"U2","categories":[]}`)
	disabledEmojiList := listFilesForIconTest(t, env, adminToken)
	unknown = findListItemByName(t, disabledEmojiList, "mystery.zzz")
	if _, ok := unknown["display_icon"]; ok {
		t.Fatalf("unknown file returned fallback icon while emoji disabled: %#v", unknown)
	}

	updateTestFileIconSettings(t, env, `[{"label":"PDF Two","icon":"P2","match":"txt pdf","tint":"#abcdef"}]`, `{"enabled":true,"showInList":true,"fallbackUnknown":true,"folderEmoji":"F3","unknownEmoji":"U2","categories":[]}`)
	folderList := listFilesForIconTest(t, env, adminToken)
	folderItem := findListItemByName(t, folderList, "icon-folder")
	if folderItem["display_icon"] != "F3" {
		t.Fatalf("folder display icon = %#v, want F3", folderItem)
	}
}

func loadTestUserByUsername(t *testing.T, env *adminUserRealParamsEnv, username string) model.User {
	t.Helper()
	var user model.User
	if err := env.db.Where("username = ?", username).First(&user).Error; err != nil {
		t.Fatalf("load user %s: %v", username, err)
	}
	return user
}

func updateTestFileIconSettings(t *testing.T, env *adminUserRealParamsEnv, rules string, emojiOptions string) {
	t.Helper()
	settings := service.NewFileSystemSettingService(repository.NewFileSystemSettingRepository(env.db), nil)
	if _, err := settings.UpdateIcons(&service.FileSystemIconSettingsPayload{
		FileIconRules: rules,
		EmojiOptions:  emojiOptions,
	}); err != nil {
		t.Fatalf("update icon settings through service: %v", err)
	}
}

func seedIconTestFile(t *testing.T, env *adminUserRealParamsEnv, user model.User, name string, contentType string) {
	t.Helper()
	physical := model.PhysicalFile{
		FileHash:    "icon-test-" + name,
		FileSize:    32,
		StoragePath: "icon-test/" + name,
		RefCount:    1,
		StorageType: "local",
		ContentType: contentType,
	}
	if err := env.db.Create(&physical).Error; err != nil {
		t.Fatalf("seed physical file %s: %v", name, err)
	}
	file := model.UserFile{
		UserID:         user.ID,
		FileName:       name,
		FilePath:       "/" + name,
		IsFolder:       false,
		FileSize:       physical.FileSize,
		PhysicalFileID: &physical.ID,
	}
	if err := env.db.Create(&file).Error; err != nil {
		t.Fatalf("seed user file %s: %v", name, err)
	}
}

func listFilesForIconTest(t *testing.T, env *adminUserRealParamsEnv, token string) []map[string]interface{} {
	t.Helper()
	resp := apiJSON(t, env.router, http.MethodGet, "/api/v1/file?page_size=50", token, nil)
	requireResponseCode(t, resp, response.CodeSuccess)
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("response data is not object: %#v", resp.Data)
	}
	rawList, ok := data["list"].([]interface{})
	if !ok {
		t.Fatalf("response list is not array: %#v", data["list"])
	}
	items := make([]map[string]interface{}, 0, len(rawList))
	for _, raw := range rawList {
		item, ok := raw.(map[string]interface{})
		if !ok {
			encoded, _ := json.Marshal(raw)
			t.Fatalf("list item is not object: %s", string(encoded))
		}
		items = append(items, item)
	}
	return items
}

func findListItemByName(t *testing.T, items []map[string]interface{}, name string) map[string]interface{} {
	t.Helper()
	for _, item := range items {
		if item["name"] == name {
			return item
		}
	}
	t.Fatalf("file %s not found in list %#v", name, items)
	return nil
}
