package watch

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func getNS(active bool, labels map[string]string) *corev1.Namespace {
	ns := &corev1.Namespace{}
	ns.SetName("test")
	ns.SetLabels(labels)

	if active {
		ns.Status.Phase = corev1.NamespaceActive
	} else {
		ns.Status.Phase = corev1.NamespaceTerminating
	}

	return ns
}

func TestHandler(t *testing.T) {
	f, err := getLocalConfigPath()
	if err != nil {
		t.Fatalf("error getting config path: %v", err)
	}

	w, err := NewNsWatch(Config{
		Label:       "test",
		ConfigFile:  f,
		ManifestDir: "../../manifests",
	})
	if err != nil {
		t.Fatalf("error creating watch: %v", err)
	}

	t.Run("without triggering label", func(t *testing.T) {
		oldNs := getNS(true, map[string]string{})
		newNs := getNS(true, map[string]string{})
		if _, ok := w.shouldRun(oldNs, newNs); ok {
			t.Fatal()
		}
	})

	t.Run("with terminating status", func(t *testing.T) {
		oldNs := getNS(true, map[string]string{})
		newNs := getNS(false, map[string]string{})
		if _, ok := w.shouldRun(oldNs, newNs); ok {
			t.Fatal()
		}
	})

	t.Run("with triggering label not true", func(t *testing.T) {
		oldNs := getNS(true, map[string]string{})
		newNs := getNS(true, map[string]string{"test": "false"})
		if _, ok := w.shouldRun(oldNs, newNs); ok {
			t.Fatal()
		}
	})

	t.Run("with no new triggering label", func(t *testing.T) {
		oldNs := getNS(true, map[string]string{"test": "true"})
		newNs := getNS(true, map[string]string{"test": "true", "other": "true"})
		if _, ok := w.shouldRun(oldNs, newNs); ok {
			t.Fatal()
		}
	})

	t.Run("with new triggering label set true", func(t *testing.T) {
		oldNs := getNS(true, map[string]string{})
		newNs := getNS(true, map[string]string{"test": "true"})
		if _, ok := w.shouldRun(oldNs, newNs); !ok {
			t.Fatal()
		}
	})
}
