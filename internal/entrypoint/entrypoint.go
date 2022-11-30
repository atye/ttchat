package entrypoint

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/atye/ttchat/internal/auth"
	"github.com/atye/ttchat/internal/irc"
	"github.com/atye/ttchat/internal/irc/client"
	"github.com/atye/ttchat/internal/terminal"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/nicklaw5/helix"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

const (
	DefaultRedirectPort = "9999"
)

var (
	newUUID = func() (string, error) {
		u, err := uuid.NewUUID()
		if err != nil {
			return "", err
		}
		return u.String(), nil
	}
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ttchat",
		Short: "Connect to twitch chats in your terminal",
		Long: `
ttchat -h
ttchat --channel GothamChess --channel chessbrah
ttchat --channel GothamChess --token $TOKEN
`,
		Run: func(cmd *cobra.Command, args []string) {
			logger := log.New(io.Discard, "", log.LstdFlags)

			channels, err := cmd.Flags().GetStringSlice("channel")
			if err != nil {
				errExit(err)
			}

			if len(channels) == 0 {
				errExit(fmt.Errorf("no channels provided"))
			}

			accessToken, err := cmd.Flags().GetString("token")
			if err != nil {
				errExit(err)
			}

			hd, err := os.UserHomeDir()
			if err != nil {
				errExit(err)
			}

			conf, err := newConfig(hd)
			if err != nil {
				errExit(err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if accessToken != "" {
				err = auth.ValidateTwitchAccessToken(ctx, accessToken)
				if err != nil {
					errExit(err)
				}
			} else {
				accessToken = conf.Token
				err = auth.ValidateTwitchAccessToken(ctx, accessToken)
				if err != nil {
					switch err {
					case auth.ErrUnauthorized:
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()

						provider, err := oidc.NewProvider(ctx, "https://id.twitch.tv/oauth2")
						if err != nil {
							errExit(err)
						}
						verifier := auth.CoreOSVerifier{Verifier: provider.Verifier(&oidc.Config{ClientID: conf.ClientID})}

						accessToken, err = auth.NewTwitchAccessToken(ctx, conf.ClientID, conf.RedirectPort, verifier, browser.OpenURL, newUUID)
						if err != nil {
							errExit(err)
						}
					default:
						errExit(err)
					}
				}
			}

			conf.Token = accessToken
			err = conf.save()
			if err != nil {
				errExit(err)
			}

			helix, err := helix.NewClient(&helix.Options{
				ClientID:        conf.ClientID,
				UserAccessToken: accessToken,
			})
			if err != nil {
				errExit(err)
			}

			displayName, err := getUserDisplayName(helix, conf.Username, accessToken)
			if err != nil {
				errExit(err)
			}

			var channelModels []*terminal.Channel
			for _, c := range channels {
				conn := irc.NewTwitch(client.NewGempirClient(logger, conf.Username, c, accessToken), logger, displayName, c)
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

type twitchAPI interface {
	GetUsers(params *helix.UsersParams) (*helix.UsersResponse, error)
}

func getUserDisplayName(twitch twitchAPI, username string, accessToken string) (string, error) {
	resp, err := twitch.GetUsers(&helix.UsersParams{Logins: []string{username}})
	if err != nil {
		return "", fmt.Errorf("getting %s's display name: %w", username, err)
	}
	if resp.ErrorMessage != "" {
		return "", fmt.Errorf("getting %s's display name: %w", username, fmt.Errorf(resp.ErrorMessage))
	}

	displayName := username
	if len(resp.Data.Users) >= 1 {
		if n := resp.Data.Users[0].DisplayName; n != "" {
			displayName = n
		}
	}
	return displayName, nil
}

func errExit(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
