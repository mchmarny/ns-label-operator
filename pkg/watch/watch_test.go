package watch

import (
	"os"
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

	if _, err := os.Stat(f); err != nil && os.IsNotExist(err) {
		t.Logf("kube config file doesn't exists: %s", f)
		t.SkipNow() // TODO: kube config in hithub action
	}

	if _, err := NewNsWatch(nil, "dapr-enabled", f, "../../manifests"); err != nil {
		t.Fatalf("error creating watch: %v", err)
	}
}
