package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// account is the part of /etc/kvm/pwd a session token needs.
//
// TokenVersion is a random 64-bit number, and it has to be carried exactly: the
// server compares it with the stored one and refuses any other. Decoded as a
// uint64 it is exact. A first version of this tool was a JavaScript one-liner,
// and JSON.parse rounded the 20-digit value to the nearest double, so every
// token it made was refused with a bare 401.
type account struct {
	Username     string `json:"username"`
	Role         string `json:"role"`
	Enabled      *bool  `json:"enabled"`
	TokenVersion uint64 `json:"tokenVersion"`
}

type accountFile struct {
	Users []account `json:"users"`
}

// readMintInput splits what `mint` reads on stdin: the JWT secret on the first
// line, then the whole of /etc/kvm/pwd.
func readMintInput(r io.Reader) (secret string, pwd []byte, err error) {
	br := bufio.NewReader(r)
	line, err := br.ReadString('\n')
	if err != nil {
		return "", nil, fmt.Errorf("reading the secret line: %w", err)
	}
	pwd, err = io.ReadAll(br)
	return strings.TrimSpace(line), pwd, err
}

// mintToken signs a session token for the first enabled administrator, the
// same claims middleware.GenerateJWT writes.
func mintToken(secret string, pwd []byte, now time.Time, ttl time.Duration) (string, string, error) {
	if len(secret) == 0 {
		return "", "", errors.New("empty secret")
	}
	var db accountFile
	if err := json.Unmarshal(pwd, &db); err != nil {
		return "", "", fmt.Errorf("parsing the account file: %w", err)
	}

	var user *account
	for i := range db.Users {
		u := &db.Users[i]
		if u.Role == "admin" && (u.Enabled == nil || *u.Enabled) && u.TokenVersion != 0 {
			user = u
			break
		}
	}
	if user == nil {
		return "", "", errors.New("no enabled administrator with a token version")
	}

	claims := struct {
		Username     string `json:"username"`
		TokenVersion uint64 `json:"tokenVersion"`
		Sub          string `json:"sub"`
		Iat          int64  `json:"iat"`
		Exp          int64  `json:"exp"`
	}{user.Username, user.TokenVersion, user.Username, now.Unix(), now.Add(ttl).Unix()}

	enc := base64.RawURLEncoding
	body, _ := json.Marshal(claims)
	signing := enc.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`)) + "." + enc.EncodeToString(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signing))
	return signing + "." + enc.EncodeToString(mac.Sum(nil)), user.Username, nil
}
