package cluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/cnrancher/autok3s/pkg/cluster"
	"github.com/cnrancher/autok3s/pkg/common"
	"github.com/cnrancher/autok3s/pkg/providers"
	autok3stypes "github.com/cnrancher/autok3s/pkg/types"
	"github.com/cnrancher/autok3s/pkg/types/apis"
	"github.com/cnrancher/autok3s/pkg/utils"

	"github.com/fsnotify/fsnotify"
	"github.com/ghodss/yaml"
	"github.com/rancher/apiserver/pkg/apierror"
	"github.com/rancher/apiserver/pkg/store/empty"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/wrangler/pkg/data/convert"
	"github.com/rancher/wrangler/pkg/schemas/validation"
	"github.com/sirupsen/logrus"
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
	base := cluster.NewBaseProvider()
	err = base.GenerateProvider(providerName)
	if err != nil {
		return types.APIObject{}, err
	}
	err = base.SetConfig(b)
	if err != nil {
		return types.APIObject{}, err
	}
	//p, err := providers.GetProvider(providerName)
	//if err != nil {
	//	return types.APIObject{}, apierror.NewAPIError(validation.NotFound, err.Error())
	//}
	//err = p.SetConfig(b)
	//p.GenerateClusterName()

	config := apis.Cluster{}
	err = convert.ToObj(data.Data(), &config)
	if err != nil {
		return types.APIObject{}, err
	}
	// save credential config
	//if providerName != "native" {
	//	if err := viper.ReadInConfig(); err != nil {
	//		return types.APIObject{}, err
	//	}
	//	credFlags := b.GetCredentialFlags()
	//	options := data.Data().Map("options")
	//	for _, credential := range credFlags {
	//		if v, ok := options[credential.Name]; ok {
	//			viper.Set(fmt.Sprintf(common.BindPrefix, providerName, credential.Name), v)
	//		}
	//	}
	//	if err := viper.WriteConfig(); err != nil {
	//		return types.APIObject{}, err
	//	}
	//}
	// get default ssh config
	sshConfig := base.GetSSHConfig()
	utils.MergeConfig(reflect.ValueOf(sshConfig).Elem(), reflect.ValueOf(&config.SSH).Elem())
	base.SSH = sshConfig

	if err := base.CreateCheck(sshConfig); err != nil {
		return types.APIObject{}, apierror.NewAPIError(validation.InvalidOption, err.Error())
	}
	go func() {
		err = base.CreateK3sCluster()
		if err != nil {
			logrus.Errorf("create cluster error: %v", err)
			base.Rollback()
		}
	}()

	return types.APIObject{}, err
}

func (c *Store) List(apiOp *types.APIRequest, schema *types.APISchema) (types.APIObjectList, error) {
	list := types.APIObjectList{}
	clusterList, err := ListCluster()
	if err != nil {
		return list, err
	}
	for _, config := range clusterList {
		id := config.Name
		config.Name = strings.Split(config.Name, ".")[0]
		obj := types.APIObject{
			Type:   schema.ID,
			ID:     id,
			Object: config,
		}
		list.Objects = append(list.Objects, obj)
	}
	return list, nil
}

func (c *Store) ByID(apiOp *types.APIRequest, schema *types.APISchema, id string) (types.APIObject, error) {
	clusterInfo, err := cluster.GetClusterByID(id)
	if err != nil {
		// find from failed cluster
		clusterInfo, err = readClusterState(filepath.Join(common.GetClusterStatePath(), fmt.Sprintf("%s_%s", id, common.StatusFailed)))
		if err != nil {
			return types.APIObject{}, apierror.NewAPIError(validation.NotFound, fmt.Sprintf("cluster %s is not found, got error: %v", id, err))
		}
	}
	clusterName := strings.Split(id, ".")[0]
	obj := apis.Cluster{
		Metadata: clusterInfo.Metadata,
		Options:  clusterInfo.Options,
		SSH:      clusterInfo.SSH,
	}
	obj.Name = clusterName
	return types.APIObject{
		Type:   schema.ID,
		ID:     id,
		Object: obj,
	}, nil
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
	//err = provider.MergeClusterOptions()
	//if err != nil {
	//	return types.APIObject{}, err
	//}
	//provider.GenerateClusterName()
	err = provider.DeleteK3sCluster(true)
	return types.APIObject{}, err
}

