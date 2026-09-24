# natsacl

`natsacl` is a small Go library for building [NATS](https://nats.io/) permissions from high-level capabilities.

Instead of manually maintaining low-level NATS subjects and JetStream API permissions, `natsacl` lets you describe what a user should be allowed to do:

```go
perms := natsacl.NewBuilder()

perms.Stream("EVENTS").
    Publish().
    Consume()

perms.KV("CONFIG").
    Read().
    Write("application.timeout")

perms.Obj("ASSETS").
    Read().
    Write()

perms.Service(
    "inventory",
    "inventory.get",
    "inventory.reserve",
).
    Request().
    Monitor()
```

The resulting publish and subscribe subject lists can then be used to generate NATS user permissions.

## Why?

JetStream features such as Streams, Key/Value stores and Object Stores rely on several internal API subjects.

Granting the right permissions therefore often means knowing implementation details such as:

```text
$JS.API.STREAM.INFO.EVENTS
$JS.API.CONSUMER.CREATE.EVENTS.*
$KV.CONFIG.>
$O.ASSETS.M.*
$SRV.INFO.inventory
```

`natsacl` translates higher-level capabilities into those subjects while keeping the resulting NATS ACLs as restrictive as possible.

## Supported capabilities

`natsacl` currently provides builders for:

* JetStream Streams
* Key/Value stores
* Object Stores
* NATS request/reply services
* NATS Micro service monitoring
* Raw publish/subscribe permissions when needed

Capabilities can be chained and combined:

```go
perms := natsacl.NewBuilder()

perms.
    Stream("ORDERS").
        Publish().
        Consume().
        Build().
    KV("CONFIG").
        Read().
        Build().
    Service("orders", "orders.create", "orders.cancel").
        Request().
        Monitor()

publishPermissions := perms.Pub()
subscribePermissions := perms.Sub()
```

## Testing

The library is tested against an embedded real `nats-server` with JetStream enabled.

Tests verify both that requested operations are allowed and that unrelated operations remain denied.

## Status

This project is currently under development.

The API and generated permissions may change while the supported NATS capabilities are being expanded and validated.

## License

0BSD. Do whatever you want with it.
