package ginmw

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"

	"harns-platform/pkg/audit"
	"harns-platform/pkg/generic"
)

type fakeRecorder struct {
	mu     sync.Mutex
	events []audit.Event
	err    error
}

func (f *fakeRecorder) Record(_ context.Context, e audit.Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, e)
	return f.err
}

func (f *fakeRecorder) RecordBatch(ctx context.Context, events []audit.Event) error {
	for _, e := range events {
		if err := f.Record(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeRecorder) Close(context.Context) error { return nil }

func (f *fakeRecorder) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.events)
}

func (f *fakeRecorder) last() audit.Event {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.events[len(f.events)-1]
}

func testRouter(f *fakeRecorder, user func(*gin.Context) (string, string)) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(generic.Trace())
	r.Use(Middleware(Options{
		Recorder:     f,
		Enabled:      true,
		MaxBodyBytes: 1024,
		ExcludePaths: []string{"/skip"},
		User:         user,
	}))
	r.POST("/test", WithOperation(audit.OpCreate, "创建"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok"})
	})
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 200})
	})
	r.POST("/skip", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 200})
	})
	r.POST("/fallback", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 200})
	})
	return r
}

func TestMiddlewareRecordsWriteRequest(t *testing.T) {
	f := &fakeRecorder{}
	r := testRouter(f, func(c *gin.Context) (string, string) {
		return "u1", "Alice"
	})

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Trace-Id", "trace-1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if f.count() != 1 {
		t.Fatalf("events = %d, want 1", f.count())
	}
	e := f.last()
	if e.TraceID != "trace-1" {
		t.Fatalf("traceId = %q", e.TraceID)
	}
	if e.OperationType != audit.OpCreate || e.OperationDesc != "创建" {
		t.Fatalf("operation = %q/%q", e.OperationType, e.OperationDesc)
	}
	if e.UserID != "u1" || e.UserName != "Alice" {
		t.Fatalf("user = %q/%q", e.UserID, e.UserName)
	}
	if e.Status != http.StatusOK || e.ResultCode != 200 {
		t.Fatalf("status/result = %d/%d", e.Status, e.ResultCode)
	}
	if !strings.Contains(e.RequestBody, "name") {
		t.Fatalf("request body = %q", e.RequestBody)
	}
	if !strings.Contains(e.ResponseBody, "ok") {
		t.Fatalf("response body = %q", e.ResponseBody)
	}
}

func TestMiddlewareSkipsGetAndExcludedPaths(t *testing.T) {
	f := &fakeRecorder{}
	r := testRouter(f, nil)

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/test", nil))
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/skip", nil))

	if f.count() != 0 {
		t.Fatalf("events = %d, want 0", f.count())
	}
}

func TestMiddlewareFallbackOperation(t *testing.T) {
	f := &fakeRecorder{}
	r := testRouter(f, nil)

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/fallback", nil))

	if f.count() != 1 {
		t.Fatalf("events = %d, want 1", f.count())
	}
	if got := f.last().OperationType; got != audit.OpCreate {
		t.Fatalf("operation = %q, want %q", got, audit.OpCreate)
	}
}

func TestMiddlewareAuditFailureDoesNotAffectResponse(t *testing.T) {
	f := &fakeRecorder{err: errors.New("queue full")}
	r := testRouter(f, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/test", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestTwoServicesShareTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	platformRecorder := &fakeRecorder{}
	platformRouter := gin.New()
	platformRouter.Use(generic.Trace())
	platformRouter.Use(Middleware(Options{Recorder: platformRecorder, Enabled: true}))
	platformRouter.POST("/platform", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 200})
	})
	platformServer := httptest.NewServer(platformRouter)
	defer platformServer.Close()

	emsRecorder := &fakeRecorder{}
	emsRouter := gin.New()
	emsRouter.Use(generic.Trace())
	emsRouter.Use(Middleware(Options{Recorder: emsRecorder, Enabled: true}))
	emsRouter.POST("/ems", func(c *gin.Context) {
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, platformServer.URL+"/platform", nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500})
			return
		}
		generic.SetTraceHeader(c.Request.Context(), req)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"code": 502})
			return
		}
		_ = resp.Body.Close()
		c.JSON(http.StatusOK, gin.H{"code": 200})
	})

	req := httptest.NewRequest(http.MethodPost, "/ems", nil)
	req.Header.Set("X-Trace-Id", "trace-shared")
	w := httptest.NewRecorder()
	emsRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if emsRecorder.count() != 1 || platformRecorder.count() != 1 {
		t.Fatalf("records = ems:%d platform:%d, want 1/1", emsRecorder.count(), platformRecorder.count())
	}
	if got := emsRecorder.last().TraceID; got != "trace-shared" {
		t.Fatalf("ems traceId = %q", got)
	}
	if got := platformRecorder.last().TraceID; got != "trace-shared" {
		t.Fatalf("platform traceId = %q", got)
	}
}
