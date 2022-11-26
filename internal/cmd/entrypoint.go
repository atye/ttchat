package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/atye/ttchat/internal/auth"
	"github.com/atye/ttchat/internal/auth/openid"
	"github.com/atye/ttchat/internal/irc"
	"github.com/atye/ttchat/internal/irc/client"
	"github.com/atye/ttchat/internal/terminal"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/nicklaw5/helix"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/twitch"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ClientID     string `yaml:"clientID"`
	Username     string `yaml:"username"`
	RedirectPort string `yaml:"redirectPort"`
	LineSpacing  int    `yaml:"lineSpacing"`
}

const (
	DefaultRedirectPort = "9999"
)

var (
	ErrNoChannel          = errors.New("no channels provided")
	ErrNoClientID         = errors.New("no clientID in configuration file")
	ErrNoUsername         = errors.New("no username in configuration file")
	ErrInvalidAccessToken = errors.New("invalid access token")
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ttchat",
		Short: "Connect to twitch chat in your terminal",
		Long: `
ttchat is a terminal application that connects to a twitch channel's
chat using a small configuration file. See repo for more details.

ttchat -h
ttchat --channel GothamChess --channel chessbrah
ttchat --channel GothamChess --token $TOKEN
`,
		Run: func(cmd *cobra.Command, args []string) {
			rand.Seed(time.Now().UTC().UnixNano())

			logger := log.New(io.Discard, "", log.LstdFlags)

			channels, err := cmd.Flags().GetStringSlice("channel")
			if err != nil {
				errExit(err)
			}

			if len(channels) == 0 {
				errExit(ErrNoChannel)
			}

			token, err := cmd.Flags().GetString("token")
			if err != nil {
				errExit(err)
			}

			hd, err := os.UserHomeDir()
			if err != nil {
				errExit(err)
			}

			conf, err := getConfig(hd)
			if err != nil {
				errExit(err)
			}

			provider, err := oidc.NewProvider(context.Background(), "https://id.twitch.tv/oauth2")
			if err != nil {
				errExit(err)
			}
			oidcVerifier := openid.CoreOSVerifier{Verifier: provider.Verifier(&oidc.Config{ClientID: conf.ClientID})}

			accessToken, err := getAccessToken(logger, token, conf, oidcVerifier)
			if err != nil {
				errExit(err)
			}

			err = validateAccessToken(accessToken)
			if err != nil {
				errExit(err)
			}

			tc, err := helix.NewClient(&helix.Options{
				ClientID:        conf.ClientID,
				UserAccessToken: accessToken,
			})
			if err != nil {
				errExit(err)
			}

			displayName, err := getUserDisplayName(conf, accessToken, tc)
			if err != nil {
				errExit(err)
			}

			var channelModels []*terminal.Channel
			for _, c := range channels {
				conn := irc.NewTwitch(client.NewGempirClient(conf.Username, c, accessToken), logger, displayName, c)
				channelModels = append(channelModels, terminal.NewChannel(conn, c, conf.LineSpacing))
			}

			if tea.NewProgram(terminal.NewModel(logger, channelModels...), tea.WithAltScreen()).Start() != nil {
				errExit(err)
			}
		},
	}

	rootCmd.Flags().StringSliceP("channel", "c", []string{}, "channels to connect to")
	err := rootCmd.MarkFlagRequired("channel")
	if err != nil {
		errExit(err)
	}

	rootCmd.Flags().StringP("token", "t", "", `provide your own oauth access token to bypass browser login (must have chat:read and chat:edit scopes)`)
	return rootCmd
}

func getConfig(hd string) (Config, error) {
	f, err := os.ReadFile(filepath.Join(hd, ".ttchat", "config.yaml"))
	if err != nil {
		return Config{}, err
	}

	var conf Config
	err = yaml.Unmarshal(f, &conf)
	if err != nil {
		return Config{}, err
	}

	if conf.ClientID == "" {
		return Config{}, ErrNoClientID
	}

	if conf.Username == "" {
		return Config{}, ErrNoUsername
	}

	if conf.RedirectPort == "" {
		conf.RedirectPort = DefaultRedirectPort
	}

	return conf, nil
}

func getAccessToken(logger *log.Logger, tokenFlagValue string, conf Config, verifier auth.TokenVerifyier) (string, error) {
	if tokenFlagValue != "" {
		return tokenFlagValue, nil
	}

	oauthConf := &oauth2.Config{
		ClientID: conf.ClientID,
		Scopes:   []string{"openid", "chat:read", "chat:edit"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  twitch.Endpoint.AuthURL,
			TokenURL: twitch.Endpoint.TokenURL,
		},
		RedirectURL: fmt.Sprintf("http://localhost:%s", conf.RedirectPort),
	}

	f := func() (string, error) {
		u, err := uuid.NewUUID()
		if err != nil {
			return "", err
		}
		return u.String(), nil
	}

	u := auth.Utils{
		OpenURL: browser.OpenURL,
		NewUUID: f,
	}

	t, err := auth.GetOAuthToken(oauthConf, verifier, u)
	if err != nil {
		return "", err
	}
	return t, nil
}

func validateAccessToken(accessToken string) error {
	r, err := http.NewRequest("GET", "https://id.twitch.tv/oauth2/validate", nil)
	if err != nil {
		return err
	}

	r.Header.Set("Authorization", fmt.Sprintf("OAuth %s", accessToken))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrInvalidAccessToken
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invaild access token: status code: %d", resp.StatusCode)
	}

	return nil
}

type twitchAPI interface {
	GetUsers(params *helix.UsersParams) (*helix.UsersResponse, error)
}

func getUserDisplayName(conf Config, accessToken string, api twitchAPI) (string, error) {
	resp, err := api.GetUsers(&helix.UsersParams{Logins: []string{conf.Username}})
	if err != nil {
		return "", err
	}
	if resp.ErrorMessage != "" {
		return "", fmt.Errorf(resp.ErrorMessage)
	}

	displayName := conf.Username
	if len(resp.Data.Users) >= 1 {
		if n := resp.Data.Users[0].DisplayName; n != "" {
			displayName = n
		}
	}
	return displayName, nil
}

func errExit(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}
