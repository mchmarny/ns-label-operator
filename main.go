package main

import (
	"errors"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	triggerValue = "true"
)

var (
	configPath   = getEnvVar("KUBECONFIG", "/Users/mchmarny/.kube/config")
	triggerLabel = getEnvVar("TRIGGER", "dapr-enabled")
	debug        = getEnvVar("DEBUG", "") == "true"
	logJSON      = getEnvVar("LOGJSON", "") == "true"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
	if debug {
		log.SetLevel(log.TraceLevel)
	}
	if logJSON {
		log.SetFormatter(&log.JSONFormatter{})
	}

	log.Info("starting watcher...")
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		log.Fatalf("error building config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("error getting client: %v", err)
	}

	factory := informers.NewSharedInformerFactory(clientset, 0)
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
