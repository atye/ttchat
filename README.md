# ttchat

Connect to a Twitch channel's chat from your terminal

# Installing

### Download release

See https://github.com/atye/ttchat/releases.

### With Go

```go get github.com/atye/ttchat```

### Clone and build
```git clone https://github.com/atye/ttchat.git```

```make build```

You should see the binary at `./bin/ttchat`.

# Configuration and Setup

 A configuration file at `$HOME/.ttchat/config.yaml` containing some account information is required for authentication.

```
clientID: "your_twitch_client_id"
username: "your_twitch_login_username"
redirectPort: "9999"
```

`clientID` is your Client ID listed on your application on https://dev.twitch.tv/console.

`username` is your username for logging in.

`redirectPort` is the port that `ttchat` will use to spin up a temporary, localhost server used in the authentication process. If `redirectPort` is empty, port 9999 is used.

Your Twitch application's list of OAuth Redirect URLs must have a match for the URL of `ttchat`. If this was your configuration:

```
clientID: "123"
username: "foo"
redirectPort: "8080"
```

`ttchat` would listen on `http://localhost:8080` for Twitch's authentication result. So, your Twitch application must have `http://localhost:8080` for a redirect URL.

# Running
`ttchat -h`

`ttchat --channel ludwig`
