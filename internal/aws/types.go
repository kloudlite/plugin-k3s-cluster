package aws

type Credentials struct {
	AccessKey string `json:"accessKey,omitempty"`
	SecretKey string `json:"secretKey,omitempty"`
}

type VPCPublicSubnet struct {
	AZ       string `json:"az"`
	SubnetID string `json:"subnetID"`
}

type VPC struct {
	ID            string            `json:"id"`
	PublicSubnets []VPCPublicSubnet `json:"publicSubnets"`
}

type Node struct {
	Name             string `json:"name"`
	AMI              string `json:"ami"`
	InstanceType     string `json:"instanceType"`
	RootVolumeSize   int    `json:"rootVolumeSize"`
	RootVolumeType   string `json:"rootVolumeType"`
	K3sVersion       string `json:"k3sVersion,omitempty"`
	AvailabilityZone string `json:"availabilityZone"`
}

type Region string

func (r *Region) String() string {
	return string(*r)
}
