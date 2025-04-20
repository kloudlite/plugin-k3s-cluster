package v1

type AWSCredentials struct {
	AwsAccessKey string `json:"awsAccessKey,omitempty"`
	AwsSecretKey string `json:"awsSecretKey,omitempty"`
}
