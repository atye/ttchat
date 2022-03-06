package auth

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
)

type TokenVerifyier interface {
	Verify(context.Context, string) (IDToken, error)
}

type IDToken interface {
	Claims(v interface{}) error
}

type Utils struct {
	OpenURL func(url string) error
	NewUUID func() (string, error)
}

var (
	errFailedStateValidation = errors.New("failed state validation")
	errFailedNonceValidation = errors.New("failed nonce validation")
	errNoIdToken             = errors.New("id_token not found")
	errNoAccessToken         = errors.New("access_token not found")
)

func GetOAuthToken(conf *oauth2.Config, verifier TokenVerifyier, util Utils) (string, error) {
	state, err := util.NewUUID()
	if err != nil {
		return "", err
	}

	nonce, err := util.NewUUID()
	if err != nil {
		return "", err
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

	err = util.OpenURL(addr)
	if err != nil {
		return "", err
	}

	select {
	case t := <-tokenCh:
		return t, nil
	case err := <-errCh:
		return "", err
	}
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
				errCh <- fmt.Errorf("failed to execute template: %v", err)
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
				errCh <- fmt.Errorf("failed to verify access token: %v", err)
				return
			}

			var claims struct {
				Nonce string `json:"nonce"`
			}

			err = idToken.Claims(&claims)
			if err != nil {
				errCh <- fmt.Errorf("failed to decode id_token claims: %v", err)
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
