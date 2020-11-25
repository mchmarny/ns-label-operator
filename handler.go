package main

import (
	corev1 "k8s.io/api/core/v1"
)

func nsChangeHandler(oldObj interface{}, newObj interface{}) {
	oldNs := oldObj.(*corev1.Namespace)
	newNs := newObj.(*corev1.Namespace)

	// not sure if that's even possible
	if newNs == nil {
		return
	}

	logger.Infof("testing ns: %s\n", newNs.GetName())

	// skip if new ns is being deleted
	if newNs.Status.Phase == corev1.NamespaceTerminating {
		return
	}

	labelValue, labelExists := newNs.GetLabels()[triggerLabel]

	// skip if the new ns doesn't have the trigger label
	if !labelExists {
		logger.Debugf("no trigger (%s) in labels: %v", triggerLabel, newNs.GetLabels())
		return
	}

	// exit if the old ns already had that label
	// no significant change, some other label is being applied
	if _, k := oldNs.GetLabels()[triggerLabel]; k {
		logger.Debugf("trigger (%s) was already in labels: %v", triggerLabel, oldNs.GetLabels())
		return
	}

	// skip if the trigger value is not true
	if labelValue != triggerValue {
		logger.Debugf("no trigger value in label: %s (want:%s, got:%s)", triggerLabel, triggerValue, labelValue)
		return
	}

	logger.Infof("triggering (%s) on ns: %s (labels: %v)", triggerLabel, newNs.GetName(), newNs.GetLabels())
	if err := trigger.run(newNs); err != nil {
		logger.Errorf("error running trigger %s on %s: %v", triggerLabel, newNs.GetName(), err)
		return
	}

	logger.Infof("trigger %s applied on %s", triggerLabel, newNs.GetName())
}
