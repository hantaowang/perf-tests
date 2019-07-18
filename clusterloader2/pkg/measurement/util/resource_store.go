/*
Copyright 2018 The Kubernetes Authors.

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

/*
This file is copy of https://github.com/kubernetes/kubernetes/blob/master/test/utils/pod_store.go
with slight changes regarding labelSelector and flagSelector applied.
*/

package util

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)


// PodStore is a convenient wrapper around cache.Store that returns list of v1.Pod instead of interface{}.
type ResourceStore struct {
	cache.Store
	stopCh    chan struct{}
	Reflector *cache.Reflector
}

// List returns to list of pods (that satisfy conditions provided to NewPodStore).
func (s *ResourceStore) List() []runtime.Object {
	return []runtime.Object(s.Store.List())
}

// Stop stops podstore watch.
func (s *ResourceStore) Stop() {
	close(s.stopCh)
}

