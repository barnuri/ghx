package ghuser

import (
	"os"
	"testing"
	"time"

	"github.com/barnuri/ghx/internal/config"
)

func TestCacheRoundTrip(t *testing.T) {
	file := cachePath(t.TempDir(), config.Profile{User: "me"})
	if _, ok := readCache(file); ok {
		t.Fatal("expected miss before write")
	}
	want := config.Identity{Name: "Me", Email: "me@x.com"}
	writeCache(file, want)
	if got, ok := readCache(file); !ok || got != want {
		t.Fatalf("got %+v, %v", got, ok)
	}
}

func TestCacheExpires(t *testing.T) {
	file := cachePath(t.TempDir(), config.Profile{User: "me"})
	writeCache(file, config.Identity{Name: "Me"})
	expired := time.Now().Add(-cacheTTL - time.Minute)
	if err := os.Chtimes(file, expired, expired); err != nil {
		t.Fatal(err)
	}
	if _, ok := readCache(file); ok {
		t.Fatal("expected miss after TTL")
	}
}
