package servers

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHAuth holds ephemeral SSH credentials. Prefer password or private key;
// both may be provided (client tries them in order). Never persisted to DB.
type SSHAuth struct {
	Password             string
	PrivateKey           string
	PrivateKeyPassphrase string
}

func takeSSHAuth(password, privateKey, passphrase string) SSHAuth {
	return SSHAuth{
		Password:             password,
		PrivateKey:           strings.TrimSpace(privateKey),
		PrivateKeyPassphrase: passphrase,
	}
}

func (a SSHAuth) empty() bool {
	return strings.TrimSpace(a.Password) == "" && a.PrivateKey == ""
}

func (a *SSHAuth) clear() {
	if a == nil {
		return
	}
	clearString(&a.Password)
	clearString(&a.PrivateKey)
	clearString(&a.PrivateKeyPassphrase)
}

func sshAuthMethods(a SSHAuth) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod
	if pass := strings.TrimSpace(a.Password); pass != "" {
		methods = append(methods, ssh.Password(pass))
	}
	if a.PrivateKey != "" {
		signer, err := parsePrivateKey([]byte(a.PrivateKey), a.PrivateKeyPassphrase)
		if err != nil {
			return nil, err
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}
	if len(methods) == 0 {
		return nil, fmt.Errorf("%w: password or private_key is required", ErrInvalidInput)
	}
	return methods, nil
}

func parsePrivateKey(pemBytes []byte, passphrase string) (ssh.Signer, error) {
	pemBytes = []byte(strings.TrimSpace(string(pemBytes)))
	if len(pemBytes) == 0 {
		return nil, fmt.Errorf("%w: private_key is empty", ErrInvalidInput)
	}
	if passphrase != "" {
		signer, err := ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(passphrase))
		if err != nil {
			return nil, fmt.Errorf("%w: invalid private key or passphrase: %v", ErrInvalidInput, err)
		}
		return signer, nil
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		var missing *ssh.PassphraseMissingError
		if errors.As(err, &missing) {
			return nil, fmt.Errorf("%w: private key is encrypted; private_key_passphrase is required", ErrInvalidInput)
		}
		return nil, fmt.Errorf("%w: invalid private key: %v", ErrInvalidInput, err)
	}
	return signer, nil
}

func sshDial(host string, port int, user string, auth []ssh.AuthMethod, wantFP string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: auth,
		HostKeyCallback: func(_ string, _ net.Addr, key ssh.PublicKey) error {
			fp := ssh.FingerprintSHA256(key)
			if wantFP != "" && fp != wantFP {
				return fmt.Errorf("host key mismatch")
			}
			return nil
		},
		Timeout: 15 * time.Second,
	}
	return ssh.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)), config)
}
