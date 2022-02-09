# chrome-discord-bridge

This code bridges Chrome "native messages" to Discord's IPC socket.

It's written in Go with no third-party dependencies.   It's intended to minimize the amount of code running natively, so that the logic for communicating with Discord can be isolated in the Chrome extension's sandbox.  It's much lighter-weight than PreMiD's node.js app.  At the same time, it's easier to reason about the correctness of the code compared to the official C++ Discord-RPC library because of Go's memory management and concurrency primitives.

The bridge itself and the library functionality for both the Discord and Chrome ends is complete and unit-tested, and there is a utility for manually setting the Discord status (without a Chrome extension) in `cmd/set-activity`.

A Chrome extension to utilize the bridge is WIP as of February 2022, see https://github.com/p00ya/browser-activity.

## Building and running

To use with a Chrome extension with the ID `nglhipbdoknhpejdpceibmeaohidgcod`, first add the extension URL on its own line in `cmd/chrome-discord-bridge/origins.txt`.  This list of allowed origins is built into the binary as an additional layer of security.

Then with Go 1.17+, run:

```
go build ./cmd/chrome-discord-bridge ./cmd/install-host
```

This will build the bridge (`chrome-discord-bridge` binary) and a helper for installing a Native Messaging Host manifest to Chrome.

Run the `install-host` command to write a manifest for the Native Messaging Host to Chrome:

```
ID='nglhipbdoknhpejdpceibmeaohidgcod'
./install-host -o "chrome-extension://${ID}/" -d 'Chrome/Discord bridge - see https://github.com/p00ya/chrome-discord-bridge)' 'io.github.p00ya.cdb' chrome-discord-bridge
```
