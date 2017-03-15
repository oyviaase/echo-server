# Echo Server Enhanced!

This version shows more env information and is already available
at `inanimate/echo-server`!!

Additionally, you can provide a `ADD_HEADERS` variable with JSON formatted
values to include as headers. By default, `X-Real-Server: echo-server` is
set to help you verify you're getting a response from the echo-server.

```
ADD_HEADERS={"X-Foo": "bar", "X-Server": "cats1.0"}
```

We also accept the following variables:

* `POD_NAME` - Can be provided via `metadata.name`
* `POD_NAMESPACE` - Can be provided via `metadata.namespace`
* `POD_IP` - Can be provided via `status.podIP`

See [here](http://stackoverflow.com/a/34418819) for more on how to define
these in your manifest.

--------------------------------------------------

A very simple HTTP echo server with support for web-sockets.

- Any messages sent from a web-socket client are echoed.
- Visit `/.ws` for a basic UI to connect and send web-socket messages.
- Requests to any other URL will return the request headers and body.
- The `PORT` environment variable sets the server port.
- No TLS support yet :(

To run as a container:

```
docker run --detach -P jmalloc/echo-server
```

To run as a service:

```
docker service create --publish 8080 jmalloc/echo-server
```
