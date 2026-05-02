// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"net/http"
	"time"
)

// ClientFor returns an *http.Client that injects the correct Mendix auth
// headers for the resolved credential of the given profile.
//
// The client refuses to send requests to hosts that are not known Mendix
// platform endpoints (see scheme.go). This is a defence against accidentally
// leaking tokens to the wrong service.
func ClientFor(ctx context.Context, profile string) (*http.Client, error) {
	cred, err := Resolve(ctx, profile)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: &authTransport{cred: cred, inner: http.DefaultTransport},
		Timeout:   30 * time.Second,
	}, nil
}

// NewClient returns an *http.Client bound to the given credential.
// Useful for tests and for callers that already have a Credential in hand.
func NewClient(cred *Credential) *http.Client {
	return &http.Client{
		Transport: &authTransport{cred: cred, inner: http.DefaultTransport},
		Timeout:   30 * time.Second,
	}
}

type authTransport struct {
	cred  *Credential
	inner http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	scheme, known := SchemeForHost(req.URL.Host)
	if !known {
		return nil, &ErrUnknownHost{Host: req.URL.Host}
	}
	if scheme != t.cred.Scheme {
		return nil, &ErrSchemeMismatch{Host: req.URL.Host, Need: scheme, Have: t.cred.Scheme}
	}
	// Clone to avoid mutating the caller's request headers.
	req = req.Clone(req.Context())
	switch scheme {
	case SchemePAT:
		req.Header.Set("Authorization", "MxToken "+t.cred.Token)
	}
	resp, err := t.inner.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	// Mendix platform APIs return 401 when no valid credential is presented,
	// and 403 when the PAT is invalid/expired (per the portal PAT docs).
	// Both mean "credential rejected" for our purposes.
	// Close the body and return (nil, err) to follow the RoundTripper contract —
	// callers must not see both a non-nil resp and a non-nil err.
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		_ = resp.Body.Close()
		return nil, &ErrUnauthenticated{Profile: t.cred.Profile}
	}
	return resp, nil
}
