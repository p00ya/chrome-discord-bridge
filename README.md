# chrome-discord-bridge

This code bridges Chrome "native messages" to Discord's IPC socket.

It's written in Go with no third-party dependencies.   It's intended to minimize the amount of code running natively, so that the logic for communicating with Discord can be isolated in the Chrome extension's sandbox.  It's much lighter-weight than PreMiD's node.js app.  At the same time, it's easier to reason about the correctness of the code compared to the official C++ Discord-RPC library because of Go's memory management and concurrency primitives.

The bridge itself is WIP as of January 2022.  However, the library functionality for both the Discord and Chrome ends is complete and unit-tested, and there is a utility for manually setting the Discord status (without a Chrome extension) in `cmd/set-activity`.