package watch

import (
	"io/ioutil"
	"os"
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

// Config defines NewNsWatch creation configuration.
type Config struct {
	// Label name that when applied to namespace with `true` will trigger application of manifests (e.g. enable-role).
	Label string `yaml:"label" json:"label"`
	// ConfigFile is the path to Kube config (e.g. ~/.kube/config). When left empty, it assumes in-cluster configuration.
	ConfigFile string `yaml:"configFile" json:"config-file"`
	// Manifests is the list of YAML snippets that will applied when matching namespace label is found. When nil, ManifestDir is used to read and populate the YAML manifests. Either Manifests or ManifestDir must be set.
	Manifests []string `yaml:"-" json:"-"`
	// ManifestDir is an optional path to directory with YAML documents. When set, it will append to the list of Manifests. Either Manifests or ManifestDir must be set.
	ManifestDir string `yaml:"manifestDir" json:"manifest-dir"`
	// Logger is an optional preconfigured logger. When nil, it will be configured for os.Stdout with logrus.InfoLevel.
	Logger *logrus.Logger `yaml:"-" json:"-"`
}

// NewNsWatch creates a new instance of NsWatch
func NewNsWatch(c Config) (*NsWatch, error) {
	if c.Logger == nil {
		c.Logger = logrus.New()
		c.Logger.SetOutput(os.Stdout)
		c.Logger.SetLevel(logrus.InfoLevel)
	}

	if c.Label == "" {
		return nil, errors.New("label required")
	}

	if c.ManifestDir == "" && len(c.Manifests) < 1 {
		return nil, errors.New("either manifest directory or at elast one YAML manifest required")
	}

	cfg, err := getConfig(c.Logger, c.ConfigFile)
	if err != nil {
		return nil, errors.Wrapf(err, "error loading config: %s", c.ConfigFile)
	}

	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating client set from (%+v): %v", cfg, err)
	}

	w := &NsWatch{
		label:       c.Label,
		config:      cfg,
		client:      cs,
		logger:      c.Logger,
		fileManager: operatorName,
		manifests:   make([]string, 0),
	}

	if len(c.Manifests) > 0 {
		w.manifests = append(w.manifests, c.Manifests...)
	}

	if c.ManifestDir != "" {
		c.Logger.Debugf("loading manifests from directory: %s", c.ManifestDir)
		files, err := getFiles(c.ManifestDir, "*.yaml")
		if err != nil {
			return nil, errors.Wrapf(err, "error listing files from: %s", c.ManifestDir)
		}

		mfs, err := getYamlContents(files)
		if err != nil {
			return nil, errors.Wrapf(err, "error reading YAML from: %s", c.ManifestDir)
		}
		w.manifests = append(w.manifests, mfs...)
	}

	return w, nil
}

// NsWatch creates a namespace watch and executes provided YAML in that namespace is labeled with a specific label=true.
type NsWatch struct {
	label       string
	config      *rest.Config
	client      *kubernetes.Clientset
	logger      *logrus.Logger
	fileManager string
	manifests   []string
	stopCh      chan struct{}
}

// GetLabel returns the configured label for which the watch is monitoring
func (w *NsWatch) GetLabel() string {
	return w.label
}

// GetFileManager returns the active file manager used by this watch. All objects created by NsWatch are associated with this file manager.
func (w *NsWatch) GetFileManager() string {
	return w.fileManager
}

// Start starts NsWatch by attaching namespace event handler. This method blocks until either an internal error or the Stop() method is invoked.
func (w *NsWatch) Start() error {
	w.logger.Infof("starting %s for label: %s", operatorName, w.label)
	factory := informers.NewSharedInformerFactory(w.client, 0) // 0 == don't sync
	informer := factory.Core().V1().Namespaces().Informer()
	w.stopCh = make(chan struct{})
	defer runtime.HandleCrash()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{UpdateFunc: w.namespaceHandler})
	go informer.Run(w.stopCh)
	if !cache.WaitForCacheSync(w.stopCh, informer.HasSynced) {
		err := errors.New("timed out waiting for caches to sync")
		runtime.HandleError(err)
		return err
	}
	<-w.stopCh
	return nil
}

// Stop stops the watcher by closing the internal stop channel.
func (w *NsWatch) Stop() {
	w.logger.Infof("stopping %s", w.fileManager)
	close(w.stopCh)
}

// getConfig creates Kubernetes config either from the provided file or using the in cluster configuration option.
func getConfig(logger *logrus.Logger, file string) (cfg *rest.Config, err error) {
	if file == "" {
		logger.Debug("using cluster config")
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, errors.Wrap(err, "error loading in cluster config")
		}
	} else {
		logger.Debugf("using config from: %s", file)
		cfg, err = cmd.BuildConfigFromFlags("", file)
		if err != nil {
			return nil, errors.Wrapf(err, "error loading config from: %s", file)
		}
	}
	return
}

// getYamlContents reads the content of passed in file paths. Splits on `---` in case of multiple YAML objects in a single file.
func getYamlContents(files []string) (out []string, err error) {
	for _, f := range files {
		b, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, errors.Wrapf(err, "error reading manifest: %s", f)
		}
		out = append(out, strings.Split(string(b), "---")...)
	}
	return
}

// getFiles finds all files matching given pattern. Kubernetes configmap mounts include version subdirectories this only returns the files in the passed dir path.
func getFiles(dir, pattern string) (out []string, err error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading dir: %s", dir)
	}
	for _, f := range files {
		if matched, err := filepath.Match(pattern, f.Name()); err != nil {
			return nil, errors.Wrapf(err, "error matching file: %s", f)
		} else if matched {
			out = append(out, path.Join(dir, f.Name()))
		}
	}
	return
}
