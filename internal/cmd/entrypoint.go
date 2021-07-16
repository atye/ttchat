// Copyright © 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/nicklaw5/helix"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"gitub.com/atye/ttchat/internal/auth"
	"gitub.com/atye/ttchat/internal/auth/openid"
	"gitub.com/atye/ttchat/internal/irc"
	"gitub.com/atye/ttchat/internal/irc/client"
	"gitub.com/atye/ttchat/internal/terminal"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/twitch"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ClientID     string `yaml:"clientID"`
	Username     string `yaml:"username"`
	RedirectPort string `yaml:"redirectPort"`
}

const (
	DefaultRedirectPort = "9999"
)

var (
	ErrNoChannel          = errors.New("no channel provided")
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
ttchat --channel ludwig
ttchat --channel ludwig --lines 5
`,
		Run: func(cmd *cobra.Command, args []string) {
			rand.Seed(time.Now().UTC().UnixNano())

			channel, err := cmd.Flags().GetString("channel")
			if err != nil {
				errExit(err)
			}

			if channel == "" {
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

			accessToken, err := getAccessToken(token, conf, oidcVerifier)
			if err != nil {
				errExit(err)
			}

			err = validateAccessToken(accessToken)
			if err != nil {
				errExit(err)
			}

			// Get user display name
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
			//

			// Create IRC client and start
			ircClient := client.NewGempirClient(conf.Username, channel, accessToken)
			c := irc.NewIRCService(displayName, channel, ircClient)
			if tea.NewProgram(terminal.NewModel(0, c), tea.WithAltScreen()).Start() != nil {
				errExit(err)
			}
		},
	}

	rootCmd.Flags().StringP("channel", "c", "", "channel to connect to")
	err := rootCmd.MarkFlagRequired("channel")
	if err != nil {
		errExit(err)
	}

	rootCmd.Flags().StringP("token", "t", "", `oauth token of the from "oauth:token" or "token"`)
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

func getAccessToken(tokenFlagValue string, conf Config, verifier auth.TokenVerifyier) (string, error) {
	if tokenFlagValue != "" {
		s := strings.Split(tokenFlagValue, ":")
		switch len(s) {
		// oauth:token
		case 2:
			return s[1], nil
		// token
		case 1:
			return s[0], nil
		default:
			return "", fmt.Errorf("failed to parse token")
		}
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
