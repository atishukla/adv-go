package model

import (
	"sync"

	v1 "k8s.io/api/core/v1"
)

// PodInfo struct to represent a Kubernetes Pod's basic information
type Pod struct {
	mu  sync.RWMutex
	pod v1.Pod
}

// Mutex for thread-safe logging
func NewPod(n *v1.Pod) *Pod {
	return &Pod{
		pod: *n,
	}
}

// Update updates the pod model, replacing it with a shallow copy of the provided pod
func (p *Pod) Update(pod *v1.Pod) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pod = *pod
}

func (p *Pod) IsScheduled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pod.Spec.NodeName != ""

}

// NodeName returns the node that the pod is scheduled against, or an empty string
func (p *Pod) NodeName() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pod.Spec.NodeName
}

// Name returns the name of the pod
func (p *Pod) Name() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pod.Name
}

// Phase returns the pod phase
func (p *Pod) Phase() v1.PodPhase {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pod.Status.Phase
}
