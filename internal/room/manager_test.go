package room_test

import (
	"fmt"
	"testing"

	"github.com/qduong42/2d-game-tower-climb/internal/room"
)

func TestManager_GetOrCreate_SameCode(t *testing.T) {
	m := room.NewManager()
	r1 := m.GetOrCreate("ABCD")
	r2 := m.GetOrCreate("ABCD")
	if r1 != r2 {
		t.Error("expected same room for same code")
	}
}

func TestManager_GetOrCreate_DifferentCode(t *testing.T) {
	m := room.NewManager()
	r1 := m.GetOrCreate("ABCD")
	r2 := m.GetOrCreate("XYZ1")
	if r1 == r2 {
		t.Error("expected different rooms for different codes")
	}
}

func TestManager_Remove(t *testing.T) {
	m := room.NewManager()
	m.GetOrCreate("ABCD")
	m.Remove("ABCD")
	r := m.GetOrCreate("ABCD")
	if r == nil {
		t.Error("expected new room after remove")
	}
}

// TestListPublicOpen_ExcludesPrivateRoom verifies that a private room does not
// appear in the room browser listing.
func TestListPublicOpen_ExcludesPrivateRoom(t *testing.T) {
	m := room.NewManager()
	m.GetOrCreateWithPrivacy("PRIV", true)

	rooms := m.ListPublicOpen()
	for _, r := range rooms {
		if r.Code == "PRIV" {
			t.Error("private room should not appear in ListPublicOpen")
		}
	}
}

// TestListPublicOpen_IncludesPublicRoom verifies that a public, non-full room
// does appear in the room browser listing.
func TestListPublicOpen_IncludesPublicRoom(t *testing.T) {
	m := room.NewManager()
	m.GetOrCreate("PUB1")

	rooms := m.ListPublicOpen()
	found := false
	for _, r := range rooms {
		if r.Code == "PUB1" {
			found = true
		}
	}
	if !found {
		t.Error("public non-full room should appear in ListPublicOpen")
	}
}

// TestListPublicOpen_ExcludesFullRoom verifies that a room at capacity (3 players)
// does not appear in the room browser listing.
func TestListPublicOpen_ExcludesFullRoom(t *testing.T) {
	m := room.NewManager()
	rm := m.GetOrCreate("FULL")
	go rm.RunForTest()
	defer rm.Stop()

	// Fill the room to MaxPlayers (3). Join blocks until the room goroutine
	// processes the request, so player count is accurate afterwards.
	for i := 0; i < 3; i++ {
		c := room.NewTestClient(fmt.Sprintf("p%d", i), fmt.Sprintf("player%d", i))
		if ok := rm.Join(c, "#ffffff"); !ok {
			t.Fatalf("failed to join player %d", i)
		}
	}

	rooms := m.ListPublicOpen()
	for _, r := range rooms {
		if r.Code == "FULL" {
			t.Error("full room (3/3) should not appear in ListPublicOpen")
		}
	}
}
