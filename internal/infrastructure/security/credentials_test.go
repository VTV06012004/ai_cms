package security

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
)

func TestHashPasswordUsesArgon2id(t *testing.T) {
	encoded, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if !strings.HasPrefix(encoded, "$argon2id$v=19$") {
		t.Fatalf("password hash = %q, want Argon2id encoding", encoded)
	}
	if !CheckPasswordHash("correct horse battery staple", encoded) {
		t.Fatal("password hash did not verify")
	}
	if CheckPasswordHash("incorrect", encoded) {
		t.Fatal("incorrect password verified")
	}
	if PasswordHashNeedsUpgrade(encoded) {
		t.Fatal("current Argon2id hash should not need an upgrade")
	}
}

func TestCheckPasswordHashSupportsLegacyPBKDF2AndRequestsUpgrade(t *testing.T) {
	expected, err := hex.DecodeString("120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b")
	if err != nil {
		t.Fatal(err)
	}
	encoded := "$pbkdf2-sha256$1$" +
		base64.RawStdEncoding.EncodeToString([]byte("salt")) + "$" +
		base64.RawStdEncoding.EncodeToString(expected)

	if !CheckPasswordHash("password", encoded) {
		t.Fatal("valid legacy PBKDF2 hash did not verify")
	}
	if PasswordHashNeedsUpgrade(encoded) != true {
		t.Fatal("legacy PBKDF2 hash should be upgraded after successful login")
	}
	if CheckPasswordHash("wrong", encoded) {
		t.Fatal("incorrect password verified against legacy hash")
	}
}

func TestCheckPasswordHashRejectsMalformedArgon2idParameters(t *testing.T) {
	for _, encoded := range []string{
		"$argon2id$v=19$m=999999999,t=3,p=2$c2FsdA$aGFzaA",
		"$argon2id$v=18$m=65536,t=3,p=2$c2FsdA$aGFzaA",
		"$argon2id$v=19$m=65536,t=3,p=2$!$!",
	} {
		if CheckPasswordHash("password", encoded) {
			t.Errorf("malformed hash %q verified", encoded)
		}
	}
}