func (c *Store) Watch(apiOp *types.APIRequest, schema *types.APISchema, w types.WatchRequest) (chan types.APIEvent, error) {
	var (
		result = make(chan types.APIEvent)
	)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return result, err
	}
	err = watcher.Add(common.GetClusterStatePath())
	if err != nil {
		return result, err
	}

	go func() {
		<-apiOp.Context().Done()
		watcher.Close()
		close(result)
	}()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				switch event.Op {
				case fsnotify.Create:
					result <- toClusterEvent(event.Op, event.Name, schema.ID)
				case fsnotify.Remove:
					if strings.HasSuffix(event.Name, fmt.Sprintf("_%s", common.StatusRunning)) ||
						strings.HasSuffix(event.Name, fmt.Sprintf("_%s", common.StatusFailed)) {
						_, fileName := filepath.Split(event.Name)
						context := strings.Split(fileName, "_")
						arrayInfo := strings.Split(context[0], ".")
						result <- types.APIEvent{
							Name:         "resource.remove",
							ResourceType: schema.ID,
							Object: types.APIObject{
								ID:   context[0],
								Type: schema.ID,
								Object: autok3stypes.Cluster{
									Metadata: autok3stypes.Metadata{
										Name:     arrayInfo[0],
										Provider: arrayInfo[len(arrayInfo)-1],
									},
								},
							},
						}
					} else if strings.HasSuffix(event.Name, fmt.Sprintf("_%s", common.StatusJoin)) {
						result <- toClusterEvent(event.Op, event.Name, schema.ID)
					}
				}
			case e, ok := <-watcher.Errors:
				if !ok {
					return
				}
				result <- types.APIEvent{
					Error: e,
				}
			}
		}
	}()

	return result, nil
}

func toClusterEvent(op fsnotify.Op, name, id string) types.APIEvent {
	if strings.HasSuffix(name, fmt.Sprintf("_%s", common.StatusCreating)) ||
		strings.HasSuffix(name, fmt.Sprintf("_%s", common.StatusFailed)) {
		r, e := ioutil.ReadFile(name)
		if e != nil {
			return types.APIEvent{
				Error: e,
			}
		}
		processCluster := &autok3stypes.Cluster{}
		e = yaml.Unmarshal(r, processCluster)
		if e != nil {
			return types.APIEvent{
				Error: e,
			}
		}
		clusterID := processCluster.Name
		processCluster.Name = strings.Split(processCluster.Name, ".")[0]
		event := types.APIEvent{
			ResourceType: id,
			Object: types.APIObject{
				ID:     clusterID,
				Type:   id,
				Object: processCluster,
			},
		}
		if strings.HasSuffix(name, fmt.Sprintf("_%s", common.StatusCreating)) {
			event.Name = "resource.create"
			return event
		}
		event.Name = "resource.change"
		return event
	}
	// event name is formed by "cfg-path/clusters/cluster-id_event",
	//e.g. .autok3s/clusters/myk3s.region.provider_Join
	context := strings.Split(name, "_")
	contextArray := strings.Split(context[0], "/")
	clusterInfo, err := cluster.GetClusterByID(contextArray[len(contextArray)-1])
	if err != nil {
		return types.APIEvent{
			Error: err,
		}
	}
	clusterID := clusterInfo.Name
	clusterInfo.Name = strings.Split(clusterInfo.Name, ".")[0]
	if op != fsnotify.Remove && strings.HasSuffix(name, fmt.Sprintf("_%s", common.StatusJoin)) {
		clusterInfo.Status.Status = "upgrading"
	}
	return types.APIEvent{
		Name:         "resource.change",
		ResourceType: id,
		Object: types.APIObject{
			ID:     clusterID,
			Type:   id,
			Object: clusterInfo,
		},
	}
}

func ListCluster() ([]*autok3stypes.ClusterInfo, error) {
	base := cluster.NewBaseProvider()
	list, err := base.ListCluster()
	return list, err
}

func readClusterState(statePath string) (*autok3stypes.Cluster, error) {
	b, err := ioutil.ReadFile(statePath)
	if err != nil {
		return nil, err
	}

	state := &autok3stypes.Cluster{}
	err = yaml.Unmarshal(b, state)
	return state, err
}
