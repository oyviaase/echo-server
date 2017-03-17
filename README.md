# Echo Server Enhanced!

A very simple HTTP echo server written in Go, meant for use on k8s
but useful anywhere. Also has support for websockets ;)

The goal of this project is to reliably provide useful information about
the current session, the instance serving the session, and help test
things like Load Balancers, Web Servers, and Browsers.

- Default GET shows plethora of useful information.
- Any messages sent from a websocket client are echoed.
- Visit `/ws` via browser for an interactive UI to connect and send websocket messages.
- The `PORT` and `SSLPORT` environment variable set the corresponding server ports.
- Has SSL/TLS support already including some play-around self-signed certs.

## Test it out!

```
docker run -d -p 8080:8080 inanimate/echo-server
```

Browsing to `localhost:8080`, you should then see something resembling the example below.

### SSL/TLS

The default encrypted port is `8443` and serves a self-signed certificate
I've generated and builtin. The the CN is `cakewalk.herpderp.com` so you
can easily distinguish it when working with echo-server.

> Obviously, this is for testing purposes only! Do not use this in
> production and definitely disallow it from being seen publicly!

If you'd like to use your own, you can clone this repo, gen your own cert, and rebuild the bin/image.

## Extras

Additionally, you can provide a `ADD_HEADERS` variable with JSON formatted
values to include as response headers. By default, `X-Real-Server: echo-server` is
set to help you verify you're getting a response from the echo-server.

```
ADD_HEADERS={"X-Foo": "bar", "X-Server": "cats1.0"}
```

We also accept the following k8s variables which are explicitly displayed
at the top of the page for quick observance.

* `POD_NAME` - Can be provided via `metadata.name`
* `POD_NAMESPACE` - Can be provided via `metadata.namespace`
* `POD_IP` - Can be provided via `status.podIP`

See [here](http://stackoverflow.com/a/34418819) for more on how to define
these in your manifest.

> You can specify any variables you want if you'd like to convey more
> info. In the example below, I use `HELM*` variables to provide info
> about the deployed chart.

## Example Output

```
-> My hostname is: echo-server-4282639374-6bvzg

-> My Pod Name is: echo-server-4282639374-6bvzg
-> My Pod Namespace is: playground
-> My Pod IP is: 10.2.1.30

-> Requesting IP: 10.2.2.0:40974

-> Request Headers:

  HTTP/1.1 GET /

  Host: example.com
  Accept-Encoding: gzip, deflate, sdch
  Accept-Language: en-US,en;q=0.8
  Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8
  Cache-Control: no-cache
  Connection: close
  Cookie: _ga=GA1.2.2092706772.1468371657
  Pragma: no-cache
  Upgrade-Insecure-Requests: 1
  User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36
  X-Forwarded-For: 192.168.1.149
  X-Forwarded-Host: example.com
  X-Forwarded-Port: 80
  X-Forwarded-Proto: http
  X-Real-Ip: 192.168.1.149


-> Response Headers:

  Content-Type: text/plain
  X-Real-Server: echo-server

 >> Note that you may also see Transfer-Encoding and Date!



-> My environment:
  ADD_HEADERS={"X-Real-Server": "echo-server"}
  APACHESUCKS_PORT=tcp://10.3.0.41:80
  APACHESUCKS_PORT_80_TCP=tcp://10.3.0.41:80
  APACHESUCKS_PORT_80_TCP_ADDR=10.3.0.41
  APACHESUCKS_PORT_80_TCP_PORT=80
  APACHESUCKS_PORT_80_TCP_PROTO=tcp
  APACHESUCKS_SERVICE_HOST=10.3.0.41
  APACHESUCKS_SERVICE_PORT=80
  APACHESUCKS_SERVICE_PORT_HTTP=80
  ECHO_SERVER_PORT=tcp://10.3.0.155:80
  ECHO_SERVER_PORT_80_TCP=tcp://10.3.0.155:80
  ECHO_SERVER_PORT_80_TCP_ADDR=10.3.0.155
  ECHO_SERVER_PORT_80_TCP_PORT=80
  ECHO_SERVER_PORT_80_TCP_PROTO=tcp
  ECHO_SERVER_SERVICE_HOST=10.3.0.155
  ECHO_SERVER_SERVICE_PORT=80
  ECHO_SERVER_SERVICE_PORT_HTTP=80
  HELM_CHART_NAME=echo-server
  HELM_IMAGE=inanimate/echo-server
  HELM_K8S_VERSION=1.5
  HELM_RELEASE_NAME=kindly-horse
  HELM_REPLICAS=3
  HELM_REVISION=1
  HELM_VERSION=v2.2.1
  HOME=/
  HOSTNAME=echo-server-4282639374-6bvzg
  KUBERNETES_PORT=tcp://10.3.0.1:443
  KUBERNETES_PORT_443_TCP=tcp://10.3.0.1:443
  KUBERNETES_PORT_443_TCP_ADDR=10.3.0.1
  KUBERNETES_PORT_443_TCP_PORT=443
  KUBERNETES_PORT_443_TCP_PROTO=tcp
  KUBERNETES_SERVICE_HOST=10.3.0.1
  KUBERNETES_SERVICE_PORT=443
  KUBERNETES_SERVICE_PORT_HTTPS=443
  NGINX_PORT=tcp://10.3.0.160:80
  NGINX_PORT_80_TCP=tcp://10.3.0.160:80
  NGINX_PORT_80_TCP_ADDR=10.3.0.160
  NGINX_PORT_80_TCP_PORT=80
  NGINX_PORT_80_TCP_PROTO=tcp
  NGINX_SERVICE_HOST=10.3.0.160
  NGINX_SERVICE_PORT=80
  NGINX_SERVICE_PORT_HTTP=80
  PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
  POD_IP=10.2.1.30
  POD_NAME=echo-server-4282639374-6bvzg
  POD_NAMESPACE=playground
  PORT=8080


-> Contents of /etc/resolv.conf:
search playground.svc.cluster.local svc.cluster.local cluster.local
nameserver 10.3.0.10
options ndots:5



-> Contents of /etc/hosts:
# Kubernetes-managed hosts file.
127.0.0.1	localhost
::1	localhost ip6-localhost ip6-loopback
fe00::0	ip6-localnet
fe00::0	ip6-mcastprefix
fe00::1	ip6-allnodes
fe00::2	ip6-allrouters
10.2.1.30	echo-server-4282639374-6bvzg



```

