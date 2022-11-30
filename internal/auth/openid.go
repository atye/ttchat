package auth

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
)

type CoreOSVerifier struct {
	Verifier *oidc.IDTokenVerifier
}

var _ TokenVerifyier = CoreOSVerifier{}

func (v CoreOSVerifier) Verify(ctx context.Context, rawToken string) (IDToken, error) {
	t, err := v.Verifier.Verify(context.Background(), rawToken)
	if err != nil {
		return nil, err
	}
	return t, nil
}
