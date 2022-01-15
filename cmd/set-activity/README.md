# set-activity

The `set-activity` command is a command-line utility for setting Discord's Rich Presence for a user with Discord running locally.  Its main purpose is for manually testing the `internal/discord` package (since it doesn't actually integrate with any games).

From the top level repository directory, build it with:

```
go build ./cmd/set-activity
```

Then run it like:

```
./set-activity monkeytype "60s test"
```

It will set your activity status on Discord to "Playing Monkeytype" with a subtitle of "60s test".  You must have the Discord app (not the web app) running locally for it to work.

You must send the process an interrupt (i.e. ^C) to exit, which will revert your Discord activity status.