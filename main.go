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
	operatorName     = "ns-label-operator"
)

var (
	configPath   = getEnvVar("KUBECONFIG", "")
	dirPath      = getEnvVar("CONFIG_DIR", configDirDefault)
	triggerLabel = getEnvVar("TRIGGER_LABEL", "dapr-enabled")
	debug        = getEnvVar("DEBUG", "false") == "true"
	logJSON      = getEnvVar("LOG_TO_JSON", "true") == "true"

	logger  = getLogger(debug, logJSON)
	trigger *triggerCmd
)

func getConfig(file string) (cfg *rest.Config, err error) {
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

func getLogger(debug, logJSON bool) *logrus.Logger {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	l.SetLevel(logrus.InfoLevel)
	if debug {
		l.SetLevel(logrus.TraceLevel)
	}
	if logJSON {
		l.SetFormatter(&logrus.JSONFormatter{})
	}
	return l
}

func main() {
	logger.Infof("starting %s", operatorName)
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
		fileManager: operatorName,
	}

	if err := trigger.init(dirPath); err != nil {
		log.Fatalf("error initializing cmd: %v", err)
	}

	factory := informers.NewSharedInformerFactory(clientSet, 0) // 0 == don't sync
	informer := factory.Core().V1().Namespaces().Informer()
	stopCh := make(chan struct{})
	defer close(stopCh)
	defer runtime.HandleCrash()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{UpdateFunc: nsChangeHandler})
	go informer.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(errors.New("timed out waiting for caches to sync"))
		return
	}
	<-stopCh
}

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
