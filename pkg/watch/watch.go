package watch

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	cmd "k8s.io/client-go/tools/clientcmd"
)

const (
	triggerValue = "true"
	operatorName = "ns-label-operator"
)

// NewNsWatch creates new NsWatch
func NewNsWatch(logger *logrus.Logger, label, configPath, manifestDir string) (*NsWatch, error) {
	if logger == nil {
		logger = logrus.New()
	}

	if label == "" {
		return nil, errors.New("label dir required")
	}

	if manifestDir == "" {
		return nil, errors.New("manifest dir required")
	}

	logger.Debugf("loading config from: %s", configPath)
	cfg, err := getConfig(logger, configPath)
	if err != nil {
		return nil, errors.Wrapf(err, "error loading config: %s", configPath)
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating client set from (%+v): %v", cfg, err)
	}

	logger.Debugf("loading manifests from: %s", manifestDir)
	files, err := getFiles(manifestDir, "*.yaml")
	if err != nil {
		return nil, errors.Wrapf(err, "error reading files from: %s", manifestDir)
	}

	w := &NsWatch{
		label:       label,
		cfg:         cfg,
		cs:          cs,
		logger:      logger,
		fileManager: operatorName,
	}

	w.manifests = make([]string, 0)
	for _, f := range files {
		w.logger.Infof("parsing %s file", f)
		b, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, errors.Wrapf(err, "error reading manifest: %s", f)
		}
		w.manifests = append(w.manifests, strings.Split(string(b), "---")...)
	}

	w.logger.Infof("found %d YAML manifest(s) from %d file(s)", len(w.manifests), len(files))

	return w, nil
}

// NsWatch executes YAML on namespaces with specific label
type NsWatch struct {
	label       string
	cfg         *rest.Config
	cs          *kubernetes.Clientset
	logger      *logrus.Logger
	fileManager string
	manifests   []string
	stopCh      chan struct{}
}

// Run starts NsWatch and blocks
func (w *NsWatch) Run() error {
	w.logger.Infof("starting %s for %s label", operatorName, w.label)

	factory := informers.NewSharedInformerFactory(w.cs, 0) // 0 == don't sync
	informer := factory.Core().V1().Namespaces().Informer()
	w.stopCh = make(chan struct{})
	defer runtime.HandleCrash()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{UpdateFunc: w.labelHandler})
	go informer.Run(w.stopCh)
	if !cache.WaitForCacheSync(w.stopCh, informer.HasSynced) {
		err := errors.New("timed out waiting for caches to sync")
		runtime.HandleError(err)
		return err
	}
	<-w.stopCh
	return nil
}

// Stop stops the watcher
func (w *NsWatch) Stop() {
	w.logger.Infof("stopping %s", w.fileManager)
	close(w.stopCh)
}

func getConfig(logger *logrus.Logger, file string) (cfg *rest.Config, err error) {
	if file == "" {
		logger.Info("using cluster config")
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, errors.Wrap(err, "error loading in cluster config")
		}
	} else {
		logger.Infof("using: %s", file)
		cfg, err = cmd.BuildConfigFromFlags("", file)
		if err != nil {
			return nil, errors.Wrapf(err, "error loading config from: %s", file)
		}
	}
	return
}

// k8s configmap mounts include version subdirectories
// so no walking down, just list the top dir files
func getFiles(dir, pattern string) ([]string, error) {
	var matches []string
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading dir: %s", dir)
	}
	for _, f := range files {
		if matched, err := filepath.Match(pattern, f.Name()); err != nil {
			return nil, errors.Wrapf(err, "error matching file: %s", f)
		} else if matched {
			matches = append(matches, path.Join(dir, f.Name()))
		}
	}
	return matches, nil
}
