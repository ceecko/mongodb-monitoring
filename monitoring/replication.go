package monitoring

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type replicationLag struct {
	host string
	lag  time.Duration
}

type ReplStatus struct {
	Members []*ReplMembers
}

type ReplMembers struct {
	Name       string
	StateStr   string
	OptimeDate time.Time
}

func getReplicationLag(ctx context.Context, cl *mongo.Client) ([]*replicationLag, error) {
	db := cl.Database("admin")
	r := db.RunCommand(ctx, bson.D{{Key: "replSetGetStatus", Value: "1"}})

	var status ReplStatus
	if err := r.Decode(&status); err != nil {
		return nil, err
	}

	var optimePrimary time.Time
	for _, m := range status.Members {
		if m.StateStr == "PRIMARY" {
			optimePrimary = m.OptimeDate
			break
		}
	}
	if optimePrimary.IsZero() {
		return nil, errors.New("no primary detected")
	}

	var lag []*replicationLag
	for _, m := range status.Members {
		if m.StateStr != "PRIMARY" {
			lag = append(lag, &replicationLag{
				host: m.Name,
				lag:  optimePrimary.Sub(m.OptimeDate),
			})
		}
	}

	return lag, nil
}
