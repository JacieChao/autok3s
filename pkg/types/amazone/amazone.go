package amazone

type Options struct {
	AccessKey     string `json:"access-key,omitempty" yaml:"access-key,omitempty"`
	SecretKey     string `json:"secret-key,omitempty" yaml:"secret-key,omitempty"`
	Endpoint      string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	SessionToken  string `json:"session-token,omitempty" yaml:"session-token,omitempty"`
	Region        string `json:"region,omitempty" yaml:"region,omitempty"`
	AMI           string `json:"ami,omitempty" yaml:"ami,omitempty"`
	KeypairName   string `json:"keypair-name,omitempty" yaml:"keypair-name,omitempty"`
	InstanceType  string `json:"instance-type,omitempty" yaml:"instance-type,omitempty"`
	SecurityGroup string `json:"security-group,omitempty" yaml:"security-group,omitempty"`
	RootSize      string `json:"root-size,omitempty" yaml:"root-size,omitempty"`
	VolumeType    string `json:"volume-type,omitempty" yaml:"volume-type,omitempty"`
	VpcId         string `json:"vpc-id,omitempty" yaml:"vpc-id,omitempty"`
	SubnetId      string `json:"subnet-id,omitempty" yaml:"subnet-id,omitempty"`
	Zone          string `json:"zone,omitempty" yaml:"zone,omitempty"`
}
