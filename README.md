# ttchat

![](demo.gif)

# Installing

### Download release

See https://github.com/atye/ttchat/releases.

### Clone and build
```
git clone https://github.com/atye/ttchat.git && cd ttchat
make build
bin/ttchat -h
```

# Setup

A configuration file at `$HOME/.ttchat/config.yaml` containing some account information is required. Optional parameters related to configuration are also available.
 
Suggested example:

```
clientID: "yourTwitchClientId"
username: "yourTwitchUsername"
lineSpacing: 1
```

| Parameter      | Description | Required |
| ----------- | ----------- | ----------- |
| clientID      | your Client ID listed on your application at https://dev.twitch.tv/console       | yes |
| username      | your username for logging in       | yes |
| lineSpacing      | the number of empty lines to put between messages       | no |
| redirectPort      | the port that `ttchat` will use to listen for Twitch's authorization result (default "9999")  | no |

Your Twitch application's list of OAuth Redirect URLs must have a match for the URL of `ttchat` which is `http://localhost:9999` by default.

Using the above suggested example, your Twitch application must have `http://localhost:9999` for an OAuth Redirect URL.

# Running

`ttchat --channel sodapoppin`

`ttchat --channel sodapoppin --channel hasanabi`

Obtaining an OAuth access token requires your authorization via web browser. See https://dev.twitch.tv/docs/authentication/getting-tokens-oauth for more details. To provide your own token, use the `--token` flag. The token must have the `chat:edit` and `chat:read` scopes.

`ttchat --channel sodapoppin --token $TOKEN`

# Usage

| Key      | Description |
| ----------- | ----------- |
| Tab/ShiftTab      | Next/previous channel       |
