package gateway_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qduong42/2d-game-tower-climb/internal/gateway"
	"github.com/qduong42/2d-game-tower-climb/internal/room"
)

func TestGateway_RejectsNonWebSocket(t *testing.T) {
	mgr := room.NewManager()
	gw := gateway.New(mgr)

	req := httptest.NewRequest(http.MethodGet, "/r/ABCD", nil)
	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Code == http.StatusSwitchingProtocols {
		t.Error("expected non-101, got 101")
	}
}

func TestGateway_RejectsMissingRoomCode(t *testing.T) {
	mgr := room.NewManager()
	gw := gateway.New(mgr)

	req := httptest.NewRequest(http.MethodGet, "/r/", nil)
	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGateway_ExtractCode(t *testing.T) {
	cases := []struct {
		path string
		want string
		ok   bool
	}{
		{"/r/ABCD", "ABCD", true},
		{"/r/", "", false},
		{"/r", "", false},
		{"/other/ABCD", "", false},
	}
	for _, tc := range cases {
		code, ok := gateway.ExtractRoomCode(tc.path)
		if ok != tc.ok || code != tc.want {
			t.Errorf("path=%q: got (%q,%v) want (%q,%v)", tc.path, code, ok, tc.want, tc.ok)
		}
	}
}

var _ = strings.TrimPrefix // avoid unused import
