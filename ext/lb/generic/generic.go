package generic

import (
	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	etypes "github.com/docker/engine-api/types/events"
	"github.com/ehazlett/interlock/config"
)

const (
	pluginName = "generic"
)

type GenericLoadBalancer struct {
	cfg    *config.ExtensionConfig
	client *client.Client
}

func log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"ext": pluginName,
	})
}

func NewGenericLoadBalancer(c *config.ExtensionConfig, cl *client.Client) (*GenericLoadBalancer, error) {
	lb := &GenericLoadBalancer{
		cfg:    c,
		client: cl,
	}

	return lb, nil
}

func (p *GenericLoadBalancer) Name() string {
	return pluginName
}

func (p *GenericLoadBalancer) HandleEvent(event *etypes.Message) error {
	return nil
}

func (p *GenericLoadBalancer) ConfigPath() string {
	return ""
}

func (p *GenericLoadBalancer) Template() string {
	return ""
}

func (p *GenericLoadBalancer) NeedsReload() bool {
	return false
}

func (p *GenericLoadBalancer) Reload(proxyContainers []types.Container) error {
	return nil
}
