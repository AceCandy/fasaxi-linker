package task

import (
	"testing"

	"github.com/fasaxi-linker/servergo/pkg/core"
)

func TestServiceGetByID(t *testing.T) {
	s := &Service{
		tasks:    []Task{{ID: 1, Name: "alpha"}, {ID: 2, Name: "beta"}},
		tasksMap: make(map[int]Task),
		watchers: make(map[int]*core.Watcher),
	}

	s.rebuildMap()

	if _, ok := s.Get(1); !ok {
		t.Fatalf("expected task id 1 to exist")
	}
	if _, ok := s.Get(3); ok {
		t.Fatalf("did not expect task id 3 to exist")
	}
}

func TestServiceIsWatchingByID(t *testing.T) {
	s := &Service{
		watchers: map[int]*core.Watcher{1: nil},
	}

	if !s.IsWatching(1) {
		t.Fatalf("expected task id 1 to be watching")
	}
	if s.IsWatching(2) {
		t.Fatalf("did not expect task id 2 to be watching")
	}
}
