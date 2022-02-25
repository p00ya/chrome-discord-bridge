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

On macOS and Linux, simply delete the manifest file from Chrome's directory:

 *  macOS: `~/Library/Application\ Support/Google/Chrome/NativeMessagingHosts/`
 *  Linux: `~/.config/google-chrome/NativeMessagingHosts`

On Windows, delete the manifest file from whichever directory you ran `install-host` from, and delete the registry key under `HKEY_CURRENT_USER\SOFTWARE\Google\Chrome\NativeMessagingHosts\` using `regedit`.
