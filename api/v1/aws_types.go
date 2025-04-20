package v1

// +kubebuilder:object:generate=true
type AwsCredentials struct {
	AccessKey string `json:"accessKey,omitempty"`
	SecretKey string `json:"secretKey,omitempty"`
}

// +kubebuilder:object:generate=true
type AwsVPCPublicSubnet struct {
	AZ       string `json:"az"`
	SubnetID string `json:"subnetID"`
}

// +kubebuilder:object:generate=true
type AwsVPC struct {
	ID            string               `json:"id"`
	PublicSubnets []AwsVPCPublicSubnet `json:"publicSubnets"`
}

// +kubebuilder:object:generate=true
type AwsNode struct {
	Name             string `json:"name"`
	AMI              string `json:"ami"`
	InstanceType     string `json:"instanceType"`
	RootVolumeSize   int    `json:"rootVolumeSize"`
	RootVolumeType   string `json:"rootVolumeType"`
	K3sVersion       string `json:"k3sVersion,omitempty"`
	AvailabilityZone string `json:"availabilityZone"`
}

// +kubebuilder:object:generate=true
type AwsRegion string

func (r *AwsRegion) String() string {
	return string(*r)
}
