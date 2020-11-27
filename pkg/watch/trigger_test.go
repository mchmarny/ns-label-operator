package watch

import (
	"context"
	"testing"
)

func TestTrigger(t *testing.T) {
	w := getTestWatchInstance(t)
	ctx := context.Background()

	t.Run("apply without ns", func(t *testing.T) {
		if err := w.apply(nil); err == nil {
			t.Fatal()
		}
	})

	t.Run("apply manifest without ns", func(t *testing.T) {
		if err := w.applyManifest(ctx, nil, ""); err == nil {
			t.Fatal()
		}
	})

	t.Run("apply manifest without yaml", func(t *testing.T) {
		ns := getNS(true, map[string]string{"test": "true"})
		if err := w.applyManifest(ctx, ns, ""); err == nil {
			t.Fatal()
		}
	})
}
