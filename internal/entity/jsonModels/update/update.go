package update

import (
	"SSHFileManager/internal/SCPpostman"
	"SSHFileManager/internal/entity/jsonModels/paketstruct"
	"golang.org/x/crypto/ssh"
)

type DownloadUpdate struct {
	Packages []paketstruct.Packet `json:"Packages"`
}

func (d *DownloadUpdate) GetArchiveFromServer(sshClient *ssh.Client) {

	for _, val := range d.Packages {

		SCPpostman.GetArchives(sshClient, val)
	}

}
