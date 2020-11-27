package main

import (
	"log"
	"os"
	"strings"

	"github.com/mchmarny/ns-label-operator/pkg/watch"
	"github.com/sirupsen/logrus"
)

var (
	configPath = getEnvVar("KUBECONFIG", "")
	dirPath    = getEnvVar("CONFIG_DIR", "")
	label      = getEnvVar("TRIGGER_LABEL", "")
	debug      = getEnvVar("DEBUG", "") == "true"
	logJSON    = getEnvVar("LOG_TO_JSON", "") == "true"
)

func main() {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
	if debug {
		logger.SetLevel(logrus.DebugLevel)
	}
	if logJSON {
		logger.SetFormatter(&logrus.JSONFormatter{
			DisableTimestamp: true,
			PrettyPrint:      true,
		})
	}

	nsw, err := watch.NewNsWatch(watch.Config{
		Label:       label,
		ConfigFile:  configPath,
		ManifestDir: dirPath,
		Logger:      logger,
	})

	if err != nil {
		log.Fatalf("error initializing watch: %v", err)
	}

	if err := nsw.Run(); err != nil {
		log.Fatalf("error running watch: %v", err)
	}
}

func getEnvVar(key, fallbackValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(val)
	}
	return fallbackValue
}
