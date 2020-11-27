package watch

import (
	"os/user"
	"path"
	"testing"

	"github.com/pkg/errors"
)

func getLocalConfigPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", errors.Wrap(err, "error getting current user")
	}
	return path.Join(usr.HomeDir, ".kube/config"), nil
}

func TestWatch(t *testing.T) {
	f, err := getLocalConfigPath()
	if err != nil {
		t.Fatalf("error getting config path: %v", err)
	}

	conf := Config{}

	t.Run("with empty config", func(t *testing.T) {
		if _, err := NewNsWatch(conf); err == nil {
			t.Fatal()
		}
	})

	t.Run("without manifest dir", func(t *testing.T) {
		conf.Label = "dapr-enabled"
		if _, err := NewNsWatch(conf); err == nil {
			t.Fatal()
		}
	})

	t.Run("with valid config", func(t *testing.T) {
		conf.Label = "dapr-enabled"
		conf.ConfigFile = f
		conf.ManifestDir = "../../manifests"
		if _, err := NewNsWatch(conf); err != nil {
			t.Fatalf("error creating watch: %v", err)
		}
	})
}
