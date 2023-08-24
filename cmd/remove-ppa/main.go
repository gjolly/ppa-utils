package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/gjolly/install-ppa/pkg/ppa"
)

func main() {
	config, err := parseArgs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = removePPA(config.PPAName, config.DryRun, config.APTConfigPath)
	if err != nil {
		log.Fatal("failed to remove PPAs:", err)
	}
}

func removePPA(ppaShort string, dryrun bool, aptConfPath string) error {
	path.Join(aptConfPath, "sources.list.d")
	ppas, err := ppa.ListPPAs(aptConfPath)
	if err != nil {
		return err
	}

	type fileInfo struct {
		Matches  int
		ToDelete bool
	}
	sourceFiles := make(map[string]*fileInfo, 0)
	keyringFiles := make(map[string]*fileInfo, 0)
	for _, ppa := range ppas {
		if info, ok := sourceFiles[ppa.SourceFile]; !ok {
			sourceFiles[ppa.SourceFile] = &fileInfo{
				Matches:  1,
				ToDelete: false,
			}
		} else {
			info.Matches++
		}

		keyringFilePath := path.Join(aptConfPath, "keyrings", path.Base(ppa.KeyringFile))
		if info, ok := keyringFiles[keyringFilePath]; !ok {
			keyringFiles[keyringFilePath] = &fileInfo{
				Matches:  1,
				ToDelete: false,
			}
		} else {
			info.Matches++
		}

		if ppa.Short() == ppaShort {
			sourceInfo := sourceFiles[ppa.SourceFile]
			sourceInfo.ToDelete = true

			keyringInfo := keyringFiles[keyringFilePath]
			keyringInfo.ToDelete = true
		}
	}

	filesToRemove := make([]string, 0)
	sourceFilesToEdit := make([]string, 0)

	for filename, file := range sourceFiles {
		if file.ToDelete && file.Matches == 1 {
			filesToRemove = append(filesToRemove, filename)
		}
		if file.ToDelete && file.Matches > 1 {
			sourceFilesToEdit = append(sourceFilesToEdit, filename)
		}
	}

	for filename, file := range keyringFiles {
		if file.ToDelete && file.Matches == 1 {
			filesToRemove = append(filesToRemove, filename)
		}
		if file.ToDelete && file.Matches > 1 {
			fmt.Fprintf(os.Stderr, "not removing keyring because it's used by another PPA: %v\n", filename)
		}
	}

	PPA, err := ppa.NewFromShort(ppaShort)
	if err != nil {
		return err
	}

	for _, file := range sourceFilesToEdit {
		err = editSourceFile(file, PPA, dryrun)
	}

	err = deleteFiles(filesToRemove, dryrun)
	if err != nil {
		return err
	}

	return nil
}

func editSourceFile(path string, PPA *ppa.PPA, dryrun bool) error {
	if dryrun {
		fmt.Fprintf(os.Stderr, "[dryrun] removing PPA line from file: %v\n", path)
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpfile, err := os.CreateTemp(os.TempDir(), "ppa")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, PPA.URL()) {
			continue
		}

		tmpfile.Write([]byte(line))
	}

	tmpfile.Close()
	os.Rename(tmpfile.Name(), path)

	return nil
}

func deleteFiles(files []string, dryrun bool) error {
	for _, file := range files {
		if dryrun {
			fmt.Fprintf(os.Stderr, "[dryrun] deleting %v\n", file)
			continue
		}

		os.Remove(file)
	}
	return nil
}

type Config struct {
	PPAName       string
	DryRun        bool
	APTConfigPath string
}

func parseArgs() (*Config, error) {
	var config Config

	name := flag.String("ppa", "", "PPA to remove")
	dryrun := flag.Bool("dryrun", false, "PPA to remove")
	aptConfigPath := flag.String("apt-config", "", "APT Config Path")

	flag.Parse()

	config.PPAName = *name
	if config.PPAName == "" {
		return &config, errors.New("-ppa cannot be empty")
	}
	config.DryRun = *dryrun

	config.APTConfigPath = *aptConfigPath
	if *aptConfigPath == "" {
		config.APTConfigPath = "/etc/apt"
	}

	return &config, nil
}
