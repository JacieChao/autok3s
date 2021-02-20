package aws

import (
	"github.com/cnrancher/autok3s/pkg/types"
	"github.com/cnrancher/autok3s/pkg/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const createUsageExample = `  autok3s -d create \
    --provider aws \
    --name <cluster name> \
    --access-key <access-key> \
    --secret-key <secret-key> \
    --master 1
`

const joinUsageExample = `  autok3s -d join \
    --provider aws \
    --name <cluster name> \
    --access-key <access-key> \
    --secret-key <secret-key> \
    --worker 1
`

const deleteUsageExample = `  autok3s -d delete \
    --provider aws \
    --name <cluster name>
    --access-key <access-key> \
    --secret-key <secret-key> 
`

const sshUsageExample = `  autok3s ssh \
    --provider aws \
    --name <cluster name> \
    --region <region> \
    --access-key <access-key> \
    --secret-key <secret-key>
`

func (p *Amazon) GetUsageExample(action string) string {
	switch action {
	case "create":
		return createUsageExample
	case "join":
		return joinUsageExample
	case "delete":
		return deleteUsageExample
	case "ssh":
		return sshUsageExample
	default:
		return ""
	}
}

func (p *Amazon) GetOptionFlags() []types.Flag {
	fs := p.GetCredentialFlags()
	fs = append(fs, p.sharedFlags()...)
	fs = append(fs, []types.Flag{
		{
			Name:  "ui",
			P:     &p.UI,
			V:     p.UI,
			Usage: "Enable K3s UI.",
		},
		{
			Name:  "cluster",
			P:     &p.Cluster,
			V:     p.Cluster,
			Usage: "Form k3s cluster using embedded etcd (requires K8s >= 1.19)",
		},
	}...)
	return fs
}

func (p *Amazon) GetDeleteFlags() []types.Flag {
	return []types.Flag{
		{
			Name:   "region",
			P:      &p.Region,
			V:      p.Region,
			Usage:  "AWS region",
			EnvVar: "AWS_DEFAULT_REGION",
		},
	}
}

func (p *Amazon) GetJoinFlags(cmd *cobra.Command) *pflag.FlagSet {
	fs := p.sharedFlags()
	return utils.ConvertFlags(cmd, fs)
}

func (p *Amazon) GetSSHFlags(cmd *cobra.Command) *pflag.FlagSet {
	fs := []types.Flag{
		{
			Name:      "name",
			P:         &p.Name,
			V:         p.Name,
			Usage:     "Set the name of the kubeconfig context",
			ShortHand: "n",
			Required:  true,
		},
		{
			Name:   "region",
			P:      &p.Region,
			V:      p.Region,
			Usage:  "AWS region",
			EnvVar: "AWS_DEFAULT_REGION",
		},
	}

	return utils.ConvertFlags(cmd, fs)
}

func (p *Amazon) GetCredentialFlags() []types.Flag {
	fs := []types.Flag{
		{
			Name:     "access-key",
			P:        &p.AccessKey,
			V:        p.AccessKey,
			Usage:    "AWS access key",
			Required: true,
			EnvVar:   "AWS_ACCESS_KEY_ID",
		},
		{
			Name:     "secret-key",
			P:        &p.SecretKey,
			V:        p.SecretKey,
			Usage:    "AWS secret key",
			Required: true,
			EnvVar:   "AWS_SECRET_ACCESS_KEY",
		},
	}

	return fs
}

func (p *Amazon) GetSSHConfig() *types.SSH {
	ssh := &types.SSH{
		User: defaultUser,
		Port: "22",
	}
	return ssh
}

func (p *Amazon) BindCredentialFlags() *pflag.FlagSet {
	nfs := pflag.NewFlagSet("", pflag.ContinueOnError)
	nfs.StringVar(&p.AccessKey, "access-key", p.AccessKey, "AWS access key")
	nfs.StringVar(&p.SecretKey, "secret-key", p.SecretKey, "AWS secret key")
	return nfs
}

func (p *Amazon) MergeClusterOptions(matched *types.Cluster) {
	p.Metadata = matched.Metadata
	p.Status = matched.Status
}

func (p *Amazon) sharedFlags() []types.Flag {
	fs := []types.Flag{
		{
			Name:   "region",
			P:      &p.Region,
			V:      p.Region,
			Usage:  "AWS region",
			EnvVar: "AWS_DEFAULT_REGION",
		},
		{
			Name:   "zone",
			P:      &p.Zone,
			V:      p.Zone,
			Usage:  "AWS zone",
			EnvVar: "AWS_ZONE",
		},
		{
			Name:   "keypair-name",
			P:      &p.KeypairName,
			V:      p.KeypairName,
			Usage:  "AWS keypair to use connect to instance",
			EnvVar: "AWS_KEYPAIR_NAME",
		},
		{
			Name:   "ami",
			P:      &p.AMI,
			V:      p.AMI,
			Usage:  "Used to specify the image to be used by the instance",
			EnvVar: "AWS_AMI",
		},
		{
			Name:   "instance-type",
			P:      &p.InstanceType,
			V:      p.InstanceType,
			Usage:  "Used to specify the type to be used by the instance",
			EnvVar: "AWS_INSTANCE_TYPE",
		},
		{
			Name:   "vpc-id",
			P:      &p.VpcID,
			V:      p.VpcID,
			Usage:  "AWS VPC id",
			EnvVar: "AWS_VPC_ID",
		},
		{
			Name:   "subnet-id",
			P:      &p.SubnetID,
			V:      p.SubnetID,
			Usage:  "AWS VPC subnet id",
			EnvVar: "AWS_SUBNET_ID",
		},
		{
			Name:   "volume-type",
			P:      &p.VolumeType,
			V:      p.VolumeType,
			Usage:  "Used to specify the EBS volume type",
			EnvVar: "AWS_VOLUME_TYPE",
		},
		{
			Name:   "root-size",
			P:      &p.RootSize,
			V:      p.RootSize,
			Usage:  "Used to specify the root disk size used by the instance (in GB)",
			EnvVar: "AWS_ROOT_SIZE",
		},
		{
			Name:   "security-group",
			P:      &p.SecurityGroup,
			V:      p.SecurityGroup,
			Usage:  "Used to specify the security group used by the instance",
			EnvVar: "AWS_SECURITY_GROUP",
		},
		{
			Name:  "iam-instance-profile-control",
			P:     &p.IamInstanceProfileForControl,
			V:     p.IamInstanceProfileForControl,
			Usage: "AWS IAM Instance Profile for k3s control nodes to deploy AWS Cloud Provider, must set with --cloud-controller-manager",
		},
		{
			Name:  "iam-instance-profile-worker",
			P:     &p.IamInstanceProfileForWorker,
			V:     p.IamInstanceProfileForWorker,
			Usage: "AWS IAM Instance Profile for k3s worker nodes, must set with --cloud-controller-manager",
		},
		{
			Name:  "request-spot-instance",
			P:     &p.RequestSpotInstance,
			V:     p.RequestSpotInstance,
			Usage: "request for spot instance",
		},
		{
			Name:  "spot-price",
			P:     &p.SpotPrice,
			V:     p.SpotPrice,
			Usage: "spot instance bid price (in dollar)",
		},
		{
			Name:  "master",
			P:     &p.Master,
			V:     p.Master,
			Usage: "Number of master node",
		},
		{
			Name:  "worker",
			P:     &p.Worker,
			V:     p.Worker,
			Usage: "Number of worker node",
		},
	}

	return fs
}
