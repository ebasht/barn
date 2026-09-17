package servers

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestParsePrivateKeyAcceptsOpenSSHPEM(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	block, err := ssh.MarshalPrivateKey(key, "")
	if err != nil {
		t.Fatal(err)
	}
	pemBytes := pem.EncodeToMemory(block)

	signer, err := parsePrivateKey(pemBytes, "")
	if err != nil {
		t.Fatalf("parsePrivateKey: %v", err)
	}
	if signer == nil {
		t.Fatal("expected signer")
	}
}

func TestParsePrivateKeyRequiresPassphraseWhenEncrypted(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	block, err := ssh.MarshalPrivateKeyWithPassphrase(key, "", []byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	pemBytes := pem.EncodeToMemory(block)

	if _, err := parsePrivateKey(pemBytes, ""); err == nil {
		t.Fatal("expected passphrase required error")
	} else if !strings.Contains(err.Error(), "passphrase") {
		t.Fatalf("unexpected error: %v", err)
	}

	signer, err := parsePrivateKey(pemBytes, "secret")
	if err != nil {
		t.Fatalf("parse with passphrase: %v", err)
	}
	if signer == nil {
		t.Fatal("expected signer")
	}
}

func TestSSHAuthMethodsRequiresCredential(t *testing.T) {
	if _, err := sshAuthMethods(SSHAuth{}); err == nil {
		t.Fatal("expected error for empty auth")
	}
	methods, err := sshAuthMethods(SSHAuth{Password: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if len(methods) != 1 {
		t.Fatalf("got %d methods", len(methods))
	}
}
