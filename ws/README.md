# WebSocket adapter

WebSocket adapter provides an WebSocket API for sending and receiving messages through the platform.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable             | Description         | Default               |
|----------------------|---------------------|-----------------------|
| MF_CLIENTS_URL       | Clients service URL | localhost:8181        |
| MF_NATS_URL          | NATS instance URL   | nats://localhost:4222 |
| MF_WS_ADAPTER_PORT   | Service WS port     | 8180                  |

## Deployment

The service is distributed as Docker container. The following snippet provides
a compose file template that can be used to deploy the service container locally:

```yaml
version: "2"
services:
  ws:
    image: mainflux/ws:[version]
    container_name: [instance name]
    ports:
      - [host machine port]:[configured port]
    environment:
      MF_CLIENTS_URL: [Clients service URL]
      MF_NATS_URL: [NATS instance URL]
      MF_WS_ADAPTER_PORT: [Service WS port]
```

To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
go get github.com/mainflux/mainflux

cd $GOPATH/src/github.com/mainflux/mainflux

# compile the ws
make ws

# copy binary to bin
make install

# set the environment variables and run the service
MF_CLIENTS_URL=[Clients service URL] MF_NATS_URL=[NATS instance URL] MF_WS_ADAPTER_PORT=[Service WS port] $GOBIN/mainflux-ws
```

## Usage

For more information about service capabilities and its usage, please check out
the [API documentation](swagger.yaml).