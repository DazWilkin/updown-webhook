# Golang implementation of a Webhook for [`updown.io`](https://updown.io)

[![build](https://github.com/DazWilkin/updown-webhook/actions/workflows/build.yml/badge.svg)](https://github.com/DazWilkin/updown-webhook/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/DazWilkin/updown-webhook)](https://goreportcard.com/report/github.com/DazWilkin/updown-webhook)

## Events

|Name|
|----|
|`check.down`|
|`check.up`|
|`check.ssl_invalid`|
|`check.ssl_valid`|
|`check.ssl_expiration`|
|`check.ssl_renewed`|
|`check.performance_drop`|

## Test

```bash
# Local
WHOOK="8888"

curl \
--request POST \
--header "Content-Type: application/json" \
-d @"samples/check.down.json" \
http://localhost:${WHOOK}
```

### Tailscale Serve

```bash
WHOOK="8888"
PROXY="10000"

# Available only within tailnet
tailscale serve https:${PROXY} / http://localhost:${WHOOK}

HOST="..."
TAILNET="...ts.net"

curl \
--request POST \
--header "Content-Type: application/json" \
-d @"samples/check.down.json" \
https://${HOST}.${TAILNET}:${PROXY}
```

### Tailscale Funnel

```bash
WHOOK="8888"
PROXY="10000"

tailscale serve https:${PROXY} / http://localhost:${WHOOK}

# Available publicly (e.g. by Updown service)
tailscale funnel ${PROXY} on

HOST="..."
TAILNET="...ts.net"

curl \
--request POST \
--header "Content-Type: application/json" \
-d @"samples/check.down.json" \
https://${HOST}.${TAILNET}:${PROXY}

```

### Updown Webhook tester

Add the Webhook to settings (console|API)

https://updown.io/recipients/test

## Prometheus

Metrics are prefixed `updown_`

|Name|Type|Description|
|----|----|-----------|
|`build_info`|Counter|A metrc with constant value '1'|
|`handler_total`|Counter||
|`handler_failures`|Counter||

### [Sigstore](https://www.sigstore.dev/)

`updown-webhook` container images are being signed by Sigstore and may be verified:

```bash
cosign verify \
--key=./cosign.pub \
ghcr.io/dazwilkin/updown-webhook:1234567890123456789012345678901234567890
```

> **NOTE** `cosign.pub` may be downloaded [here](./cosign.pub)

To install `cosign`, e.g.:

```bash
go install github.com/sigstore/cosign/cmd/cosign@latest
```

<hr/>
<br/>
<a href="https://www.buymeacoffee.com/dazwilkin" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/default-orange.png" alt="Buy Me A Coffee" height="41" width="174"></a>

