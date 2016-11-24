package osin

import (
	"encoding/base64"
	"errors"
	"strings"
	"github.com/valyala/fasthttp"
)

// Parse basic authentication header
type BasicAuth struct {
	Username string
	Password string
}

// Parse bearer authentication header
type BearerAuth struct {
	Code string
}

// CheckClientSecret determines whether the given secret matches a secret held by the client.
// Public clients return true for a secret of ""
func CheckClientSecret(client Client, secret string) bool {
	switch client := client.(type) {
	case ClientSecretMatcher:
		// Prefer the more secure method of giving the secret to the client for comparison
		return client.ClientSecretMatches(secret)
	}
	// Fallback to the less secure method of extracting the plain text secret from the client for comparison
	return client.GetSecret() == secret
}

// Return authorization header data
func CheckBasicAuth(r *fasthttp.RequestCtx) (*BasicAuth, error) {
	if getFormValue(r, "Authorization") == "" {
		return nil, nil
	}

	s := strings.SplitN(getFormValue(r, "Authorization"), " ", 2)
	if len(s) != 2 || s[0] != "Basic" {
		return nil, errors.New("Invalid authorization header")
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return nil, err
	}
	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return nil, errors.New("Invalid authorization message")
	}

	return &BasicAuth{Username: pair[0], Password: pair[1]}, nil
}

// Return "Bearer" token from request. The header has precedence over query string.
func CheckBearerAuth(r *fasthttp.RequestCtx) *BearerAuth {
	authHeader := getFormValue(r, "Authorization")
	authForm := getFormValue(r, "code")
	if authHeader == "" && authForm == "" {
		return nil
	}
	token := authForm
	if authHeader != "" {
		s := strings.SplitN(authHeader, " ", 2)
		if (len(s) != 2 || strings.ToLower(s[0]) != "bearer") && token == "" {
			return nil
		}
		//Use authorization header token only if token type is bearer else query string access token would be returned
		if len(s) > 0 && strings.ToLower(s[0]) == "bearer" {
			token = s[1]
		}
	}
	return &BearerAuth{Code: token}
}

// getClientAuth checks client basic authentication in params if allowed,
// otherwise gets it from the header.
// Sets an error on the response if no auth is present or a server error occurs.
func getClientAuth(w *Response, r *fasthttp.RequestCtx, allowQueryParams bool) *BasicAuth {

	if allowQueryParams {
		// Allow for auth without password
		client_secret := getFormValue(r, "client_secret")
		if client_secret != "" {
			auth := &BasicAuth{
				Username: getFormValue(r, "client_id"),
				Password: client_secret,
			}
			if auth.Username != "" {
				return auth
			}
		}
	}

	auth, err := CheckBasicAuth(r)
	if err != nil {
		w.SetError(E_INVALID_REQUEST, "")
		w.InternalError = err
		return nil
	}
	if auth == nil {
		w.SetError(E_INVALID_REQUEST, "")
		w.InternalError = errors.New("Client authentication not sent")
		return nil
	}
	return auth
}

func getFormValue(r *fasthttp.RequestCtx, key string) string {
	ret := r.FormValue(key)
	//ret := r.QueryArgs().Peek(key)
	//if ret == nil {
	//	r.FormValue(key)
	//	ret = r.PostArgs().Peek(key)
	//}
	if ret == nil {
		println("getFormValue ", key, ":")
		return ""
	}
	println("getFormValue ", key, ":", string(ret))
	return string(ret)
}