package openid

import (
	"context"

	"github.com/atye/ttchat/internal/auth"
	"github.com/coreos/go-oidc/v3/oidc"
)

type CoreOSVerifier struct {
	Verifier *oidc.IDTokenVerifier
}

var _ auth.TokenVerifyier = CoreOSVerifier{}

func (v CoreOSVerifier) Verify(ctx context.Context, rawToken string) (auth.IDToken, error) {
	t, err := v.Verifier.Verify(context.Background(), rawToken)
	if err != nil {
		return nil, err
	}
	return t, nil
}
