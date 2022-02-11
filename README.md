# chrome-discord-bridge

This code bridges Chrome "native messages" to Discord's IPC socket.

It's written in Go with no third-party dependencies.   It's intended to minimize the amount of code running natively, so that the logic for communicating with Discord can be isolated in the Chrome extension's sandbox.  It's much lighter-weight than PreMiD's node.js app.  At the same time, it's easier to reason about the correctness of the code compared to the official C++ Discord-RPC library because of Go's memory management and concurrency primitives.

chrome-discord-bridge is considered stable.

## Installation and Usage

Download an appropriate binary for your computer from the Github [releases page](https://github.com/p00ya/chrome-discord-bridge/releases).  Windows is not currently supported!

Extract the archive with:

```
tar -xf chrome-discord-bridge_*.tar.xz
```

On macOS, you will also need to run:

```
xattr -d com.apple.quarantine chrome-discord-bridge
```

Then register the utility with Chrome (for the current user) with:

```
./chrome-discord-bridge -install
```

Note that you will have to re-run this last command if you move the binary, because the Chrome manifest contains an absolute path to the binary.

### Browser Activity

chrome-discord-bridge is intended to be paired with the [Browser Activity](https://github.com/p00ya/browser-activity) Chrome extension.

## Development

Each Chrome extension that will be used with `chrome-discord-bridge` must be added to `cmd/chrome-discord-bridge/origins.txt`.  This list of allowed origins is part of the manifest file that Chrome enforces, and is also built into the binary as an additional layer of security.

You can determine Chrome extension IDs by loading chrome://extensions in Chrome.  For example, to add the Chrome extension with the ID `nglhipbdoknhpejdpceibmeaohidgcod`, add a line in `origins.txt` like:

```
chrome-extension://nglhipbdoknhpejdpceibmeaohidgcod/
```

Then with Go 1.17+, run:

```
go build ./cmd/chrome-discord-bridge
```

This will build the `chrome-discord-bridge` binary.  To write a manifest for the Native Messaging Host to Chrome (for just the current system user), run:

```
./chrome-discord-bridge -install
```

You will need to re-run the previous command if the path to the binary changes.
