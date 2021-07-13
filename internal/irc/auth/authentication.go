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

func GetOAuthToken(conf *oauth2.Config, verifier TokenVerifyier, u Utils) (string, error) {
	state, err := u.NewUUID()
	if err != nil {
		return "", err
	}

	nonce, err := u.NewUUID()
	if err != nil {
		return "", err
	}

	addr, err := buildUserLoginURL(conf, state, nonce)
	if err != nil {
		return "", err
	}

	err = u.OpenURL(addr)
	if err != nil {
		return "", err
	}

	token, err := listenForToken(state, nonce, conf, verifier)
	if err != nil {
		return "", err
	}
	return token, nil
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

func listenForToken(state string, nonce string, conf *oauth2.Config, verifier TokenVerifyier) (accessToken string, err error) {
	u, err := url.Parse(conf.RedirectURL)
	if err != nil {
		return "", err
	}

	var port struct {
		Port string
	}
	port.Port = u.Port()

	t := template.Must(template.New("example").Parse(html))

	svr := &http.Server{Addr: fmt.Sprintf(":%s", u.Port())}
	svr.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			err = t.Execute(w, &port)
			if err != nil {
				err = fmt.Errorf("failed to execute template: %v", err)
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

			if s := getValue(data, "state"); s != state {
				err = errFailedStateValidation
				return
			}

			tkn := getValue(data, "id_token")
			if tkn == "" {
				err = errNoIdToken
				return
			}

			idToken, err = verifier.Verify(context.Background(), tkn)
			if err != nil {
				err = fmt.Errorf("failed to verify access token: %v", err)
				return
			}

			var claims struct {
				Nonce string `json:"nonce"`
			}

			err = idToken.Claims(&claims)
			if err != nil {
				err = fmt.Errorf("failed to decode id_token claims: %v", err)
				return
			}

			if claims.Nonce != nonce {
				err = errFailedNonceValidation
				return
			}

			at := getValue(data, "access_token")
			if at == "" {
				err = errNoAccessToken
				return
			}

			accessToken = at
		}
	})

	if svrErr := svr.ListenAndServe(); svrErr != http.ErrServerClosed {
		return "", err
	}
	return
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
