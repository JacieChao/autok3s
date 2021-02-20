package cluster

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/cnrancher/autok3s/pkg/common"
	"github.com/cnrancher/autok3s/pkg/providers"
	"github.com/cnrancher/autok3s/pkg/types"
	"github.com/cnrancher/autok3s/pkg/utils"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/syncmap"
)

const (
	k3sVersion             = ""
	k3sChannel             = "stable"
	k3sInstallScript       = "https://get.k3s.io"
	master                 = "0"
	worker                 = "0"
	ui                     = false
	cloudControllerManager = false
	embedEtcd              = false
	defaultCidr            = "10.42.0.0/16"
	dockerScript           = "curl -sSL https://get.docker.com | sh - %s"
)

type ProviderBase struct {
	*types.Metadata
	types.Status
	*types.SSH
	Options interface{}
	p       providers.Provider
	m       *sync.Map
	logger  *logrus.Logger
}

func NewBaseProvider() *ProviderBase {
	return &ProviderBase{
		Metadata: &types.Metadata{
			UI:                     ui,
			CloudControllerManager: cloudControllerManager,
			K3sVersion:             k3sVersion,
			K3sChannel:             k3sChannel,
			InstallScript:          k3sInstallScript,
			Cluster:                embedEtcd,
			DockerScript:           dockerScript,
		},
		Status: types.Status{
			MasterNodes: make([]types.Node, 0),
			WorkerNodes: make([]types.Node, 0),
		},
		m: new(syncmap.Map),
	}
}

func (b *ProviderBase) GenerateProvider(name string) error {
	b.Provider = name
	p, err := providers.GetProvider(name)
	if err != nil {
		return err
	}
	b.p = p
	return nil
}

func (b *ProviderBase) SetConfig(config []byte) error {
	c := types.Cluster{}
	err := json.Unmarshal(config, &c)
	if err != nil {
		return err
	}
	sourceMeta := reflect.ValueOf(b.Metadata).Elem()
	targetMeta := reflect.ValueOf(&c.Metadata).Elem()
	utils.MergeConfig(sourceMeta, targetMeta)
	sourceOption := reflect.ValueOf(&b.Options).Elem()
	targetOption := reflect.ValueOf(&c.Options).Elem()
	utils.MergeConfig(sourceOption, targetOption)

	return nil
}

func (b *ProviderBase) GetSSHConfig() *types.SSH {
	return b.p.GetSSHConfig()
}

func (b *ProviderBase) GetProviderName() string {
	return b.Provider
}

func (b *ProviderBase) GenerateClusterName() {
	b.p.GenerateClusterName()
}

func (b *ProviderBase) CreateCheck(ssh *types.SSH) error {
	b.p.SetMetadata(b.Metadata)
	b.p.GenerateClusterName()
	return b.p.CreateCheck(ssh)
}

func (b *ProviderBase) Rollback() error {
	return b.p.Rollback()
}

func (b *ProviderBase) CreateK3sCluster() (err error) {
	logrus.Infof("[%s] prepare for k3s cluster...\n", b.GetProviderName())
	b.Options = b.p.GetOptions()
	c := &types.Cluster{
		Metadata: *b.Metadata,
		Options:  b.Options,
		Status: types.Status{
			Status: common.StatusCreating,
		},
	}
	defer func() {
		if err != nil {
			// save failed status
			if c == nil {
				c = &types.Cluster{
					Metadata: *b.Metadata,
					Options:  b.Options,
					Status:   types.Status{},
				}
			}
			c.Status.Status = common.StatusFailed
			SaveCluster(c)
		}
		if err == nil && len(b.Status.MasterNodes) > 0 {
			fmt.Println(common.UsageInfoTitle)
			fmt.Printf(common.UsageContext, b.p.GenerateClusterName())
			fmt.Println(common.UsagePods)
			if b.UI {
				if b.CloudControllerManager {
					fmt.Printf("\nK3s UI URL: https://<using `kubectl get svc -A` get UI address>:8999\n")
				} else {
					fmt.Printf("\nK3s UI URL: https://%s:8999\n", b.Status.MasterNodes[0].PublicIPAddress[0])
				}
			}
			fmt.Println("")
		}
	}()
	// save cluster
	err = SaveCluster(c)
	if err != nil {
		return err
	}

	// generate instance by provider
	b.p.SetMetadata(b.Metadata)
	c, err = b.p.Prepare(b.SSH)
	if err != nil {
		return err
	}
	// deploy k3s cluster
	if err = InitK3sCluster(c); err != nil {
		return err
	}
	logrus.Infof("========= get cluster %++v", c)
	// deploy manifest for provider
	extraManifests := b.p.GenerateManifest()
	if extraManifests != nil {
		if err = DeployExtraManifest(c, extraManifests); err != nil {
			return err
		}
		logrus.Infof("[%s] successfully deployed manifests", b.Provider)
	}

	//b.logger.Infof("[%s] successfully executed create logic\n", b.GetProviderName())
	return nil
}

func (b *ProviderBase) DeleteK3sCluster(force bool) error {
	isConfirmed := true

	if !force {
		isConfirmed = utils.AskForConfirmation(fmt.Sprintf("[%s] are you sure to delete cluster %s", b.Provider, b.Name))
	}
	if isConfirmed {
		logrus.Infof("----- get meta %v", *b.Metadata)
		logrus.Infof("------ merged options %v", b.p.GetOptions())
		contextName := b.p.GenerateClusterName()
		err := b.p.DeleteK3sCluster(force)
		if err != nil {
			return err
		}
		err = OverwriteCfg(contextName)
		if err != nil && !force {
			return fmt.Errorf("[%s] synchronizing .cfg file error, msg: %v", b.Provider, err)
		}
		err = DeleteClusterState(b.Name, b.Provider)
		if err != nil && !force {
			return fmt.Errorf("[%s] synchronizing .state file error, msg: %v", b.Provider, err)
		}

		logrus.Infof("[%s] successfully deleted cluster %s\n", b.Provider, b.Name)
	}
	return nil
}

func (b *ProviderBase) ListCluster() ([]*types.ClusterInfo, error) {
	kubeCfg := fmt.Sprintf("%s/%s", common.CfgPath, common.KubeCfgFile)
	clusterList := []*types.ClusterInfo{}
	stateList, err := ListClusterState()
	if err != nil {
		return nil, err
	}
	for _, state := range stateList {
		b.GenerateProvider(state.Provider)
		b.p.SetMetadata(&state.Metadata)
		b.p.SetOptions(state.Options)
		contextName := b.p.GenerateClusterName()
		isExist, _, err := b.p.IsClusterExist()
		if err != nil {
			logrus.Errorf("failed to check provider %s cluster %s exist, got error: %v ", state.Provider, state.Name, err)
			continue
		}
		if !isExist {
			logrus.Warnf("cluster %s (provider %s) is not exist, will remove from config", state.Name, state.Provider)
			// remove kube config if cluster not exist
			if err := OverwriteCfg(contextName); err != nil {
				logrus.Errorf("failed to remove unexist cluster %s from kube config", state.Name)
			}
			if err := DeleteState(state.Name, state.Provider); err != nil {
				logrus.Errorf("failed to remove unexist cluster %s from state: %v", state.Name, err)
			}
			continue
		}
		clusterList = append(clusterList, b.p.GetCluster(kubeCfg))
	}

	return clusterList, nil
}
