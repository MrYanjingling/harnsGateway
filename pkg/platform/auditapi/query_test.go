package auditapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"harns-platform/pkg/audit"
)

type fakeReader struct {
	q audit.Query
}

func (f *fakeReader) List(_ context.Context, q audit.Query) ([]*audit.Event, int64, error) {
	f.q = q
	return []*audit.Event{}, 0, nil
}

func TestListNormalizesAppAndPaging(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	f := &fakeReader{}
	r.GET("/auditLogs", List(f))

	req := httptest.NewRequest(http.MethodGet, "/auditLogs?app=ems&page=2&pageSize=20", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if f.q.App != audit.AppEMS {
		t.Fatalf("app = %q, want %q", f.q.App, audit.AppEMS)
	}
	if f.q.Page != 2 || f.q.PageSize != 20 {
		t.Fatalf("page/pageSize = %d/%d", f.q.Page, f.q.PageSize)
	}
}
