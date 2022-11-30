package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
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

func TestNewTwitchAccessToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		want := "123"

		openURL := func(url string) error {
			go func() {
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
			}()
			return nil
		}

		got, err := NewTwitchAccessToken(context.Background(), "clientID", "9999", verifier{}, openURL, func() (string, error) {
			return "000", nil
		})
		if err != nil {
			t.Fatal(err)
		}

		if got != want {
			t.Errorf("expected access token %s, got %s", want, got)
		}
	})
}
