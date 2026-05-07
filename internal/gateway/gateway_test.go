package gateway_test

import (
	"encoding/json"
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

// TestServeRooms_PublicRoomAppears verifies that a public, non-full room is
// returned by GET /rooms.
func TestServeRooms_PublicRoomAppears(t *testing.T) {
	mgr := room.NewManager()
	mgr.GetOrCreate("PUB1")
	gw := gateway.New(mgr)

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	w := httptest.NewRecorder()
	gw.ServeRooms(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var rooms []room.RoomInfo
	if err := json.NewDecoder(w.Body).Decode(&rooms); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	found := false
	for _, r := range rooms {
		if r.Code == "PUB1" {
			found = true
		}
	}
	if !found {
		t.Error("expected PUB1 in /rooms response")
	}
}

// TestServeRooms_PrivateRoomHidden verifies that a private room does not appear
// in the GET /rooms response.
func TestServeRooms_PrivateRoomHidden(t *testing.T) {
	mgr := room.NewManager()
	mgr.GetOrCreateWithPrivacy("PRIV", true)
	gw := gateway.New(mgr)

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	w := httptest.NewRecorder()
	gw.ServeRooms(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var rooms []room.RoomInfo
	if err := json.NewDecoder(w.Body).Decode(&rooms); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	for _, r := range rooms {
		if r.Code == "PRIV" {
			t.Error("private room should not appear in /rooms response")
		}
	}
}

// TestServeRooms_MethodNotAllowed verifies that non-GET requests to /rooms
// are rejected.
func TestServeRooms_MethodNotAllowed(t *testing.T) {
	mgr := room.NewManager()
	gw := gateway.New(mgr)

	req := httptest.NewRequest(http.MethodPost, "/rooms", nil)
	w := httptest.NewRecorder()
	gw.ServeRooms(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
