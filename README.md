# MongoDB Monitoring

A very simple monitoring for MongoDB clusters.

- Reads `config.yaml`
- Discovers topology via `rs.config()`
- Pings every member
- Reports back to SignalFx as `mongodb.up` with `host` dimension

## Running

`mongodb-monitoring --help`
