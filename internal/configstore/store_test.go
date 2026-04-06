package configstore

import (
	"sync"
	"testing"

	"github.com/johtani/smarthome/subcommand"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

func TestStore_GetSet(t *testing.T) {
	initial := subcommand.Config{
		Owntone:   owntone.Config{URL: "http://a"},
		Switchbot: switchbot.Config{Token: "t", Secret: "s"},
		Yamaha:    yamaha.Config{URL: "http://y"},
		Commands:  subcommand.NewCommands(),
	}
	s := New(initial)

	got := s.Get()
	if got.Owntone.URL != "http://a" {
		t.Fatalf("Get() initial url = %q, want %q", got.Owntone.URL, "http://a")
	}

	next := got
	next.Owntone.URL = "http://b"
	s.Set(next)

	got = s.Get()
	if got.Owntone.URL != "http://b" {
		t.Fatalf("Get() updated url = %q, want %q", got.Owntone.URL, "http://b")
	}
}

func TestStore_ConcurrentAccess(t *testing.T) {
	initial := subcommand.Config{
		Owntone:   owntone.Config{URL: "http://a"},
		Switchbot: switchbot.Config{Token: "t", Secret: "s"},
		Yamaha:    yamaha.Config{URL: "http://y"},
		Commands:  subcommand.NewCommands(),
	}
	s := New(initial)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cfg := s.Get()
			cfg.Owntone.URL = "http://updated"
			if i%2 == 0 {
				s.Set(cfg)
			} else {
				_ = s.Get()
			}
		}(i)
	}
	wg.Wait()
}
