package beacon

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/ehazlett/interlock/config"
	"github.com/ehazlett/interlock/events"
	"github.com/ehazlett/interlock/utils"
	"github.com/prometheus/client_golang/prometheus/push"
	"golang.org/x/net/context"
)

const (
	pluginName = "beacon"
)

var (
	errChan chan (error)
)

type Beacon struct {
	cfg       *config.ExtensionConfig
	client    *client.Client
	monitored map[string]int
}

func log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"ext": pluginName,
	})
}

type eventArgs struct {
	Image string
}

func NewBeacon(c *config.ExtensionConfig, cl *client.Client) (*Beacon, error) {
	// parse config base dir
	c.ConfigBasePath = filepath.Dir(c.ConfigPath)

	errChan = make(chan error)
	go func() {
		for err := range errChan {
			log().Error(err)
		}
	}()

	ext := &Beacon{
		cfg:       c,
		client:    cl,
		monitored: map[string]int{},
	}

	containerID, err := utils.GetContainerID()
	if err != nil {
		return nil, err
	}

	// ticker to push to gateway if configured
	d, err := time.ParseDuration(c.StatsInterval)
	if err != nil {
		return nil, fmt.Errorf("unable to parse stat interval: %s", err)
	}
	t := time.NewTicker(d)
	go func() {
		for range t.C {
			log().Debug("stats ticker")
			ext.collectStats()

			gw := c.StatsPrometheusPushGatewayURL
			if gw != "" {
				job := "beacon"
				groups := map[string]string{
					"source":    "beacon",
					"container": containerID,
				}
				log().Debug("pushing to gateway")
				if err := push.Collectors(
					job,
					groups,
					gw,
					counterTotalContainers,
					counterTotalImages,
					counterTotalVolumes,
					counterTotalNetworks,
					counterCpuTotalUsage,
					counterMemoryUsage,
					counterMemoryMaxUsage,
					counterMemoryPercent,
					counterNetworkRxBytes,
					counterNetworkRxPackets,
					counterNetworkRxErrors,
					counterNetworkRxDropped,
					counterNetworkTxBytes,
					counterNetworkTxPackets,
					counterNetworkTxErrors,
					counterNetworkTxDropped,
				); err != nil {
					log().Errorf("error pushing to gateway: %s", err)
				}
			}
		}
	}()

	return ext, nil
}

func (b *Beacon) Name() string {
	return pluginName
}

func (b *Beacon) HandleEvent(event *events.Message) error {
	switch event.Status {
	case "interlock-start":
		// scan all containers and start metrics
		opts := types.ContainerListOptions{
			All:  false,
			Size: false,
		}
		containers, err := b.client.ContainerList(context.Background(), opts)
		if err != nil {
			return err
		}

		for _, c := range containers {
			b.monitored[c.ID] = 1
		}

		log().Debugf("containers: %v", b.monitored)
	case "start":
		log().Debugf("checking container for stats: id=%s", event.ID)
		b.monitored[event.ID] = 1
	case "kill", "die", "stop", "destroy":
		log().Debugf("resetting stats: id=%s", event.ID)
		delete(b.monitored, event.ID)

		if err := b.resetStats(event.ID); err != nil {
			return err
		}
	}

	return nil
}
