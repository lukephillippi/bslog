# bslog

🍃 _Log archival as MongoDB BSON documents for [`log/slog`](https://pkg.go.dev/log/slog)_

[![Test](https://github.com/lukephillippi/bslog/actions/workflows/test.yaml/badge.svg)](https://github.com/lukephillippi/bslog/actions/workflows/test.yaml)
[![Go Reference](https://pkg.go.dev/badge/go.luke.ph/bslog.svg)](https://pkg.go.dev/go.luke.ph/bslog)

## Overview

The `bslog` package archives ordinary logs as queryable MongoDB BSON documents with minimal configuration.

Built on top of [the standard library's `log/slog` package](https://pkg.go.dev/log/slog), it offers a straightforward way to:

- ✅ Archive logs in MongoDB as native BSON documents
- ✅ Preserve the full structure and types of structured log attributes
- ✅ Wrap any existing `log/slog` `Handler` implementation
- ✅ Maintain full compatibility with existing `log/slog` integrations

It's lightweight, easy to configure, and integrates seamlessly with any existing `log/slog`-based logging setup.

## Installing

1. First, use `go get` to install the latest version of the package:

   ```shell
   go get -u go.luke.ph/bslog@latest
   ```

1. Next, include the package in your application:

   ```go
   import "go.luke.ph/bslog"
   ```

## Usage

The `bslog` package is designed for simplicity. Just wrap your existing handler and provide MongoDB collection details:

```go
client, err := mongo.Connect(context.Background())
if err != nil {
    // Handle the error...
}
defer client.Disconnect(context.Background())

handler := bslog.NewHandler(
   slog.NewTextHandler(os.Stdout, nil),
   client.Database("app").Collection("logs"),
)

slog.SetDefault(slog.New(handler))
```

## License

The package is released under [the Unlicense license](./LICENSE.md).

## References

- [pkg.go.dev/log/slog](https://pkg.go.dev/log/slog)
- [Structured Logging with slog](https://go.dev/blog/slog)
- [A Guide to Writing `slog` Handlers](https://github.com/golang/example/blob/master/slog-handler-guide/README.md)
