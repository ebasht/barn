package servers

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

// sshRunRoot runs a shell script with root privileges.
// If the SSH user is already root, the script runs as-is.
// Otherwise it uses sudo -n (passwordless) or sudo -S with sudoPassword.
func sshRunRoot(client *ssh.Client, script, sudoPassword string) (string, error) {
	uidOut, err := sshRun(client, "id -u")
	if err != nil {
		return uidOut, fmt.Errorf("id -u: %w", err)
	}
	if strings.TrimSpace(uidOut) == "0" {
		return sshRun(client, script)
	}

	quoted := strconvQuote(script)
	if pass := strings.TrimSpace(sudoPassword); pass != "" {
		// -S reads password from stdin; -p '' suppresses the prompt text in output.
		cmd := fmt.Sprintf(`printf '%%s\n' %s | sudo -S -p '' -- bash -c %s`, strconvQuote(pass), quoted)
		out, err := sshRun(client, cmd)
		if err != nil {
			return out, fmt.Errorf("sudo failed (проверьте пароль и права sudo): %w", err)
		}
		return out, nil
	}

	out, err := sshRun(client, "sudo -n -- bash -c "+quoted)
	if err != nil {
		return out, fmt.Errorf("нужен root, passwordless sudo или пароль для sudo: %w", err)
	}
	return out, nil
}

func ensureRootOrSudo(client *ssh.Client, sudoPassword string) error {
	uidOut, err := sshRun(client, "id -u")
	if err != nil {
		return fmt.Errorf("id -u: %w", err)
	}
	if strings.TrimSpace(uidOut) == "0" {
		return nil
	}
	if pass := strings.TrimSpace(sudoPassword); pass != "" {
		out, err := sshRun(client, fmt.Sprintf(`printf '%%s\n' %s | sudo -S -p '' -v`, strconvQuote(pass)))
		if err != nil {
			return fmt.Errorf("sudo недоступен: %s", truncateLog(out, 200))
		}
		return nil
	}
	out, err := sshRun(client, "sudo -n true")
	if err != nil {
		return fmt.Errorf("SSH-пользователь не root и нет passwordless sudo — войдите как root или укажите пароль для sudo: %s", truncateLog(out, 200))
	}
	return nil
}

func strconvQuote(s string) string {
	return fmt.Sprintf("%q", s)
}
