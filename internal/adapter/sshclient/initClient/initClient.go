package initClient

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"log"
	"os"
)

func InitSSHClient() *ssh.Client {

	hostKeyCallback, err := knownhosts.New(os.Getenv("KNOWN_HOSTS"))
	if err != nil {
		log.Fatal(err)
	}

	key, err := os.ReadFile(os.Getenv("ID_RSA"))
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKeyWithPassphrase(key, []byte(os.Getenv("PASSPHRASE")))
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "chelovek_ubuntu",
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
	}

	// Connect to the remote server and perform the SSH handshake.
	client, err := ssh.Dial("tcp", "172.28.24.248:22", config)

	if err != nil {
		log.Fatalf("unable to connect: %v", err)
	}

	return client
}
