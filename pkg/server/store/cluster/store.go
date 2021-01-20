package cluster

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	com "github.com/cnrancher/autok3s/cmd/common"
	"github.com/cnrancher/autok3s/pkg/cluster"
	"github.com/cnrancher/autok3s/pkg/common"
	"github.com/cnrancher/autok3s/pkg/providers"
	autok3stypes "github.com/cnrancher/autok3s/pkg/types"
	"github.com/cnrancher/autok3s/pkg/types/apis"
	"github.com/cnrancher/autok3s/pkg/utils"
	"github.com/rancher/apiserver/pkg/apierror"
	"github.com/rancher/apiserver/pkg/store/empty"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/wrangler/pkg/data/convert"
	"github.com/rancher/wrangler/pkg/schemas/validation"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Store struct {
	empty.Store
}

func (c *Store) Create(apiOp *types.APIRequest, schema *types.APISchema, data types.APIObject) (types.APIObject, error) {
	providerName := data.Data().String("provider")
	b, err := json.Marshal(data.Data())
	if err != nil {
		return types.APIObject{}, err
	}
	p, err := providers.GetProvider(providerName)
	if err != nil {
		return types.APIObject{}, apierror.NewAPIError(validation.NotFound, err.Error())
	}
	logrus.Infof("-------- get body %s", string(b))
	err = p.SetConfig(b)
	p.GenerateClusterName()

	config := apis.Cluster{}
	err = convert.ToObj(data.Data(), &config)
	if err != nil {
		return types.APIObject{}, err
	}
	logrus.Infof("======= check config %++v", p)
	logrus.Infof("33333333 get ssh config %++v", config.SSH)
	// save credential config
	if providerName != "native" {
		if err := viper.ReadInConfig(); err != nil {
			return types.APIObject{}, err
		}
		credFlags := p.GetCredentialFlags()
		options := data.Data().Map("options")
		for _, credential := range credFlags {
			if v, ok := options[credential.Name]; ok {
				viper.Set(fmt.Sprintf(common.BindPrefix, providerName, credential.Name), v)
			}
		}
		if err := viper.WriteConfig(); err != nil {
			return types.APIObject{}, err
		}
	}

	go func() {
		err = p.CreateK3sCluster(&config.SSH)
		if err != nil {
			logrus.Errorf("create cluster error: %v", err)
			err = p.Rollback()
			logrus.Errorf("rollback cluster error: %v", err)
		}
	}()

	return types.APIObject{}, err
}

func (c *Store) List(apiOp *types.APIRequest, schema *types.APISchema) (types.APIObjectList, error) {
	list := types.APIObjectList{}
	v := common.CfgPath
	if v == "" {
		return list, apierror.NewAPIError(validation.InvalidOption, "state file is empty")
	}
	// get all clusters from state
	clusters, err := utils.ReadYaml(v, common.StateFile)
	if err != nil {
		return list, fmt.Errorf("read state file error, msg: %v\n", err)
	}

	result, err := cluster.ConvertToClusters(clusters)
	if err != nil {
		return list, fmt.Errorf("failed to unmarshal state file, msg: %v\n", err)
	}
	var (
		p           providers.Provider
		clusterList []*autok3stypes.Cluster
	)

	kubeCfg := fmt.Sprintf("%s/%s", common.CfgPath, common.KubeCfgFile)
	for _, r := range result {
		p, err = com.GetProviderByState(r)
		if err != nil {
			logrus.Errorf("failed to convert cluster options for cluster %s", r.Name)
			continue
		}
		isExist, _, err := p.IsClusterExist()
		if err != nil {
			logrus.Errorf("failed to check cluster %s exist, got error: %v ", r.Name, err)
			continue
		}
		if !isExist {
			logrus.Warnf("cluster %s is not exist, will remove from config", r.Name)
			// remove kube config if cluster not exist
			if err := cluster.OverwriteCfg(r.Name); err != nil {
				logrus.Errorf("failed to remove unexist cluster %s from kube config", r.Name)
			}
			continue
		}
		config := p.GetCluster(kubeCfg)
		context := strings.Split(config.Name, ".")
		config.Name = context[0]
		obj := types.APIObject{
			Type:   schema.ID,
			ID:     r.Name,
			Object: config,
		}
		list.Objects = append(list.Objects, obj)
		clusterList = append(clusterList, &autok3stypes.Cluster{
			Metadata: r.Metadata,
			Options:  r.Options,
			Status:   r.Status,
		})
	}
	// remove useless clusters from .state.
	if err := cluster.FilterState(clusterList); err != nil {
		return list, fmt.Errorf("failed to remove useless clusters\n")
	}
	return list, nil
}

func (c *Store) ByID(apiOp *types.APIRequest, schema *types.APISchema, id string) (types.APIObject, error) {
	content := strings.Split(id, ".")
	providerName := content[len(content)-1]
	v := common.CfgPath
	if v == "" {
		return types.APIObject{}, fmt.Errorf("[cluster] cfg path is empty")
	}

	clusters, err := utils.ReadYaml(v, common.StateFile)
	if err != nil {
		return types.APIObject{}, err
	}

	converts, err := cluster.ConvertToClusters(clusters)
	if err != nil {
		return types.APIObject{}, fmt.Errorf("[cluster] failed to unmarshal state file, msg: %s", err)
	}
	for _, con := range converts {
		if con.Provider == providerName && con.Name == id {
			return types.APIObject{
				Type:   schema.ID,
				ID:     id,
				Object: con,
			}, nil
		}
	}
	return types.APIObject{}, apierror.NewAPIError(validation.NotFound, fmt.Sprintf("cluster %s is not found", id))
}

func (c *Store) Delete(apiOp *types.APIRequest, schema *types.APISchema, id string) (types.APIObject, error) {
	context := strings.Split(id, ".")
	providerName := context[len(context)-1]
	provider, err := providers.GetProvider(providerName)
	if err != nil {
		return types.APIObject{}, apierror.NewAPIError(validation.NotFound, err.Error())
	}
	config := autok3stypes.Cluster{
		Metadata: autok3stypes.Metadata{
			Name:     context[0],
			Provider: providerName,
		},
	}
	if len(context) == 3 {
		config.Options = map[string]interface{}{
			"region": context[1],
		}
	}
	b, err := json.Marshal(config)
	if err != nil {
		return types.APIObject{}, err
	}
	err = provider.SetConfig(b)
	if err != nil {
		return types.APIObject{}, err
	}
	err = provider.MergeClusterOptions()
	if err != nil {
		return types.APIObject{}, err
	}
	provider.GenerateClusterName()
	err = provider.DeleteK3sCluster(true)
	return types.APIObject{}, err
}

func (c *Store) Watch(apiOp *types.APIRequest, schema *types.APISchema, w types.WatchRequest) (chan types.APIEvent, error) {
	var (
		result    = make(chan types.APIEvent, 100)
		countLock sync.Mutex
	)

	go func() {
		<-apiOp.Context().Done()
		countLock.Lock()
		close(result)
		result = nil
		countLock.Unlock()
	}()

	return result, nil
}
