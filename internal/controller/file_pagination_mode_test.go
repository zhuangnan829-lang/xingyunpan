package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/pkg/response"
)

func TestFileListPaginationModesFollowFileSystemSettings(t *testing.T) {
	env := newAdminUserRealParamsEnv(t)
	adminToken := loginToken(t, env.router, "root", "rootpass1", response.CodeSuccess)

	var admin model.User
	if err := env.db.Where("username = ?", "root").First(&admin).Error; err != nil {
		t.Fatalf("load root user: %v", err)
	}
	baseTime := time.Now().Add(-time.Hour)
	for i := 1; i <= 5; i++ {
		file := model.UserFile{
			BaseModel: model.BaseModel{
				CreatedAt: baseTime.Add(time.Duration(i) * time.Minute),
				UpdatedAt: baseTime.Add(time.Duration(i) * time.Minute),
			},
			UserID:   admin.ID,
			FileName: fmt.Sprintf("page-mode-%d.txt", i),
			FilePath: fmt.Sprintf("/page-mode-%d.txt", i),
			IsFolder: false,
			FileSize: int64(i),
		}
		if err := env.db.Create(&file).Error; err != nil {
			t.Fatalf("seed file %d: %v", i, err)
		}
	}

	setFilePaginationMode(t, env, "cursor", 2)
	cursorResp := apiJSON(t, env.router, http.MethodGet, "/api/v1/file?page=3&page_size=50", adminToken, nil)
	requireResponseCode(t, cursorResp, response.CodeSuccess)
	cursorData := responseMap(t, cursorResp)
	assertPaginationMeta(t, cursorData, "cursor", 2, 2)
	firstCursor := stringField(t, cursorData, "next_cursor")
	if firstCursor == "" {
		t.Fatalf("cursor mode did not return next_cursor: %#v", cursorData)
	}
	assertListLength(t, cursorData, 2)

	cursorNext := apiJSON(t, env.router, http.MethodGet, "/api/v1/file?page_size=50&cursor="+firstCursor, adminToken, nil)
	requireResponseCode(t, cursorNext, response.CodeSuccess)
	cursorNextData := responseMap(t, cursorNext)
	assertPaginationMeta(t, cursorNextData, "cursor", 2, 2)
	if stringField(t, cursorNextData, "next_cursor") == "" {
		t.Fatalf("cursor mode second page did not return next_cursor")
	}
	assertListLength(t, cursorNextData, 2)

	setFilePaginationMode(t, env, "offset", 2)
	offsetResp := apiJSON(t, env.router, http.MethodGet, "/api/v1/file?page=2&page_size=50&cursor="+firstCursor, adminToken, nil)
	requireResponseCode(t, offsetResp, response.CodeSuccess)
	offsetData := responseMap(t, offsetResp)
	assertPaginationMeta(t, offsetData, "offset", 2, 2)
	if next := stringField(t, offsetData, "next_cursor"); next != "" {
		t.Fatalf("offset mode next_cursor = %q, want empty", next)
	}
	assertListLength(t, offsetData, 2)

	setFilePaginationMode(t, env, "hybrid", 2)
	hybridOffsetResp := apiJSON(t, env.router, http.MethodGet, "/api/v1/file?page=2&page_size=50", adminToken, nil)
	requireResponseCode(t, hybridOffsetResp, response.CodeSuccess)
	hybridOffsetData := responseMap(t, hybridOffsetResp)
	assertPaginationMeta(t, hybridOffsetData, "hybrid", 2, 2)
	if next := stringField(t, hybridOffsetData, "next_cursor"); next != "" {
		t.Fatalf("hybrid page mode without cursor next_cursor = %q, want empty", next)
	}

	hybridCursorResp := apiJSON(t, env.router, http.MethodGet, "/api/v1/file?page_size=50&cursor="+firstCursor, adminToken, nil)
	requireResponseCode(t, hybridCursorResp, response.CodeSuccess)
	hybridCursorData := responseMap(t, hybridCursorResp)
	assertPaginationMeta(t, hybridCursorData, "hybrid", 2, 2)
	if next := stringField(t, hybridCursorData, "next_cursor"); next == "" {
		t.Fatalf("hybrid cursor mode missed next_cursor")
	}
}

func setFilePaginationMode(t *testing.T, env *adminUserRealParamsEnv, mode string, maxPageSize int) {
	t.Helper()
	if err := env.db.Model(&model.FileSystemSetting{}).Where("id > 0").Updates(map[string]interface{}{
		"list_pagination_mode": mode,
		"max_page_size":        maxPageSize,
	}).Error; err != nil {
		t.Fatalf("update file pagination settings: %v", err)
	}
}

func responseMap(t *testing.T, resp response.Response) map[string]interface{} {
	t.Helper()
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("response data is not an object: %#v", resp.Data)
	}
	return data
}

func assertPaginationMeta(t *testing.T, data map[string]interface{}, mode string, pageSize int, maxPageSize int) {
	t.Helper()
	if got := stringField(t, data, "pagination_mode"); got != mode {
		t.Fatalf("pagination_mode = %q, want %q", got, mode)
	}
	if got := intField(t, data, "page_size"); got != pageSize {
		t.Fatalf("page_size = %d, want %d", got, pageSize)
	}
	if got := intField(t, data, "max_page_size"); got != maxPageSize {
		t.Fatalf("max_page_size = %d, want %d", got, maxPageSize)
	}
}

func assertListLength(t *testing.T, data map[string]interface{}, want int) {
	t.Helper()
	list, ok := data["list"].([]interface{})
	if !ok {
		t.Fatalf("list is not an array: %#v", data["list"])
	}
	if len(list) != want {
		t.Fatalf("list length = %d, want %d", len(list), want)
	}
}

func stringField(t *testing.T, data map[string]interface{}, key string) string {
	t.Helper()
	value, _ := data[key].(string)
	return value
}

func intField(t *testing.T, data map[string]interface{}, key string) int {
	t.Helper()
	switch value := data[key].(type) {
	case float64:
		return int(value)
	case int:
		return value
	case string:
		parsed, _ := strconv.Atoi(value)
		return parsed
	default:
		t.Fatalf("field %s is not numeric: %#v", key, data[key])
		return 0
	}
}
