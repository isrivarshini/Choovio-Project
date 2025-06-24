# Mainflux CoAP Adapter

Mainflux CoAP adapter provides an [CoAP](http://coap.technology/) API for sending messages through the
platform.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable              | Description            | Default                 |
|-----------------------|------------------------|-------------------------|
| MF_COAP_ADAPTER_PORT  | adapter listening port | `5683`                  |
| MF_NATS_URL           | NATS instance URL      | `nats://localhost:4222` |
| MF_CLIENTS_URL        | clients service URL    | `http://localhost:8181` |

## Deployment

The service is distributed as Docker container. The following snippet provides
a compose file template that can be used to deploy the service container locally:

```yaml
version: "2"
services:
  adapter:
    image: mainflux/coap-adapter:[version]
    container_name: [instance name]
    ports:
      - [host machine port]:[configured port]
    environment:
      MF_COAP_ADAPTER_PORT: [Service HTTP port]
      MF_NATS_URL: [NATS instance URL]
      MF_CLIENTS_URL: [Manager service URL]
```

Running this service outside of container requires working instance of the NATS service.
To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
go get github.com/mainflux/mainflux

cd $GOPATH/src/github.com/mainflux/mainflux

# compile the http
make coap

# copy binary to bin
make install

# set the environment variables and run the service
MF_MANAGER_URL=[Manager service URL] MF_NATS_URL=[NATS instance URL] MF_COAP_ADAPTER_PORT=[Service HTTP port] $GOBIN/mainflux-coap
```

## Usage

Since CoAP protocol does not support `Authorization` header (option), in order to send CoAP messages,
client valid key must be present in `Uri-Query` option.
