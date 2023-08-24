package ppa

import "fmt"

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
