package clientModel

import "golang.org/x/crypto/ssh"

type SSHClient struct {
	sshClient *ssh.Client
}
