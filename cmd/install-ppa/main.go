package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp/armor"
)

type Config struct {
	PPAName       string
	APTConfigPath string
	Distro        string
	KeyID         string
}

func main() {
	config, err := parseArgs()
	if err != nil {
		log.Fatal("invalid arguments", err)
	}

	ppaURL, err := getPPAURL(config.PPAName)
	if err != nil {
		log.Fatalf("failed to get PPA url for '%v': %v", config.PPAName, err)
	}

	keyURL, err := getKeyURL(config.KeyID)

	baseFileName := generateBaseFileName(config.PPAName)
	keyPathAbs := path.Join(config.APTConfigPath, "keyrings", baseFileName+".gpg")
	keyPathRel := path.Join("/etc/apt", "keyrings", baseFileName+".gpg")
	sourcePath := path.Join(config.APTConfigPath, "sources.list.d", baseFileName+".list")

	err = downloadKey(keyURL, keyPathAbs)
	if err != nil {
		log.Fatal("failed to download key file", err)
	}

	if config.Distro == "" {
		config.Distro, err = getDistro()
		if err != nil {
			log.Fatal("failed to get current distro, use --distro", err)
		}
	}

	err = writeSourceFile(ppaURL, config.Distro, sourcePath, keyPathRel)
	if err != nil {
		log.Fatal("failed to write source file", err)
	}
}

func parseArgs() (*Config, error) {
	var config Config

	// Define command line flags
	ppaName := flag.String("ppa", "", "PPA Name")
	aptConfigPath := flag.String("apt-config", "", "APT Config Path")
	distro := flag.String("distro", "", "Distribution")
	keyID := flag.String("key-id", "", "GPG key ID")

	// Parse command line flags
	flag.Parse()

	// Check if required flags are provided
	if *ppaName == "" {
		flag.Usage()
		return nil, errors.New("PPA name required")
	}
	if *keyID == "" {
		flag.Usage()
		return nil, errors.New("Key ID required")
	}
	config.PPAName = *ppaName
	config.APTConfigPath = *aptConfigPath
	config.Distro = *distro
	config.KeyID = strings.ToLower(*keyID)

	if *aptConfigPath == "" {
		config.APTConfigPath = "/etc/apt"
	}

	return &config, nil
}

func getPPAURL(PPAName string) (string, error) {
	parts := strings.Split(PPAName, ":")
	if len(parts) != 2 || parts[0] != "ppa" {
		return "", fmt.Errorf("invalid input format")
	}

	usernameRepo := parts[1]
	//https://ppa.launchpadcontent.net/gjolly/test-ppa/ubuntu
	return fmt.Sprintf("https://ppa.launchpadcontent.net/%s/ubuntu", usernameRepo), nil
}

func getKeyURL(keyID string) (string, error) {
	return fmt.Sprintf("https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x%s", keyID), nil
}

func generateBaseFileName(ppaName string) string {
	elemts := strings.Split(ppaName, "/")

	return elemts[len(elemts)-1]
}

func downloadKey(url, destPath string) error {
	// Make a GET request to the URL to download the armored GPG key
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Check if the response status code is not OK
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download GPG key: %v [%v]", url, response.StatusCode)
	}

	// Create a new file to write the dearmored GPG key
	outputFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Read the response body and dearmor the GPG key
	armorBlock, err := armor.Decode(response.Body)
	if err != nil {
		return err
	}

	// Write the dearmored GPG key to the output file
	_, err = io.Copy(outputFile, armorBlock.Body)
	if err != nil {
		return err
	}

	return nil
}

func writeSourceFile(url, distro, destPath, keyPath string) error {
	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	content := fmt.Sprintf("deb [signed-by=%v] %v %v main", keyPath, url, distro)

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func getDistro() (string, error) {
	// Read the os-release file
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", err
	}

	// Split the content into lines
	lines := strings.Split(string(data), "\n")

	// Find and extract the VERSION_CODENAME
	for _, line := range lines {
		if strings.HasPrefix(line, "VERSION_CODENAME=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return parts[1], nil
			}
		}
	}

	return "", fmt.Errorf("VERSION_CODENAME not found in os-release")
}
