package apollo

import (
	"sync"

	"errors"

	"github.com/ServiceComb/go-archaius/core"
	"github.com/zouyx/agollo"
)

const (
	ApolloName     = "apollo"
	ApolloSource   = "apolloSource"
	ApolloPriority = 0
)

//ApolloConfigSource implements ConfigSource interface
type ApolloConfigSource struct {
	sync.RWMutex
	Configurations map[string]interface{}
	initSuccess    bool
	eventChannel   <-chan *agollo.ChangeEvent
	callBack       core.DynamicConfigCallback
	ApolloConfig   *agollo.AppConfig
}

func NewApolloConfigSource(apolloConfig *agollo.AppConfig) core.ConfigSource {

	configs := make(map[string]interface{})
	aps := new(ApolloConfigSource)
	aps.Configurations = configs
	aps.ApolloConfig = apolloConfig
	aps.init()
	return aps
}

func (ap *ApolloConfigSource) GetSourceName() string {
	return ApolloName
}

func (ap *ApolloConfigSource) GetPriority() int {
	return ApolloPriority
}

func (ap *ApolloConfigSource) GetConfigurationByKey(key string) (interface{}, error) {
	ap.init()
	ap.RLock()
	value, ok := ap.Configurations[key]
	ap.RUnlock()
	if ok {
		return value, nil
	}
	//TODO Whether we should return different type or not?
	return nil, errors.New("key not exist")
}

func (ap *ApolloConfigSource) GetConfigurations() (map[string]interface{}, error) {
	ap.init()
	configs := make(map[string]interface{})
	it := agollo.GetApolloConfigCache().NewIterator()
	for {
		entry := it.Next()
		if entry != nil {
			configs[string(entry.Key)] = entry.Value
			continue

		}
		break
	}
	return configs, nil
}

func (ap *ApolloConfigSource) DynamicConfigHandler(callback core.DynamicConfigCallback) error {
	ap.init()
	ap.callBack = callback
	go ap.eventHandler()
	return nil
}

func (ap *ApolloConfigSource) Cleanup() error {
	return nil
}

func (ap *ApolloConfigSource) GetConfigurationsByDI(dimensionInfo string) (map[string]interface{}, error) {
	return nil, nil
}

func (ap *ApolloConfigSource) AddDimensionInfo(dimensionInfo string) (map[string]string, error) {
	return nil, nil
}

func (ap *ApolloConfigSource) GetConfigurationByKeyAndDimensionInfo(key, dimensionInfo string) (interface{}, error) {
	return nil, nil
}

func (ap *ApolloConfigSource) init() {
	if !ap.initSuccess {
		agollo.InitCustomConfig(func() (*agollo.AppConfig, error) {
			return ap.ApolloConfig, nil
		})
		agollo.Start()
		ap.eventChannel = agollo.ListenChangeEvent()
		it := agollo.GetApolloConfigCache().NewIterator()
		for it != nil {
			entry := it.Next()
			if entry != nil {
				ap.Lock()
				ap.Configurations[string(entry.Key)] = entry.Value
				ap.Unlock()
				continue
			}
			break
		}

		ap.initSuccess = true
	}
}

func (ap *ApolloConfigSource) eventHandler() {
	for {
		changeEvent := <-ap.eventChannel

		for key, configChange := range changeEvent.Changes {
			event := &core.Event{
				EventSource: ApolloSource,
				Key:         key,
				Value:       configChange.NewValue,
			}
			switch configChange.ChangeType {
			case agollo.ADDED:
				event.EventType = core.Create
			case agollo.MODIFIED:
				event.EventType = core.Update
			case agollo.DELETED:
				event.EventType = core.Delete
			}
			ap.Lock()
			ap.Configurations[key] = configChange.NewValue
			ap.Unlock()

			ap.callBack.OnEvent(event)
		}
	}
}
