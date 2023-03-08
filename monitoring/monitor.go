package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ceecko/mongodb-monitoring/config"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Monitoring struct {
	authToken string
	clusters  []config.Cluster
	interval  time.Duration
}

func NewMonitoring(cfg *config.Config) *Monitoring {
	interval := 60
	if cfg.Interval != 0 {
		interval = cfg.Interval
	}

	return &Monitoring{
		authToken: cfg.SignalFx.AuthToken,
		clusters:  cfg.Clusters,
		interval:  time.Duration(interval) * time.Second,
	}
}

func (m *Monitoring) Start(ctx context.Context) error {
	logrus.Infof("Starting MongoDB monitoring every %s", m.interval.String())
	logrus.Infof("Detected %d clusters from config", len(m.clusters))

	go m.run(ctx, m.clusters)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(m.interval):
			go m.run(ctx, m.clusters)
		}
	}
}

func (m *Monitoring) run(ctx context.Context, clusters []config.Cluster) {
	var wg sync.WaitGroup
	logrus.Debug("Started metrics gathering")

	for _, c := range clusters {
		c := c
		wg.Add(len(c.Uris))

		for _, uri := range c.Uris {
			uri := uri
			l := logrus.WithField("cluster", uri)
			go func() {
				defer wg.Done()

				hosts, err := discoverTopology(context.Background(), uri, c.Username, c.Password)
				if err != nil {
					l.Error(err)
					return
				}

				var wgM sync.WaitGroup
				wgM.Add(len(hosts))
				for _, host := range hosts {
					host := host
					l := l.WithField("host", host)
					go func() {
						defer wgM.Done()
						if err := m.gatherMetrics(context.Background(), host, c.Username, c.Password); err != nil {
							l.Error(err)
							return
						}
					}()
				}

				wgM.Wait()
			}()
		}
	}

	wg.Wait()
}

func (m *Monitoring) gatherMetrics(ctx context.Context, host, user, password string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	credential := options.Credential{
		AuthSource: "admin",
		Username:   user,
		Password:   password,
	}

	c, err := mongo.Connect(ctx, options.Client().
		SetDirect(true).
		ApplyURI(fmt.Sprintf("mongodb://%s", host)).
		SetAuth(credential),
	)
	if err != nil {
		return err
	}

	defer func() {
		if err = c.Disconnect(ctx); err != nil {
			logrus.Error(err)
		}
	}()

	ctxP, cancelP := context.WithTimeout(ctx, 10*time.Second)
	defer cancelP()
	err = c.Ping(ctxP, nil)
	if err != nil {
		return err
	}

	logrus.Debugf("Ping success %s", host)

	return m.sendMetrics(ctx, []*datapoint.Datapoint{
		sfxclient.GaugeF("mongodb.up", map[string]string{"host": host}, 1.0),
	})
}

func (m *Monitoring) sendMetrics(ctx context.Context, datapoints []*datapoint.Datapoint) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client := sfxclient.NewHTTPSink()
	client.AuthToken = m.authToken
	return client.AddDatapoints(ctx, datapoints)
}
