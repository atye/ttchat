# ttchat

Connect to a Twitch channel's chat from your terminal.

![](demo.gif)

# Installing

### Download release

See https://github.com/atye/ttchat/releases.

### Clone and build
```git clone https://github.com/atye/ttchat.git```

```make build```

You should see the binary at `./bin/ttchat`.

# Configuration and Setup

 A configuration file at `$HOME/.ttchat/config.yaml` containing some account information is required for authentication.

```
clientID: "yourTwitchClientId"
username: "yourTwitchUsername"
redirectPort: "9999"
lineSpacing: 1
```

`clientID` is your Client ID listed on your application on https://dev.twitch.tv/console.

`username` is your username for logging in.

`redirectPort` is the port that `ttchat` will use to spin up a temporary, localhost server used in the authentication process if you don't provide your own access token (see [Running](#running)). If `redirectPort` is empty, port 9999 is used.

`lineSpacing` is the number of empty lines to put between messages. If `lineSpacing` is empty, there will be 0 line spaces between messages.

Your Twitch application's list of OAuth Redirect URLs must have a match for the URL of `ttchat`. If this was your configuration:

```
clientID: "123"
username: "foo"
redirectPort: "8080"
```

`ttchat` would listen on `http://localhost:8080` for Twitch's authentication result. So, your Twitch application must have `http://localhost:8080` for a redirect URL.

# Running

`ttchat --channel sodapoppin`

`ttchat --channel sodapoppin --channel hasanabi`

Obtaining an OAuth access token requires your authorization via web browser. See https://dev.twitch.tv/docs/authentication/getting-tokens-oauth for more details. To provide your own token, use the `--token` flag. The token must have the `chat:edit` and `chat:read` scopes.

`ttchat --channel sodapoppin --token $TOKEN`

Use `Tab` and `ShiftTab` to navigate channels.
