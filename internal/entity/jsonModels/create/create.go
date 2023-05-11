package create

import (
	"SSHFileManager/internal/SCPpostman"
	"SSHFileManager/internal/entity/jsonModels/paketstruct"
	"archive/zip"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Create struct {
	Name    string             `json:"name"`
	Version string             `json:"version"`
	Targets []interface{}      `json:"targets"`
	Packets paketstruct.Packet `json:"packets"`
}

func (c *Create) SendArchiveToServer(sshClient *ssh.Client) {

	//Getting list with needed files
	var filenames []string

	//filenames := []string{"archive_this1/gopher.txt", "archive_this1/readme.txt", "archive_this1/todo.txt"}

	for _, val := range c.Targets {
		switch i := val.(type) {
		case string:
			data, _ := filepath.Glob(i)

			for _, val := range data {

				//newVal := strings.ReplaceAll(val, "\\", "/")

				filenames = append(filenames, val)
			}

		case map[string]interface{}:
			data, _ := filepath.Glob(i["path"].(string))
			for _, val := range data {
				if strings.HasSuffix(val, i["exclude"].(string)[1:]) {
					continue
				} else {
					//newVal := strings.ReplaceAll(val, "\\", "/")
					filenames = append(filenames, val)
				}
			}
		default:
			log.Fatal("unexpected type in targets")
		}

	}

	// имя архива
	zipName := "data.zip"

	// создаем файл архива
	zipFile, err := os.Create(zipName)
	if err != nil {
		fmt.Println(err)
		return
	}
	//defer zipFile.Close()

	// создаем новый архив
	zipWriter := zip.NewWriter(zipFile)
	//defer zipWriter.Close()

	// добавляем каждый файл в архив
	for _, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Println(err)
			return
		}
		//defer file.Close()

		// создаем структуру для описания файла в архиве
		zipHeader := &zip.FileHeader{
			Name: filepath.Base(filename),
		}
		zipHeader.Method = zip.Deflate

		// записываем структуру в архив
		zipEntry, err := zipWriter.CreateHeader(zipHeader)
		if err != nil {
			fmt.Println(err)
			return
		}

		// копируем содержимое файла в архив
		_, err = io.Copy(zipEntry, file)
		if err != nil {
			fmt.Println(err)
			return
		}

		file.Close()
	}
	zipWriter.Close()
	zipFile.Close()

	//Creating folder at server
	session, err := sshClient.NewSession()
	if err != nil {
		log.Fatal("error creating new session", err)
	}
	cmd := fmt.Sprintf("mkdir -p %s/%s/%s/", os.Getenv("HOME_PATH"), c.Name, c.Version)
	defer session.Close()

	err = session.Run(cmd)
	if err != nil {
		log.Fatalf("%s has failed: [%w] %s",
			cmd,
			err,
		)
	}

	//Sending archive at server
	serverPathArch := fmt.Sprintf("%s/%s/%s/data.zip", os.Getenv("HOME_PATH"), c.Name, c.Version)
	err = SCPpostman.SendFileWithScp(sshClient, "./data.zip", serverPathArch)
	if err != nil {
		log.Fatal("sending was failed", err)
	}

	//Remove added in root files patch.tar.gz and dependencies.txt
	zipFile.Close()
	err = os.Remove("./data.zip")
	if err != nil {
		log.Fatal("remove error", err)
	}

	//Check packages dependencies
	if len(c.Packets.Name) != 0 {

		//Get dependencies from c.Packets
		packetsDep, err := json.Marshal(c.Packets)
		if err != nil {
			log.Fatalf("Error Marshal file:", err)
		}

		dependencies, err := os.Create("dependencies.txt")
		if err != nil {
			log.Fatalf("Error Creating file:", err)
		}
		dependencies.Write(packetsDep)
		defer dependencies.Close()

		//Sending dependencies at server
		serverPathDep := fmt.Sprintf("/home/chelovek_ubuntu/%s/%s/dependencies.txt", c.Name, c.Version)
		err = SCPpostman.SendFileWithScp(sshClient, "./dependencies.txt", serverPathDep)
		if err != nil {
			log.Fatal("sending was failed", err)
		}

		dependencies.Close()
		err = os.Remove("./dependencies.txt")
		if err != nil {
			log.Fatal("remove error", err)
		}

	}

}
