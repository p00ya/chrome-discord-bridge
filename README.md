# chrome-discord-bridge

chrome-discord-bridge acts as a "Chrome native messaging host", and forwards Chrome "native messages" to Discord's IPC socket.

It's considered stable.

## Installation and Usage

chrome-discord-bridge is intended to be paired with the "Browser Activity" Chrome extension.  See the instructions at https://p00ya.github.io/browser-activity on how to install chrome-discord-bridge and the extension.

## Security

chrome-discord-bridge runs natively with no sandbox.  It's been designed to be easy to audit, so that users can be confident installing it.

There's not much source code.  The logic is fairly minimal because chrome-discord-bridge doesn't do much more than proxy bytes between stdin/stdout and Discord's IPC socket.

On macOS and Linux, it has no dependencies (other than the Go standard library), so there's no implicit trust in third-party software.  On Windows, it depends only on official packages from Google and Microsoft for registry and named pipes APIs.

The source code is written in Go and easy to read.  Go's primitives make it easy to reason about concurrency and memory management.  Go's static typing provides compile-time guarantees about correctness.

It's easy to verify the code works, because it's well-tested.  There are unit tests, and also supplementary utilities for manually testing with Chrome and Discord.

From within the browser, only trusted Chrome extensions can invoke chrome-discord-bridge.  The chrome-discord-bridge binary hardcodes a list of allowed origins (extension IDs).  There are two layers of checks: one enforced by Chrome using the installation manifest, and another within chrome-discord-bridge itself when it checks its command-line arguments.

## Development

Each Chrome extension that will be used with `chrome-discord-bridge` must be added to `cmd/chrome-discord-bridge/origins.txt`.

You can determine Chrome extension IDs by loading chrome://extensions in Chrome.  For example, to add the Chrome extension with the ID `nglhipbdoknhpejdpceibmeaohidgcod`, add a line in `origins.txt` like:

```
chrome-extension://nglhipbdoknhpejdpceibmeaohidgcod/
```

Then with Go 1.17+, run:

    go build ./cmd/chrome-discord-bridge

This will build the `chrome-discord-bridge` binary.  To write a manifest for the Native Messaging Host to Chrome (for just the current system user), run:

    ./chrome-discord-bridge -install


You will need to re-run the previous command if the path to the binary changes.

### Testing

Unit tests can be run with:

    go test -v ./internal/...

Some supplementary utilities for manually testing with Chrome and Discord can be found in `cmd/echo` and `cmd/set-activity` respectively.  Consult the `README.md` files in those sub-directories for additional information.
