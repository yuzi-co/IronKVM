package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

const pwdFixture = `{"users":[
 {"username":"viewer","role":"user","enabled":true,"tokenVersion":5},
 {"username":"off","role":"admin","enabled":false,"tokenVersion":6},
 {"username":"admin","role":"admin","enabled":true,"password":"not-read","tokenVersion":13831960392123456789}]}`

func TestMintCarriesTheTokenVersionExactly(t *testing.T) {
	// The value is above 2^53, where a double cannot hold every integer.
	tok, user, err := mintToken("s3cret", []byte(pwdFixture), time.Unix(1000, 0), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if user != "admin" {
		t.Fatalf("minted for %q, want the enabled administrator", user)
	}

	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatalf("not a JWT: %q", tok)
	}
	mac := hmac.New(sha256.New, []byte("s3cret"))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if base64.RawURLEncoding.EncodeToString(mac.Sum(nil)) != parts[2] {
		t.Fatal("the signature does not verify with the secret")
	}

	body, _ := base64.RawURLEncoding.DecodeString(parts[1])
	if !strings.Contains(string(body), `"tokenVersion":13831960392123456789`) {
		t.Fatalf("token version not carried exactly: %s", body)
	}
	var c struct {
		Username, Sub string
		Iat, Exp      int64
	}
	_ = json.Unmarshal(body, &c)
	if c.Username != "admin" || c.Sub != "admin" || c.Exp-c.Iat != 3600 {
		t.Fatalf("claims wrong: %+v", c)
	}
}

func TestMintRefusesAFileWithNoAdministrator(t *testing.T) {
	if _, _, err := mintToken("s", []byte(`{"users":[{"username":"u","role":"user","tokenVersion":1}]}`), time.Now(), time.Hour); err == nil {
		t.Fatal("expected an error")
	}
}

func TestReadMintInputSplitsSecretFromAccounts(t *testing.T) {
	s, pwd, err := readMintInput(strings.NewReader("abc  \n{\"users\":[]}\n"))
	if err != nil || s != "abc" || !strings.HasPrefix(string(pwd), `{"users"`) {
		t.Fatalf("got %q %q %v", s, pwd, err)
	}
}
