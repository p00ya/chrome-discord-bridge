# install-host

The `install-host` command is a command-line utility for registering a Native Messaging Host with Chrome.

From the top level repository directory, build it with:

```
go build ./cmd/install-host
```

Then run it like:

```
./install-host -o 'chrome-extension://foo/' com.example.extension_name path/to/binary
```

On Linux and macOS, it will write a manifest file to Chrome's directory.  On Windows, it will write the manifest to the working directory, and also write a value to the Windows registry.

## Uninstallation

Delete the manifest file from Chrome's directory (`~/Library/Application\ Support/Google/Chrome/NativeMessagingHosts/` on macOS).