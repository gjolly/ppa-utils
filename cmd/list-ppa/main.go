package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
)

type PPA struct {
	Owner          string `json:"owner"`
	Name           string `json:"name"`
	Distro         string `json:"distro"`
	SourceFile     string `json:"source_file"`
	KeyFingerprint string `json:"key_fingerprint"`
	KeyringFile    string `json:"keyring_file"`
}

func (ppa PPA) URL() string {
	return fmt.Sprintf("https://ppa.launchpadcontent.net/%v/%v/ubuntu", ppa.Owner, ppa.Name)
}

func (ppa PPA) SourceLine() string {
	return fmt.Sprintf("deb [signed-by=%v] https://ppa.launchpadcontent.net/%v/%v/ubuntu %v main", ppa.KeyringFile, ppa.Owner, ppa.Name, ppa.Distro)
}

func main() {
	ppas, err := listPPAs("/etc/apt/sources.list.d")
	if err != nil {
		log.Fatal("failed to list PPAs:", err)
	}

	ppaJson, _ := json.Marshal(ppas)
	fmt.Printf("%s\n", ppaJson)

}

func listPPAs(directory string) ([]*PPA, error) {
	var PPAs []*PPA

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".list" {
			return nil
		}

		filePPAs, err := parseSourceFile(path)
		if err != nil {
			log.Panicln("failed to parse source file:", path)
			return nil
		}

		PPAs = append(PPAs, filePPAs...)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return PPAs, nil
}

func parseSourceFile(path string) ([]*PPA, error) {
	var PPAs []*PPA
	re := regexp.MustCompile(`^deb (?P<sig>\[signed-by=.*\] )http[s]{0,1}://ppa\.launchpadcontent\.net/(?P<owner>[^/]+)/(?P<name>[^/]+)/ubuntu (?P<distro>[^ ]+) main *$`)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) == 0 {
			continue
		}

		ppa := &PPA{
			SourceFile: path,
		}
		for i, name := range re.SubexpNames() {
			if name == "name" {
				ppa.Name = matches[i]
			}
			if name == "owner" {
				ppa.Owner = matches[i]
			}
			if name == "distro" {
				ppa.Distro = matches[i]
			}
			if name == "sig" {
				sig := matches[i]
				sig = strings.TrimRight(sig, "] ")
				sig = strings.TrimPrefix(sig, "[signed-by=")
				ppa.KeyringFile = sig
			}
		}

		ppa.KeyFingerprint, _ = getFingerprint(ppa.KeyringFile)

		if err := scanner.Err(); err != nil {
			return PPAs, err
		}

		PPAs = append(PPAs, ppa)
	}

	return PPAs, nil
}

func getFingerprint(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	keyring, err := openpgp.ReadKeyRing(file)
	if err != nil {
		return "", err
	}

	for _, key := range keyring {
		return hex.EncodeToString(key.PrimaryKey.Fingerprint), nil
	}

	return "", nil
}
