# Match Engine

## Description
A match engine built on top of Raft, designed for exchange transactions.

## Build
To build the project, run the following command:
```sh
go build -o ./out/mind ./app/main.go
```

## Initialization

### Single Mode
To initialize a single node:
```sh
./out/mind server init --home ./private/node1
```
### Cluster Mode
To initialize a cluster with two nodes:
```sh
./out/mind server init --home ./private/node1
./out/mind server init --home ./private/node2
```

## Configuration

### Single Mode
**./private/node1/app_config.json**
```json
{
  "api_port": 3000,
  "node_id": 1,
  "url": "http://localhost:7701",
  "peers": {
    "1": "http://localhost:7701"
  },
  "join": false,
  "data_dir": ""
}
```
---
### Cluster Mode
**./private/node1/app_config.json**
```json
{
  "api_port": 3000,
  "node_id": 1,
  "url": "http://localhost:7701",
  "peers": {
    "1": "http://localhost:7701",
    "2": "http://localhost:7702"
  },
  "join": false,
  "data_dir": ""
}
```
**./private/node2/app_config.json**
```json
{
  "api_port": 3001,
  "node_id": 2,
  "url": "http://localhost:7702",
  "peers": {
    "1": "http://localhost:7701",
    "2": "http://localhost:7702"
  },
  "join": false,
  "data_dir": ""
}
```

## Run

### Single Mode
To run a single node:

```sh
./out/mind server run --home ./private/node1
```

### Cluster Mode
To run a cluster with two nodes:
```sh
./out/mind server run --home ./private/node1
./out/mind server run --home ./private/node2
```