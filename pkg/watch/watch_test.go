package watch

import (
	"os/user"
	"path"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func getLocalConfigPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", errors.Wrap(err, "error getting current user")
	}
	return path.Join(usr.HomeDir, ".kube/config"), nil
}

func getTestWatchInstance(t *testing.T) *NsWatch {
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
	return w
}

func TestNsWatch(t *testing.T) {
	f, err := getLocalConfigPath()
	if err != nil {
		t.Fatalf("error getting config path: %v", err)
	}

	conf := Config{}

	t.Run("with non existing directory", func(t *testing.T) {
		if _, err := getFiles("bad-dir", "*.yaml"); err == nil {
			t.Fatal()
		}
	})

	t.Run("with empty config", func(t *testing.T) {
		if _, err := NewNsWatch(conf); err == nil {
			t.Fatal()
		}
	})

	t.Run("without manifest dir", func(t *testing.T) {
		conf.Label = "test"
		if _, err := NewNsWatch(conf); err == nil {
			t.Fatal()
		}
	})

	t.Run("with valid config using dir", func(t *testing.T) {
		conf.ConfigFile = f
		conf.ManifestDir = "../../manifests"
		if _, err := NewNsWatch(conf); err != nil {
			t.Fatalf("error creating watch: %v", err)
		}
	})

	t.Run("with valid config using yaml", func(t *testing.T) {
		conf.ManifestDir = ""
		conf.Manifests = []string{
			`apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: secret-reader
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get"]`,
			`apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: dapr-secret-reader
subjects:
- kind: ServiceAccount
  name: default
roleRef:
  kind: Role
  name: secret-reader
  apiGroup: rbac.authorization.k8s.io`,
		}
		if _, err := NewNsWatch(conf); err != nil {
			t.Fatalf("error creating watch: %v", err)
		}
	})

	t.Run("with valid config using yaml", func(t *testing.T) {
		w, err := NewNsWatch(conf)
		if err != nil {
			t.Fatalf("error creating watch: %v", err)
		}
		defer w.Stop()
		go func() {
			if err := w.Start(); err != nil {
				panic(err)
			}
		}()
		time.Sleep(5 * time.Second)
		ns := getNS(true, map[string]string{})
		if err := w.apply(ns); err != nil {
			t.Fatalf("error applying test manifest: %v", err)
		}
	})
}
