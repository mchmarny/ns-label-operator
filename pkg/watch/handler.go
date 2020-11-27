package watch

import (
	corev1 "k8s.io/api/core/v1"
)

func (w *NsWatch) namespaceHandler(oldObj interface{}, newObj interface{}) {
	if newObj == nil {
		return
	}

	oldNs := oldObj.(*corev1.Namespace)
	newNs := newObj.(*corev1.Namespace)

	w.logger.Debugf("processing namespace: %s", newNs.GetName())

	// skip if new ns is being deleted
	if newNs.Status.Phase == corev1.NamespaceTerminating {
		return
	}

	labelValue, labelExists := newNs.GetLabels()[w.label]

	// skip if the new ns doesn't have the trigger label
	if !labelExists {
		w.logger.Debugf("no trigger (%s) in labels: %v", w.label, newNs.GetLabels())
		return
	}

	// exit if the old ns already had that label
	// no significant change, some other label is being applied
	if _, k := oldNs.GetLabels()[w.label]; k {
		w.logger.Debugf("trigger (%s) was already in labels: %v", w.label, oldNs.GetLabels())
		return
	}

	// skip if the trigger value is not true
	if labelValue != triggerValue {
		w.logger.Debugf("no trigger value in label: %s (want:%s, got:%s)", w.label, triggerValue, labelValue)
		return
	}

	w.logger.Debugf("triggering:%s on:%s in ns:%s", w.label, newNs.GetLabels(), newNs.GetName())
	if err := w.apply(newNs); err != nil {
		w.logger.Errorf("error running trigger %s on %s: %v", w.label, newNs.GetName(), err)
		return
	}

	w.logger.Infof("trigger:%s applied on namespace:%s", w.label, newNs.GetName())
}
