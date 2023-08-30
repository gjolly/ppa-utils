package ppa

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp/armor"
)

func Install(aptConfigPath, ppaName, distro, keyID string) error {
	ppaURL, err := getPPAURL(ppaName)
	if err != nil {
		return fmt.Errorf("failed to get PPA url for '%v': %v", ppaName, err)
	}

	keyURL, err := getKeyURL(keyID)

	baseFileName := generateBaseFileName(ppaName)
	keyPathAbs := path.Join(aptConfigPath, "keyrings", baseFileName+".gpg")
	keyPathRel := path.Join("/etc/apt", "keyrings", baseFileName+".gpg")
	sourcePath := path.Join(aptConfigPath, "sources.list.d", baseFileName+".list")

	err = downloadKey(keyURL, keyPathAbs)
	if err != nil {
		return fmt.Errorf("failed to download key file: %v", err)
	}

	if distro == "" {
		distro, err = getDistro()
		if err != nil {
			return fmt.Errorf("failed to get current distro, use --distro: %v", err)
		}
	}

	err = writeSourceFile(ppaURL, distro, sourcePath, keyPathRel)
	if err != nil {
		return fmt.Errorf("failed to write source file: %v", err)
	}

	return nil
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
