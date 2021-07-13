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

package openid

import (
	"context"

	"github.com/atye/ttchat/internal/irc/auth"
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
