# ttchat

Connect to a Twitch channel's chat from your terminal

# Installing

  ### With Go

```go get github.com/atye/ttchat```

### Clone and build
```git clone https://github.com/atye/ttchat.git```

```make build```

You should see the binary at `./bin/ttchat`.

# Configuration and Setup

 `ttchat` requires a configuration file in `$HOME/.ttchat/config.yaml` containing some account information for authentication.

```
clientID: "your_twitch_client_id"
username: "your_twitch_login_username"
redirectPort: "9999"
```

`clientID` is your Client ID listed on your application on https://dev.twitch.tv/console.

`username` is your username for logging in.

`redirectPort` is the port that `ttchat` will use to spin up a temporary, localhost server used in the authentication process. If `redirectPort` is empty, port 9999 is used. **Make sure to keep reading.**

Your Twitch application's OAuth Redirect URLs must have a match for the URL of the local server that `ttchat` spins up. If this was your configuration:

```
clientID: "your_twitch_client_id"
username: "your_twitch_login_username"
redirectPort: "8080"
```
`ttchat` will listen on `http://localhost:8080` for Twitch's authentication result. So, your Twitch application must have `http://localhost:8080` for a redirect URL.

If your `redirectPort` is `9000`, `ttchat` will listen on `http://localhost:9000` so your Twitch application must have `http://localhost:9000` for a redirect URL.

# Running
`ttchat -h`

`ttchat --channel ludwig`

`ttchat --channel ludwing --lines 10`
