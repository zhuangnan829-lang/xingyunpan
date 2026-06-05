package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

type fakeAdminShareControllerService struct {
	query      *service.AdminShareListQuery
	metricsDay int
	deletedID  uint
	batchIDs   []uint
}

func (s *fakeAdminShareControllerService) List(ctx context.Context, query *service.AdminShareListQuery) ([]service.AdminSharePayload, int64, error) {
	s.query = query
	return []service.AdminSharePayload{{ShareID: 7}}, 1, nil
}

func (s *fakeAdminShareControllerService) Metrics(ctx context.Context, expiringWithinDays int) (*service.AdminShareMetricsPayload, error) {
	s.metricsDay = expiringWithinDays
	return &service.AdminShareMetricsPayload{TotalShares: 9}, nil
}

func (s *fakeAdminShareControllerService) Delete(ctx context.Context, shareID uint) error {
	s.deletedID = shareID
	return nil
}

func (s *fakeAdminShareControllerService) BatchDelete(ctx context.Context, shareIDs []uint) error {
	s.batchIDs = shareIDs
	return nil
}

func TestAdminShareControllerParsesListFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fake := &fakeAdminShareControllerService{}
	controller := NewAdminShareController(fake)
	router := gin.New()
	router.GET("/admin/shares", controller.List)

	req := httptest.NewRequest(http.MethodGet, "/admin/shares?page=2&page_size=50&keyword=report&status=unavailable&owner_id=8&min_downloads=3&expiring_within_days=5&max_downloads_reached=true&unavailable=true", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d body=%s", rec.Code, rec.Body.String())
	}
	if fake.query == nil {
		t.Fatalf("service List was not called")
	}
	if fake.query.Page != 2 || fake.query.PageSize != 50 || fake.query.Keyword != "report" || fake.query.Status != "unavailable" || fake.query.OwnerID != 8 {
		t.Fatalf("query basics = %#v", fake.query)
	}
	if fake.query.MinDownloads == nil || *fake.query.MinDownloads != 3 {
		t.Fatalf("min downloads = %#v, want 3", fake.query.MinDownloads)
	}
	if fake.query.ExpiringWithinDays == nil || *fake.query.ExpiringWithinDays != 5 {
		t.Fatalf("expiring days = %#v, want 5", fake.query.ExpiringWithinDays)
	}
	if fake.query.MaxDownloadsReached == nil || !*fake.query.MaxDownloadsReached {
		t.Fatalf("max downloads reached = %#v, want true", fake.query.MaxDownloadsReached)
	}
	if fake.query.Unavailable == nil || !*fake.query.Unavailable {
		t.Fatalf("unavailable = %#v, want true", fake.query.Unavailable)
	}
}

func TestAdminShareControllerMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fake := &fakeAdminShareControllerService{}
	controller := NewAdminShareController(fake)
	router := gin.New()
	router.GET("/admin/shares/metrics", controller.Metrics)

	req := httptest.NewRequest(http.MethodGet, "/admin/shares/metrics?expiring_within_days=6", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d body=%s", rec.Code, rec.Body.String())
	}
	if fake.metricsDay != 6 {
		t.Fatalf("metrics day = %d, want 6", fake.metricsDay)
	}
	var resp response.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != response.CodeSuccess {
		t.Fatalf("response code = %d, want success", resp.Code)
	}
}
