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
	triggerValue = "true"
)

var (
	configPath   = getEnvVar("KUBECONFIG", "")
	triggerLabel = getEnvVar("TRIGGER", "dapr-enabled")
	debug        = getEnvVar("DEBUG", "") == "true"
	logJSON      = getEnvVar("LOGJSON", "") == "true"

	logger  = logrus.New()
	trigger *triggerCmd
)

func getConfig(file string) (conf *rest.Config, err error) {
	if file == "" {
		conf, err = rest.InClusterConfig()
		if err != nil {
			return nil, errors.Wrap(err, "error loading in cluster config")
		}
	} else {
		conf, err = cmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			return nil, errors.Wrapf(err, "error loading config from: %s", file)
		}
	}
	return
}

func main() {
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.WarnLevel)
	if debug {
		logger.SetLevel(logrus.TraceLevel)
	}
	if logJSON {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	logger.Infof("loading configuration... (KUBECONFIG=%s)", configPath)
	config, err := getConfig(configPath)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("error getting client: %v", err)
	}

	trigger = &triggerCmd{
		clientSet: clientset,
		logger:    logger,
	}

	factory := informers.NewSharedInformerFactory(clientset, 0) // 0 == don't sync
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
