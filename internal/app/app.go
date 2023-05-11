package app

import (
	"SSHFileManager/internal/adapter/sshclient/initClient"
	"SSHFileManager/internal/entity/jsonModels/create"
	"SSHFileManager/internal/entity/jsonModels/update"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func RunApp() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}

	sshclient := initClient.InitSSHClient()

	defer sshclient.Close()

	argsWithoutProg := os.Args[1:]

	for _, val := range argsWithoutProg {

		switch val {

		case "create":
			var sendArch create.Create

			jsonSendArch, _ := os.ReadFile("./packet.json")

			err = json.Unmarshal(jsonSendArch, &sendArch)
			if err != nil {
				log.Fatal("PIZDa", err)
			}
			sendArch.SendArchiveToServer(sshclient)

			fmt.Printf("packet %s, was aded on server", sendArch.Name)

		case "update":

			var downloadArch update.DownloadUpdate

			jsonDownloadARch, _ := os.ReadFile("./packages.json")

			err = json.Unmarshal(jsonDownloadARch, &downloadArch)

			downloadArch.GetArchiveFromServer(sshclient)

			fmt.Printf("packages %s, was updated", downloadArch.Packages)

		}

	}

}
