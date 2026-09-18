package servers

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/ssh"
)

// sshRunRoot runs a shell script with root privileges.
// If the SSH user is already root, the script runs as-is.
// Otherwise it uses sudo -n (passwordless) or sudo -S with sudoPassword on stdin
// (password is never placed in the remote argv).
func sshRunRoot(client *ssh.Client, script, sudoPassword string) (string, error) {
	uidOut, err := sshRun(client, "id -u")
	if err != nil {
		return uidOut, fmt.Errorf("id -u: %w", err)
	}
	if strings.TrimSpace(uidOut) == "0" {
		return sshRun(client, script)
	}

	if pass := strings.TrimSpace(sudoPassword); pass != "" {
		// Password line, then script body for bash -s.
		stdin := pass + "\n" + script
		out, err := sshRunWithStdin(client, "sudo -S -p '' -- bash -s", stdin)
		if err != nil {
			return out, fmt.Errorf("sudo failed (проверьте пароль и права sudo): %w", err)
		}
		return out, nil
	}

	out, err := sshRunWithStdin(client, "sudo -n -- bash -s", script)
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
		out, err := sshRunWithStdin(client, "sudo -S -p '' -v", pass+"\n")
		if err != nil {
			return fmt.Errorf("sudo недоступен: %s", sanitizeInstallLog(truncateLog(out, 200)))
		}
		return nil
	}
	out, err := sshRun(client, "sudo -n true")
	if err != nil {
		return fmt.Errorf("SSH-пользователь не root и нет passwordless sudo — войдите как root или укажите пароль для sudo: %s", sanitizeInstallLog(truncateLog(out, 200)))
	}
	return nil
}

func sshRunWithStdin(client *ssh.Client, cmd, stdin string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return "", err
	}
	var outBuf bytes.Buffer
	session.Stdout = &outBuf
	session.Stderr = &outBuf

	if err := session.Start(cmd); err != nil {
		return outBuf.String(), err
	}
	if _, err := io.WriteString(stdinPipe, stdin); err != nil {
		_ = session.Wait()
		return outBuf.String(), err
	}
	_ = stdinPipe.Close()
	err = session.Wait()
	return outBuf.String(), err
}
