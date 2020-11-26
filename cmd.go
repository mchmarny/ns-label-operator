package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type triggerCmd struct {
	cfg         *rest.Config
	cs          *kubernetes.Clientset
	logger      *logrus.Logger
	fileManager string
	yamlFiles   []string
}

func (r *triggerCmd) init(yamlPath string) error {
	r.logger.Debugf("file: %s", yamlPath)

	info, err := os.Stat(yamlPath)
	if os.IsNotExist(err) {
		return errors.Wrapf(err, "yaml file not found: %s", yamlPath)
	}

	if info.IsDir() {
		return errors.Wrapf(err, "%s is a directory, expected file", yamlPath)
	}

	b, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return errors.Wrapf(err, "error reading file: %s", yamlPath)
	}

	r.yamlFiles = make([]string, 0)
	for _, y := range strings.Split(string(b), "---") {
		r.yamlFiles = append(r.yamlFiles, y)
	}

	return nil
}

func (r *triggerCmd) run(ns *corev1.Namespace) error {
	if ns == nil {
		return errors.New("nil namespace")
	}
	r.logger.Debugf("ns: %s", ns.Name)

	ctx := context.Background()
	for _, y := range r.yamlFiles {
		if err := r.serverApply(ctx, ns, y); err != nil {
			r.logger.Errorf("error applying yaml (%s): %v", y, err)
			continue
		}
	}
	return nil
}

func (r *triggerCmd) serverApply(ctx context.Context, ns *corev1.Namespace, deploymentYAML string) error {
	if ns == nil {
		return errors.New("nil namespace")
	}
	if deploymentYAML == "" {
		return errors.New("empty deployment YAML vairable")
	}

	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(r.cfg)
	if err != nil {
		return errors.Wrapf(err, "error creating discovery client using: %v", r.cfg)
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(r.cfg)
	if err != nil {
		return errors.Wrapf(err, "error creating dynamic client using: %v", r.cfg)
	}

	// 3. Decode YAML manifest into unstructured.Unstructured
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode([]byte(deploymentYAML), nil, obj)
	if err != nil {
		return errors.Wrapf(err, "error decoding YAML: %v", deploymentYAML)
	}
	obj.SetNamespace(ns.GetName())
	r.logger.Infof("name:%s, ns:%s kind:%s, version:%s",
		obj.GetName(), obj.GetNamespace(), gvk.GroupKind(), gvk.Version)

	// 4. Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return errors.Wrapf(err, "error creating REST mapping: %v", gvk.GroupKind())
	}

	r.logger.Infof("resource: %v, scope: %v", mapping.Resource, mapping.Scope)

	// 5. Obtain REST interface for the GVR
	dr := dyn.Resource(mapping.Resource).Namespace(ns.Name)

	// 6. Marshal object into JSON
	data, err := json.Marshal(obj)
	if err != nil {
		return errors.Wrapf(err, "error marshaling object: %v", obj)
	}

	r.logger.Infof("patching %s in %s... ", obj.GetName(), obj.GetNamespace())

	// 7. Create or Update the object with SSA
	//     types.ApplyPatchType indicates SSA.
	//     FieldManager specifies the field owner ID.
	_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: r.fileManager,
	})

	if err != nil {
		return errors.Wrapf(err, "error applying %s to %s", string(data), obj.GetName())
	}

	return nil
}
