// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"

	"github.com/dgrijalva/jwt-go"
)

// Validator data model
type Validator struct {
	clientID      string
	usernameClaim string
	algorithm     string
	wellKnownURL  string
}

// NewAuthValidator creates a new auth0 validator based on the specified parameters
func NewAuthValidator(domain, clientID, usernameClaim, algorithm string) (Validator, error) { // nolint
	if domain == "" {
		return Validator{}, errors.New("missing Domain")
	}
	if clientID == "" {
		return Validator{}, errors.New("missing ClientID")
	}
	if usernameClaim == "" {
		return Validator{}, errors.New("missing UsernameClaim")
	}
	if algorithm == "" {
		return Validator{}, errors.New("missing Algorithm")
	}

	validator := Validator{
		clientID:      clientID,
		usernameClaim: usernameClaim,
		algorithm:     algorithm,
		wellKnownURL:  "https://" + path.Join(domain, ".well-known/jwks.json"),
	}

	return validator, nil
}

// VerifyToken verifies the specified token
func (av Validator) VerifyToken(token string) (map[string]interface{}, error) {
	// Using jwt.MapClaims because our username field is set dynamically
	// based on environment
	claims := jwt.MapClaims{}
	jwtToken, err := jwt.ParseWithClaims(token, claims, av.getPemCert)
	if err != nil {
		return nil, err
	}
	if !jwtToken.Valid {
		return nil, errors.New("invalid token")
	}

	allClaims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("unable to map claims")
	}

	if err = allClaims.Valid(); err != nil {
		return nil, errors.New("claims are not valid")
	}

	return allClaims, nil
}

type jwks struct {
	Keys []jsonWebKeys `json:"keys"`
}

type jsonWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

func (av Validator) getPemCert(token *jwt.Token) (interface{}, error) {
	cert := ""
	resp, err := http.Get(av.wellKnownURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var jwks = jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)
	if err != nil {
		return "", err
	}

	for k := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		return "", errors.New("unable to find appropriate key")
	}

	rsaPublicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
	if err != nil {
		return "", err
	}

	return rsaPublicKey, nil
}
