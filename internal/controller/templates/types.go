package templates

import (
	corev1 "k8s.io/api/core/v1"
)

type ClusterLifecycleSpecTemplateArgs struct {
	Tolerations   []corev1.Toleration
	NodeSelector  map[string]string
	JobImage      string
	CloudProvider string

	TFWorkspaceName      string
	TFWorkspaceNamespace string

	ValuesJSON string

	// OutputSecretName      string
	// OutputSecretNamespace string
}
