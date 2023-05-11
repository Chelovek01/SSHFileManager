package SCPpostman

import (
	"SSHFileManager/internal/archivator"
	"SSHFileManager/internal/entity/jsonModels/paketstruct"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

func SendFileWithScp(c *ssh.Client, filePath string, remotePath string) error {

	clientGoScp, err := scp.NewClientBySSH(c)
	if err != nil {
		fmt.Println("Error creating new SSH session from existing connection", err)
	}
	defer clientGoScp.Close()

	f, _ := os.Open(filePath)

	err = clientGoScp.CopyFromFile(context.Background(), *f, remotePath, "0655")
	if err != nil {
		fmt.Println("Error while copying file ", err)
	}
	f.Close()
	return err

}

// GetArchives Download needed packages with dependencis
func GetArchives(c *ssh.Client, pac paketstruct.Packet) error {

	session, err := c.NewSession()
	if err != nil {
		log.Fatal("error creating new session", err)
	}

	defer session.Close()

	session.Stdout = &bytes.Buffer{}
	session.Stderr = &bytes.Buffer{}

	cmd := fmt.Sprintf("ls %s/%s/", os.Getenv("HOME_PATH"), pac.Name)
	err = session.Run(cmd)

	if err != nil {
		log.Fatalf("%s has failed: [%w] %s",
			cmd,
			err,
			session.Stderr.(*bytes.Buffer).String(),
		)
	}

	existVersionOnServer := strings.Split(session.Stdout.(*bytes.Buffer).String(), "\n")

	version, err := versionMatches(existVersionOnServer, pac.Version)
	if err != nil {
		log.Printf("need upload packages for %s err=%s", pac.Name, err)
	}

	//Download archive
	clietnGoScpArch, err := scp.NewClientBySSH(c)
	if err != nil {
		fmt.Println("Error creating new SSH session from existing connection", err)
	}

	fileArch, _ := os.Create("./downloaded/data.zip")
	remotePathArch := fmt.Sprintf("%s/%s/%s/%s", os.Getenv("HOME_PATH"), pac.Name, version, "data.zip")

	err = clietnGoScpArch.CopyFromRemote(context.Background(), fileArch, remotePathArch)

	// Open a zip archive for reading.
	archivator.UnArchive("./extractedPackages", "./downloaded/data.zip")
	fileArch.Close()
	os.Remove("./downloaded/data.zip")

	//Download dependencies
	clietnGoScpDep, err := scp.NewClientBySSH(c)
	if err != nil {
		fmt.Println("Error creating new SSH session from existing connection", err)
	}

	fileDep, _ := os.Create("downloaded/dependencies.txt")

	remotePathDep := fmt.Sprintf("%s/%s/%s/%s", os.Getenv("HOME_PATH"), pac.Name, version, "dependencies.txt")
	err = clietnGoScpDep.CopyFromRemote(context.Background(), fileDep, remotePathDep)
	if err != nil {
		fmt.Println("dependencies does not exists for this package ")

		os.Remove("downloaded/dependencies.txt")

		clietnGoScpDep.Close()

	} else {

		clietnGoScpDep.Close()

		dep, err := os.ReadFile("downloaded/dependencies.txt")
		if err != nil {
			log.Printf("file downloaded/dependencies.txt does not exists")
		}

		var newPaket paketstruct.Packet
		err = json.Unmarshal(dep, &newPaket)
		if err != nil {
			log.Fatal(err)
		}

		os.Remove("downloaded/dependencies.txt")

		GetArchives(c, newPaket)

	}

	return err
}

// Matching needed versions
func versionMatches(pkgVersion []string, version string) (string, error) {

	fatalErr := errors.New("needed packages on server does not exist")

	if len(pkgVersion) == 0 {

		return "", fatalErr
	}

	var finalVer string

	var allowedVer []float64

	if version == "" {

		for _, val := range pkgVersion {

			if val == "" || val == " " {
				continue
			}

			flotVal, err := strconv.ParseFloat(val, 10)
			if err != nil {
				log.Fatal(err)
			}
			allowedVer = append(allowedVer, flotVal)
		}
		sort.Float64s(allowedVer)
		return fmt.Sprintf("%.2f", allowedVer[len(allowedVer)-1]), nil
	}

	op, ver := parseVersionConstraint(version)

	floatVer, err := strconv.ParseFloat(ver, 10)
	if err != nil {
		log.Fatal("got incorrect version ")
	}

	for _, val := range pkgVersion {

		if val == "" || val == " " {
			continue
		}

		floatVal, err := strconv.ParseFloat(val, 10)
		if err != nil {
			log.Fatal("got incorrect version ")
		}

		switch op {

		case ">":
			res := floatVal > floatVer
			if res == true {

				allowedVer = append(allowedVer, floatVal)
			}
		case ">=":
			res := floatVal >= floatVer
			if res == true {

				allowedVer = append(allowedVer, floatVal)
			}

		case "<":
			res := floatVal < floatVer
			if res == true {

				allowedVer = append(allowedVer, floatVal)
			}
		case "<=":
			res := floatVal <= floatVer
			if res == true {

				allowedVer = append(allowedVer, floatVal)
			}
		case "=":

			res := floatVal == floatVer
			if res == true {

				allowedVer = append(allowedVer, floatVal)
			}

		}

	}

	if len(allowedVer) > 0 {

		sort.Float64s(allowedVer)

		switch op {
		case ">":
			finalVer = fmt.Sprintf("%.2f", allowedVer[len(allowedVer)-1])
		case ">=":
			finalVer = fmt.Sprintf("%.2f", allowedVer[len(allowedVer)-1])
		case "<":
			finalVer = fmt.Sprintf("%.2f", allowedVer[len(allowedVer)-1])
		case "<=":
			finalVer = fmt.Sprintf("%.2f", allowedVer[len(allowedVer)-1])
		case "=":
			finalVer = fmt.Sprintf("%.2f", allowedVer[len(allowedVer)-1])
		}

		if version == "" {
			return fmt.Sprintf("%.2f", allowedVer[len(allowedVer)-1]), nil
		}
	} else {
		return "", fatalErr
	}

	return finalVer, nil
}

// Parsing version
func parseVersionConstraint(version string) (string, string) {
	if strings.HasPrefix(version, ">=") {
		return ">=", version[2:]
	}

	if strings.HasPrefix(version, ">") {
		return ">", version[1:]
	}

	if strings.HasPrefix(version, "<=") {
		return "<=", version[2:]
	}

	if strings.HasPrefix(version, "<") {
		return "<", version[1:]
	}

	if strings.HasPrefix(version, "=") {
		return "=", version[1:]
	}

	return "", ""
}
