package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

func TestAdminUserListQueryFiltersAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newAdminUserControllerTestDB(t)
	seedAdminUserListData(t, db)

	controller := NewAdminUserController(service.NewAdminUserService(
		repository.NewUserRepository(db),
		repository.NewUserGroupRepository(db),
	))
	router := gin.New()
	router.GET("/api/v1/admin/users", controller.List)

	tests := []struct {
		name          string
		query         string
		wantTotal     int64
		wantItemCount int
		wantUsernames []string
	}{
		{
			name:          "keyword matches username",
			query:         "keyword=alpha",
			wantTotal:     1,
			wantItemCount: 1,
			wantUsernames: []string{"alpha"},
		},
		{
			name:          "keyword matches email",
			query:         "keyword=bravo@example.com",
			wantTotal:     1,
			wantItemCount: 1,
			wantUsernames: []string{"bravo"},
		},
		{
			name:          "keyword matches group",
			query:         "keyword=Pro",
			wantTotal:     2,
			wantItemCount: 2,
			wantUsernames: []string{"charlie", "bravo"},
		},
		{
			name:          "role filter",
			query:         "role=admin",
			wantTotal:     2,
			wantItemCount: 2,
			wantUsernames: []string{"delta", "alpha"},
		},
		{
			name:          "status filter",
			query:         "status=disabled",
			wantTotal:     1,
			wantItemCount: 1,
			wantUsernames: []string{"charlie"},
		},
		{
			name:          "group id filter",
			query:         "user_group_id=2",
			wantTotal:     2,
			wantItemCount: 2,
			wantUsernames: []string{"charlie", "bravo"},
		},
		{
			name:          "group name filter",
			query:         "user_group_name=Guest",
			wantTotal:     1,
			wantItemCount: 1,
			wantUsernames: []string{"echo"},
		},
		{
			name:          "page size",
			query:         "page=1&page_size=2",
			wantTotal:     5,
			wantItemCount: 2,
			wantUsernames: []string{"echo", "delta"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?"+tt.query, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("HTTP status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
			}

			var body response.Response
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			dataBytes, err := json.Marshal(body.Data)
			if err != nil {
				t.Fatalf("marshal data: %v", err)
			}
			var result service.AdminUserListResult
			if err := json.Unmarshal(dataBytes, &result); err != nil {
				t.Fatalf("decode list result: %v", err)
			}

			if result.Total != tt.wantTotal {
				t.Fatalf("total = %d, want %d", result.Total, tt.wantTotal)
			}
			if len(result.Items) != tt.wantItemCount {
				t.Fatalf("items length = %d, want %d", len(result.Items), tt.wantItemCount)
			}
			for i, username := range tt.wantUsernames {
				if result.Items[i].Username != username {
					t.Fatalf("items[%d].username = %q, want %q", i, result.Items[i].Username, username)
				}
			}
		})
	}
}

func newAdminUserControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "admin-users.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.UserGroup{}); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

func seedAdminUserListData(t *testing.T, db *gorm.DB) {
	t.Helper()
	groups := []model.UserGroup{
		{Name: "Default", MaxCapacity: 1024},
		{Name: "Pro", MaxCapacity: 2048},
		{Name: "Guest", MaxCapacity: 512},
	}
	for i := range groups {
		if err := db.Create(&groups[i]).Error; err != nil {
			t.Fatalf("seed group: %v", err)
		}
	}

	users := []model.User{
		{Username: "alpha", Email: "alpha@example.com", Password: "password", Role: "admin", Enabled: true, UserGroupID: groups[0].ID, Capacity: 1024},
		{Username: "bravo", Email: "bravo@example.com", Password: "password", Role: "user", Enabled: true, UserGroupID: groups[1].ID, Capacity: 2048},
		{Username: "charlie", Email: "charlie@example.com", Password: "password", Role: "user", Enabled: false, UserGroupID: groups[1].ID, Capacity: 2048},
		{Username: "delta", Email: "delta@example.com", Password: "password", Role: "admin", Enabled: true, UserGroupID: groups[0].ID, Capacity: 1024},
		{Username: "echo", Email: "echo@example.com", Password: "password", Role: "user", Enabled: true, UserGroupID: groups[2].ID, Capacity: 512},
	}
	for i := range users {
		if err := db.Select("*").Create(&users[i]).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}
	if err := db.Model(&model.User{}).Where("username = ?", "charlie").Update("enabled", false).Error; err != nil {
		t.Fatalf("seed disabled user: %v", err)
	}
}
