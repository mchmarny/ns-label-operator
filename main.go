package main

import (
	"log"
	"os"
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
	triggerValue     = "true"
	configDirDefault = "/config"
)

var (
	configPath   = getEnvVar("KUBECONFIG", "")
	triggerLabel = getEnvVar("TRIGGER_LABEL", "dapr-enabled")
	debug        = getEnvVar("DEBUG", "") == "true"
	logJSON      = getEnvVar("LOG_TO_JSON", "") == "true"
	dirPath      = getEnvVar("CONFIG_DIR", configDirDefault)
	fileManager  = getEnvVar("FILE_MANAGER", "ns-label-operator")

	logger  = getLogger(debug, logJSON)
	trigger *triggerCmd
)

func getConfig(file string) (cfg *rest.Config, err error) {
	if file == "" {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, errors.Wrap(err, "error loading in cluster config")
		}
	} else {
		cfg, err = cmd.BuildConfigFromFlags("", file)
		if err != nil {
			return nil, errors.Wrapf(err, "error loading config from: %s", file)
		}
	}
	return
}

func getLogger(debug, logJSON bool) *logrus.Logger {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	l.SetLevel(logrus.WarnLevel)
	if debug {
		l.SetLevel(logrus.TraceLevel)
	}
	if logJSON {
		l.SetFormatter(&logrus.JSONFormatter{})
	}
	return l
}

func main() {
	logger.Infof("loading client from %s...", configPath)
	cfg, err := getConfig(configPath)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("error creating client set from (%+v): %v", cfg, err)
	}

	trigger = &triggerCmd{
		cfg:         cfg,
		cs:          clientSet,
		logger:      logger,
		fileManager: fileManager,
	}

	if err := trigger.init(dirPath); err != nil {
		log.Fatalf("error initializing cmd: %v", err)
	}

	factory := informers.NewSharedInformerFactory(clientSet, 0) // 0 == don't sync
	informer := factory.Core().V1().Namespaces().Informer()
	stopper := make(chan struct{})
	defer close(stopper)
	defer runtime.HandleCrash()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: nsChangeHandler,
	})
	go informer.Run(stopper)
	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		runtime.HandleError(errors.New("timed out waiting for caches to sync"))
		return
	}
	<-stopper
}

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
