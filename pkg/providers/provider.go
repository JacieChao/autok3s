package providers

import (
	"fmt"
	"sync"

	"github.com/cnrancher/autok3s/pkg/types"
	"github.com/cnrancher/autok3s/pkg/types/apis"

	"github.com/sirupsen/logrus"
)

// Factory is a function that returns a Provider.Interface.
type Factory func() (Provider, error)

var (
	providersMutex sync.Mutex
	providers      = make(map[string]Factory)
)

// Provider is an abstract, pluggable interface for k3s provider.
type Provider interface {
	// GetProviderName returns provider name
	GetProviderName() string
	// GetUsageExample return command usage example.
	GetUsageExample(action string) string
	// GetCreateFlags return create flags.
	GetCreateFlags() []types.Flag
	// GetOptionFlags return create flags of provider options.
	GetOptionFlags() []types.Flag
	// GetJoinFlags return join command flags.
	GetJoinFlags() []types.Flag
	// GetDeleteFlags return delete command flags.
	GetDeleteFlags() []types.Flag
	// GetSSHFlags return SSH command flags.
	GetSSHFlags() []types.Flag
	// GetCredentialFlags return credential flags.
	GetCredentialFlags() []types.Flag
	// GenerateClusterName Generate cluster id.
	GenerateClusterName() string
	// GenerateMasterExtraArgs return create/join extra master args for different provider.
	GenerateMasterExtraArgs(cluster *types.Cluster, master types.Node) string
	// GenerateWorkerExtraArgs return create/join extra worker args for different provider.
	GenerateWorkerExtraArgs(cluster *types.Cluster, worker types.Node) string
	// CreateK3sCluster create K3s cluster interface.
	CreateK3sCluster() error
	// JoinK3sNode join K3s node interface.
	JoinK3sNode() error
	// DeleteK3sCluster delete K3s cluster interface.
	DeleteK3sCluster(f bool) error
	// SSHK3sNode ssh to specified K3s node.
	SSHK3sNode(node string) error
	// IsClusterExist check cluster is exist.
	IsClusterExist() (bool, []string, error)
	// MergeClusterOptions merge exist cluster options
	MergeClusterOptions() error
	// DescribeCluster shows detailed cluster information.
	DescribeCluster(kubecfg string) *types.ClusterInfo
	// GetCluster return cluster simple information.
	GetCluster(kubecfg string) *types.ClusterInfo
	// GetSSHConfig return default ssh config for provider.
	GetSSHConfig() *types.SSH
	// SetConfig set cluster configuration of provider.
	SetConfig(config []byte) error
	// CreateCheck validate create flags.
	CreateCheck() error
	// SetMetadata merge metadata configs for provider.
	SetMetadata(config *types.Metadata)
	// SetOptions merge provider options.
	SetOptions(opt []byte) error
	// JoinCheck validate join flags.
	JoinCheck() error
	// GetClusterOptions return cluster config options.
	GetClusterOptions() []types.Flag
	// GetCreateOptions return create command options.
	GetCreateOptions() []types.Flag
	// GetProviderOptions convert options to specified provider option interface.
	GetProviderOptions(opt []byte) (interface{}, error)
	// BindCredential persistent credential from flags to db.
	BindCredential() error
	// RegisterCallbacks register callback functions which is used for execute logic after create/join
	RegisterCallbacks(name, event string, fn func(interface{}))
}

// RegisterProvider registers a provider.Factory by name.
func RegisterProvider(name string, p Factory) {
	providersMutex.Lock()
	defer providersMutex.Unlock()
	if _, found := providers[name]; !found {
		logrus.Debugf("registered provider %s", name)
		providers[name] = p
	}
}

// GetProvider creates an instance of the named provider, or nil if
// the name is unknown.  The error return is only used if the named provider
// was known but failed to initialize.
func GetProvider(name string) (Provider, error) {
	providersMutex.Lock()
	defer providersMutex.Unlock()
	f, found := providers[name]
	if !found {
		return nil, fmt.Errorf("provider %s is not registered", name)
	}
	return f()
}

// ListProviders list all providers that have registered to autok3s
func ListProviders() []apis.Provider {
	providersMutex.Lock()
	defer providersMutex.Unlock()
	list := make([]apis.Provider, 0)
	for p := range providers {
		list = append(list, apis.Provider{
			Name: p,
		})
	}
	return list
}
