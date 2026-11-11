package evidence

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// PodStatus is the A2 view of a pod, shared with the MCP tools so that the
// fixed baseline and the agent see exactly the same data.
func PodStatus(p *corev1.Pod) map[string]any { return podStatus(p) }

// Rollout is the A5 view (owner, revisions, template diff) for the pod's workload.
func Rollout(ctx context.Context, cs kubernetes.Interface, pod *corev1.Pod) (any, error) {
	return (&collector{cs: cs}).rollout(ctx, pod)
}

// Endpoints is the A8 view (services selecting the pod and their ready endpoints).
func Endpoints(ctx context.Context, cs kubernetes.Interface, pod *corev1.Pod) (any, error) {
	return (&collector{cs: cs}).endpoints(ctx, pod)
}
