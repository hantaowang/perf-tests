/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const (
	uninitialized = iota
	up
	down
	none
)

// WaitForPodOptions is an options used by WaitForPods methods.
type WaitForResourceOptions struct {
	Type								string
	Namespace           string
	LabelSelector       string
	FieldSelector       string
	DesiredPodCount     int
	EnableLogging       bool
	CallerName          string
	WaitForPodsInterval time.Duration
	Resource						Resource
}

// WaitForPods waits till disire nuber of pods is running.
// Pods are be specified by namespace, field and/or label selectors.
// If stopCh is closed before all pods are running, the error will be returned.
func WaitForResource(clientSet clientset.Interface, stopCh <-chan struct{}, options *WaitForResourceOptions) error {
	// TODO(#269): Change to shared podStore.
	store, err := options.Resource.NewStore(clientSet, options.Namespace, options.LabelSelector, options.FieldSelector)
	if err != nil {
		return fmt.Errorf("store creation error: %v", err)
	}
	defer store.Stop()

	startUpStatus := options.Resource.CreateStartUpStatus([]runtime.Object{})
	selectorsString := CreateSelectorsString(options.Namespace, options.LabelSelector, options.FieldSelector)
	scaling := uninitialized
	var oldResources []runtime.Object
	for {
		select {
		case <-stopCh:
			return startUpStatus.TimeoutError(options)
		case <-time.After(options.WaitForPodsInterval):
			resources := store.List()
			startUpStatus.RecomputeStartUpStatus(resources)
			if scaling != uninitialized {
				diff := options.Resource.ComputeDiff(oldResources, resources)
				deletedResources := diff.DeletedResources()
				if scaling != down && len(deletedResources) > 0 {
					klog.Errorf("%s: %s: %d %s disappeared: %v", options.CallerName, selectorsString, len(deletedResources), options.Type, strings.Join(deletedResources, ", "))
					klog.Infof("%s: %v", options.CallerName, diff.String(sets.NewString()))
				}
				addedResources := diff.AddedResources()
				if scaling != up && len(addedResources) > 0 {
					klog.Errorf("%s: %s: %d %s appeared: %v", options.CallerName, selectorsString, len(addedResources), options.Type, strings.Join(addedResources, ", "))
					klog.Infof("%s: %v", options.CallerName, diff.String(sets.NewString()))
				}
			} else {
				switch {
				case len(resources) == options.DesiredPodCount:
					scaling = none
				case len(resources) < options.DesiredPodCount:
					scaling = up
				case len(resources) > options.DesiredPodCount:
					scaling = down
				}
			}
			if options.EnableLogging {
				klog.Infof("%s: %s: %s", options.CallerName, selectorsString, startUpStatus.String())
			}
			// We allow inactive pods (e.g. eviction happened).
			// We wait until there is a desired number of pods running and all other pods are inactive.
			if options.Resource.StartUpComplete(resources, startUpStatus) {
				return nil
			}
			oldResources = resources
		}
	}
}
