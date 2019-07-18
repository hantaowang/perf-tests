package util

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"
)


type Resource interface {
	NewStore(c clientset.Interface, namespace string, labelSelector string, fieldSelector string) (*ResourceStore, error)
	CreateStartUpStatus([]runtime.Object) ResourceStartUpStatus
	ComputeDiff([]runtime.Object, []runtime.Object) ResourceDiff
	StartUpComplete([]runtime.Object, ResourceStartUpStatus) bool
}

type ResourceStartUpStatus interface {
	RecomputeStartUpStatus([]runtime.Object)
	String() string
	TimeoutError(options *WaitForResourceOptions) error
}

type ResourceDiff interface {
	AddedResources() []string
	DeletedResources() []string
	String(ignorePhases sets.String) string
}

