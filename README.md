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
```
```sh
./out/mind server init --home ./private/node2
```

## Configuration matching server

### Single Mode
**./private/node1/app_config.json**
```json
{
  "api_port": 3000,
  "node_id": 1,
  "node_url": "http://127.0.0.1:8081",
  "peers": {
    "1": "http://127.0.0.1:8081",   
  },
  "join": false,
  "data_dir": "",
  "etcd_endpoints": []
}
```
---
### Cluster Mode
**./private/node1/app_config.json**
```json
{
  "api_port": 3000,
  "node_id": 1,
  "node_url": "http://127.0.0.1:8081",
  "peers": {
    "1": "http://127.0.0.1:8081",
    "2": "http://127.0.0.1:8082"    
  },
  "join": false,
  "data_dir": "",
  "etcd_endpoints": []
}
```
**./private/node2/app_config.json**
```json
{
  "api_port": 3001,
  "node_id": 2,
  "node_url": "http://127.0.0.1:8082",
  "peers": {
    "1": "http://127.0.0.1:8081",
    "2": "http://127.0.0.1:8082"    
  },
  "join": false,
  "data_dir": "",
  "etcd_endpoints": []
}
```

## Run matching server

### Single Mode
To run a single node:

```sh
./out/mind server run --home ./private/node1
```

### Cluster Mode
To run a cluster with two nodes:
```sh
./out/mind server run --home ./private/node1
```
```sh
./out/mind server run --home ./private/node2
```

### Using etcd for service discovery
```sh
docker-compose -f ./deploy/etcd-docker-compose.yml up -d
```

note: should choose proper "platform"
```sh
platform: linux/arm64
```

## Testing

### Append Message
```sh
curl -X POST http://127.0.0.1:3000/api/v1/helloworld/message -H "Content-Type: application/json" -d '{"message":"version3"}'
```

### Get Message
```sh
curl -X GET http://127.0.0.1:3000/api/v1/helloworld/messages
```

### Get Leader
```sh
curl -X GET http://127.0.0.1:3000/api/v1/explorer/leader
```
