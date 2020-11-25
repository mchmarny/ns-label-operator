package main

import (
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

func nsChangeHandler(oldObj interface{}, newObj interface{}) {
	oldNs := oldObj.(*corev1.Namespace)
	newNs := newObj.(*corev1.Namespace)

	// not sure if that's even possible
	if newNs == nil {
		return
	}

	log.Infof("testing ns: %s\n", newNs.GetName())

	// skip if new ns is being deleted
	if newNs.Status.Phase == corev1.NamespaceTerminating {
		return
	}

	labelValue, labelExists := newNs.GetLabels()[triggerLabel]

	// skip if the new ns doesn't have the trigger label
	if !labelExists {
		log.Debugf("no trigger (%s) in labels: %v", triggerLabel, newNs.GetLabels())
		return
	}

	// exit if the old ns already had that label
	// no significant change, some other label is being applied
	if _, k := oldNs.GetLabels()[triggerLabel]; k {
		log.Debugf("trigger (%s) was already in labels: %v", triggerLabel, oldNs.GetLabels())
		return
	}

	// skip if the trigger value is not true
	if labelValue != triggerValue {
		log.Debugf("no trigger value in label: %s (want:%s, got:%s)", triggerLabel, triggerValue, labelValue)
		return
	}

	log.Infof("triggering (%s) on ns: %s (labels: %v)", triggerLabel, newNs.GetName(), newNs.GetLabels())
}
