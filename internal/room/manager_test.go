package room_test

import (
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
