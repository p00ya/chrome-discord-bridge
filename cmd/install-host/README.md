# install-host

The `install-host` command is a command-line utility for registering a Native Messaging Host with Chrome.

Currently it only works on macOS.

From the top level repository directory, build it with:

```
go build ./cmd/install-host
```

Then run it like:

```
./install-host -o 'chrome-extension://foo/' com.example.extension_name path/to/binary
```

It will write a manifest file to Chrome's directory.

## Uninstallation

Delete the manifest file from Chrome's directory (`~/Library/Application\ Support/Google/Chrome/NativeMessagingHosts/` on macOS).