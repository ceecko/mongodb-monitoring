package monitoring

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Config ReplSetConfig
}

type ReplSetConfig struct {
	Members []Topology
}

type Topology struct {
	Host string
}

func discoverTopology(ctx context.Context, uri, user, password string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	credential := options.Credential{
		AuthSource: "admin",
		Username:   user,
		Password:   password,
	}

	c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetAuth(credential))
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = c.Disconnect(ctx); err != nil {
			logrus.Error(err)
		}
	}()

	logrus.Debugf("Getting topology for %s", uri)
	db := c.Database("admin")
	r := db.RunCommand(ctx, bson.D{{Key: "replSetGetConfig", Value: "1"}})
	var cfg Config

	if err = r.Decode(&cfg); err != nil {
		return nil, err
	}

	hosts := make([]string, 0, len(cfg.Config.Members))
	for _, h := range cfg.Config.Members {
		hosts = append(hosts, h.Host)
	}

	logrus.Debugf("Topology for %s is %+v", uri, hosts)

	return hosts, nil
}
