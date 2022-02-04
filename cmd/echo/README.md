# echo

The `echo` command is a command-line utility for testing the "native messaging host" library implementation in `internal/chrome` with a Chrome extension.

It uses the native messaging host protocol to repeat the request payload back as the response to Chrome.

From the top level repository directory, build it with:

```
go build ./cmd/echo
```

## Testing with Chrome

Open chrome://extensions/ and enable Developer Mode.  Click "Load unpacked" and select the `assets/host-test` subdirectory.  Note the ID that is assigned to the extension (in this example, we assume `nglhipbdoknhpejdpceibmeaohidgcod`).

Register the native messaging host, substituting the appropriate extension ID below:

```
go build ./cmd/install-host

ID='nglhipbdoknhpejdpceibmeaohidgcod'
./install-host -o "chrome-extension://${ID}/" -d 'Echo (native messaging host testing utility - see https://github.com/p00ya/chrome-discord-bridge)' 'io.github.p00ya.cdb_echo' echo
```

Find the "Host Test" extension in Chrome's extensions menu and click it.  From the popup, click the buttons to test the native messaging host.
