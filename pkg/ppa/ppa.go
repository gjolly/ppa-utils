package ppa

import (
	"fmt"
	"strings"
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
	if ppa.KeyringFile != "" {
		return fmt.Sprintf("deb [signed-by=%v] https://ppa.launchpadcontent.net/%v/%v/ubuntu %v main", ppa.KeyringFile, ppa.Owner, ppa.Name, ppa.Distro)
	}

	return fmt.Sprintf("deb https://ppa.launchpadcontent.net/%v/%v/ubuntu %v main", ppa.Owner, ppa.Name, ppa.Distro)
}

func (ppa PPA) Short() string {
	return fmt.Sprintf("ppa:%v/%v", ppa.Owner, ppa.Name)
}

func NewFromShort(shortHandle string) (*PPA, error) {
	parts := strings.Split(shortHandle, ":")
	if len(parts) != 2 || parts[0] != "ppa" {
		return nil, fmt.Errorf("invalid input format")
	}

	info := strings.Split(parts[1], "/")

	if len(info) != 2 {
		return nil, fmt.Errorf("invalid input format")
	}

	return &PPA{
		Owner: info[0],
		Name:  info[1],
	}, nil
}
