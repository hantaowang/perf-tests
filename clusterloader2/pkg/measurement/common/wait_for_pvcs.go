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

package common

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/perf-tests/clusterloader2/pkg/measurement"
	measurementutil "k8s.io/perf-tests/clusterloader2/pkg/measurement/util"
	"k8s.io/perf-tests/clusterloader2/pkg/util"
)

const (
	defaultWaitForPVCsTimeout         = 60 * time.Second
	defaultWaitForPVCsInterval        = 5 * time.Second
	waitForRunningPVCsMeasurementName = "WaitForRunningPVCs"
)

func init() {
	if err := measurement.Register(waitForRunningPVCsMeasurementName, createWaitForRunningPVCsMeasurement); err != nil {
		klog.Fatalf("Cannot register %s: %v", waitForRunningPVCsMeasurementName, err)
	}
}

func createWaitForRunningPVCsMeasurement() measurement.Measurement {
	return &waitForRunningPVCsMeasurement{}
}

type waitForRunningPVCsMeasurement struct{}

// Execute waits until desired number of PVCs are bound or until timeout happens.
// PVCs can be specified by field and/or label selectors.
// If namespace is not passed by parameter, all-namespace scope is assumed.
func (w *waitForRunningPVCsMeasurement) Execute(config *measurement.MeasurementConfig) ([]measurement.Summary, error) {
	desiredPVCCount, err := util.GetInt(config.Params, "desiredPVCCount")
	if err != nil {
		return nil, err
	}
	namespace, err := util.GetStringOrDefault(config.Params, "namespace", metav1.NamespaceAll)
	if err != nil {
		return nil, err
	}
	labelSelector, err := util.GetStringOrDefault(config.Params, "labelSelector", "")
	if err != nil {
		return nil, err
	}
	fieldSelector, err := util.GetStringOrDefault(config.Params, "fieldSelector", "")
	if err != nil {
		return nil, err
	}
	timeout, err := util.GetDurationOrDefault(config.Params, "timeout", defaultWaitForPVCsTimeout)
	if err != nil {
		return nil, err
	}

	stopCh := make(chan struct{})
	time.AfterFunc(timeout, func() {
		close(stopCh)
	})
	options := &measurementutil.WaitForPodOptions{
		Namespace:           namespace,
		LabelSelector:       labelSelector,
		FieldSelector:       fieldSelector,
		DesiredPodCount:     desiredPVCCount,
		EnableLogging:       true,
		CallerName:          w.String(),
		WaitForPodsInterval: defaultWaitForPVCsInterval,
	}
	return nil, measurementutil.WaitForPods(config.ClusterFramework.GetClientSets().GetClient(), stopCh, options)
}

// Dispose cleans up after the measurement.
func (*waitForRunningPVCsMeasurement) Dispose() {}

// String returns a string representation of the measurement.
func (*waitForRunningPVCsMeasurement) String() string {
	return waitForRunningPVCsMeasurementName
}