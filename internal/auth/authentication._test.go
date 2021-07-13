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
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/twitch"
)

const (
	clientID = "1234"
)

type verifier struct{}

func (v verifier) Verify(ctx context.Context, tkn string) (IDToken, error) {
	return token{}, nil
}

type token struct{}

func (t token) Claims(v interface{}) error {
	err := json.Unmarshal(json.RawMessage([]byte(`{"nonce": "000"}`)), v)
	if err != nil {
		return err
	}
	return nil
}

func TestGetOAuthToken(t *testing.T) {
	conf := &oauth2.Config{
		ClientID: clientID,
		Scopes:   []string{"openid", "chat:read", "chat:edit"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  twitch.Endpoint.AuthURL,
			TokenURL: twitch.Endpoint.TokenURL,
		},
		RedirectURL: fmt.Sprintf("http://localhost:%s", "9999"),
	}

	want := "123"

	u := Utils{
		OpenURL: func(url string) error {
			go func() {
				for {
					r, err := http.NewRequest("GET", "http://localhost:9999/callback", nil)
					if err != nil {
						panic(err)
					}

					q := r.URL.Query()
					q.Add("data", fmt.Sprintf("#access_token=123&id_token=%s&state=000", want))
					r.URL.RawQuery = q.Encode()

					callbackErr := fmt.Errorf("")
					for callbackErr != nil {
						_, callbackErr = http.DefaultClient.Do(r)
						time.Sleep(500 * time.Millisecond)
					}
				}
			}()
			return nil
		},
		NewUUID: func() (string, error) {
			return "000", nil
		},
	}

	tkn, err := GetOAuthToken(conf, verifier{}, u)
	if err != nil {
		t.Fatal(err)
	}

	if tkn != want {
		t.Errorf("expected access token %s, got %s", want, tkn)
	}
}
