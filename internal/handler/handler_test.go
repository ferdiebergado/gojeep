package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ferdiebergado/goexpress"
	"github.com/ferdiebergado/goweb/internal/config"
	"github.com/ferdiebergado/goweb/internal/handler"
	"github.com/stretchr/testify/assert"
)

func TestHandlerHandleDashboard(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()

	mockCfg := config.TemplateConfig{
		Path:         "../../web/templates",
		LayoutFile:   "layout.html",
		PartialsPath: "partials",
		PagesPath:    "pages",
	}
	tmpl, err := handler.NewTemplate(mockCfg)
	if err != nil {
		t.Fatalf("cant parse template: %v", err)
	}
	h := handler.NewBaseHandler(tmpl)

	r := goexpress.New()
	r.Get("/dashboard", h.HandleDashboard)
	r.ServeHTTP(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode, "Status code should match")
	assert.Contains(t, rr.Body.String(), "Dashboard", "Body should contain the same text")
}
