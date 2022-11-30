package auth

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/twitch"
)

type TokenVerifyier interface {
	Verify(context.Context, string) (IDToken, error)
}

type IDToken interface {
	Claims(v interface{}) error
}

type OpenURL func(url string) error
type NewUUID func() (string, error)

var (
	errFailedStateValidation = errors.New("failed state validation")
	errFailedNonceValidation = errors.New("failed nonce validation")
	errNoIdToken             = errors.New("id_token not found")
	errNoAccessToken         = errors.New("access_token not found")
	ErrUnauthorized          = errors.New("unauthorized")
)

func NewTwitchAccessToken(ctx context.Context, clientID string, redirectPort string, verifier TokenVerifyier, openURL OpenURL, newUUID NewUUID) (string, error) {
	state, err := newUUID()
	if err != nil {
		return "", err
	}

	nonce, err := newUUID()
	if err != nil {
		return "", err
	}

	conf := &oauth2.Config{
		ClientID: clientID,
		Scopes:   []string{"openid", "chat:read", "chat:edit"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  twitch.Endpoint.AuthURL,
			TokenURL: twitch.Endpoint.TokenURL,
		},
		RedirectURL: fmt.Sprintf("http://localhost:%s", redirectPort),
	}

	u, err := url.Parse(conf.RedirectURL)
	if err != nil {
		return "", err
	}

	addr, err := buildUserLoginURL(conf, state, nonce)
	if err != nil {
		return "", err
	}

	opts := redirectOpts{
		state:    state,
		nonce:    nonce,
		port:     port{Port: u.Port()},
		t:        template.Must(template.New("example").Parse(html)),
		verifier: verifier,
	}

	errCh := make(chan error)
	tokenCh := make(chan string)

	go listenForRedirect(opts, errCh, tokenCh)

	err = openURL(addr)
	if err != nil {
		return "", err
	}

	select {
	case t := <-tokenCh:
		return t, nil
	case err := <-errCh:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func ValidateTwitchAccessToken(ctx context.Context, accessToken string) error {
	r, err := http.NewRequestWithContext(ctx, "GET", "https://id.twitch.tv/oauth2/validate", nil)
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
		return ErrUnauthorized
	}

	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading error responde body: %w", err)
		}
		return fmt.Errorf("validating access token: %w", errors.New(string(b)))
	}

	return nil
}

func buildUserLoginURL(conf *oauth2.Config, state string, nonce string) (string, error) {
	authURL, err := url.Parse(conf.AuthCodeURL(state))
	if err != nil {
		return "", err
	}
	q := authURL.Query()
	q.Set("nonce", nonce)
	authURL.RawQuery = q.Encode()

	return strings.Replace(authURL.String(), "response_type=code", "response_type=token+id_token", 1), nil
}

type redirectOpts struct {
	state    string
	nonce    string
	port     port
	t        *template.Template
	verifier TokenVerifyier
}

type port struct {
	Port string
}

func listenForRedirect(opts redirectOpts, errCh chan error, tokenCh chan string) {
	svr := &http.Server{Addr: fmt.Sprintf(":%s", opts.port.Port)}
	svr.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			err := opts.t.Execute(w, opts.port)
			if err != nil {
				errCh <- fmt.Errorf("executing template: %v", err)
				return
			}
		case "/callback":
			defer func() {
				go func() {
					svr.Shutdown(context.Background())
				}()
			}()

			data := strings.Split(r.URL.Query().Get("data"), "&")
			var idToken IDToken

			if s := getValue(data, "state"); s != opts.state {
				errCh <- errFailedStateValidation
				return
			}

			tkn := getValue(data, "id_token")
			if tkn == "" {
				errCh <- errNoIdToken
				return
			}

			idToken, err := opts.verifier.Verify(context.Background(), tkn)
			if err != nil {
				errCh <- fmt.Errorf("verifying access token: %v", err)
				return
			}

			var claims struct {
				Nonce string `json:"nonce"`
			}

			err = idToken.Claims(&claims)
			if err != nil {
				errCh <- fmt.Errorf("decoding token claims: %v", err)
				return
			}

			if claims.Nonce != opts.nonce {
				errCh <- errFailedNonceValidation
				return
			}

			accessToken := getValue(data, "access_token")
			if accessToken == "" {
				errCh <- errNoAccessToken
				return
			}

			tokenCh <- accessToken
		}
	})

	if svrErr := svr.ListenAndServe(); svrErr != http.ErrServerClosed {
		errCh <- svrErr
	}
}

func getValue(pairs []string, key string) string {
	for _, p := range pairs {
		values := strings.Split(p, "=")
		if len(values) != 2 {
			continue
		}
		if strings.Contains(values[0], key) {
			return values[1]
		}
	}
	return ""
}

var html = `
<!DOCTYPE html>
<html>

<head>

    <script src="https://code.jquery.com/jquery-3.1.1.min.js"
        integrity="sha256-hVVnYaiADRTO2PzUGmuLJr8BLUSjGIZsDYGmIJLv2b8=" crossorigin="anonymous"></script>

</head>


<body>
    <p id="msg"></p>
    <script>
        if (location.hash.includes('access_token')) {
            $.ajax({
                url: "http://localhost:{{.Port}}/callback",
                data: {
                    data: location.hash
                },
                success: function (response) {
                    $("#msg").text('Thank you! Go back to your terminal.')
                },
                error: function (xhr) {
					$("#msg").text(xhr.responseText)
                },
				statusCode: {
					0: function(response) {
						setTimeout(function() {
							$.ajax(this);
							return
						}, 500)
					}
				}
            })
        }
    </script>

</body>

</html>
`
