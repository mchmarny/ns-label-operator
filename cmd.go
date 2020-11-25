package main

import (
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	yamlPath = "/config.yaml"
)

type triggerCmd struct {
	clientSet *kubernetes.Clientset
	logger    *logrus.Logger
}

func (r *triggerCmd) run(ns *corev1.Namespace) error {
	r.logger.Debugf("ns: %+v, yaml: %s", ns, yamlPath)

	return nil
}
