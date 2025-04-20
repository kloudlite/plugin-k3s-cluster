/*
Copyright 2025.

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

package v1

import (
	"github.com/kloudlite/operator/toolkit/reconciler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=aws;
type CloudProvider string

const (
	CloudProviderAWS CloudProvider = "aws"
)

func (c CloudProvider) String() string {
	return string(c)
}

// +kubebuilder:validation:Enum=running;stopped;
type ClusterState string

const (
	ClusterStateRunning ClusterState = "running"
	ClusterStateStopped ClusterState = "stopped"
)

type AWS struct {
	Credentials *AwsCredentials `json:"credentials,omitempty"`
	Region      AwsRegion       `json:"region"`
	VPC         *AwsVPC         `json:"vpc,omitempty"`

	MasterNodes []AwsNode `json:"masterNodes"`
}

// K3sClusterSpec defines the desired state of K3sCluster.
type K3sClusterSpec struct {
	CloudProvider CloudProvider `json:"cloudProvider"`
	AWS           *AWS          `json:"aws,omitempty"`

	ClusterState ClusterState `json:"clusterState"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// K3sCluster is the Schema for the k3sclusters API.
type K3sCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   K3sClusterSpec    `json:"spec,omitempty"`
	Status reconciler.Status `json:"status,omitempty"`
}

func (s *K3sCluster) EnsureGVK() {
	if s != nil {
		s.SetGroupVersionKind(GroupVersion.WithKind("K3sCluster"))
	}
}

func (s *K3sCluster) GetStatus() *reconciler.Status {
	return &s.Status
}

func (s *K3sCluster) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (s *K3sCluster) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

// +kubebuilder:object:root=true

// K3sClusterList contains a list of K3sCluster.
type K3sClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []K3sCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&K3sCluster{}, &K3sClusterList{})
}
