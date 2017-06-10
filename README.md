# Overview

Simple container that expose a service with common endpoints for testing a service (echo, echoheaders, hostname, fqdn, ip).

## Usage

```bash
docker run -d -p 8080:8080 --rm fabriziopandini/hello
```

** /echo endpoint **

```bash
curl --data "hello echo" localhost:8080/echo
```

or

```bash
curl --data "hello echo" localhost:8080/
```

Returns echo of the request body.

** /echoheaders endpoint **

```bash
curl localhost:8080/echoheaders
```

Returns echo of the request headers.

** /hostname endpoint **

```bash
curl localhost:8080/hostname
```

Returns the container hostname.

** /fqdn endpoint **

```bash
curl localhost:8080/fqdn
```

Returns the container fully qualified name.

** /ip endpoint **

```bash
curl localhost:8080/ip
```

Returns the list of containers ip.

** /env endpoint **

```bash
curl localhost:8080/env
```

Returns the list of container env variables.

** /exit/exitCode endpoint **

```bash
curl localhost:8080/exit/0
curl localhost:8080/exit/1
```

Exit from the service and returns the given code

** /healthz endpoint **

```bash
curl localhost:8080/healthz
```

Returns 200 and a message with uptime info

** /healthz-fail/failAfter endpoint **

```bash
curl localhost:8080/healthz-fail/10
```

if uptime less than failAfter, returns 200 and a message with uptime info, otherwise returns 500 and a message with uptime info,