package ppa

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

func Remove(aptConfigPath, ppaShort string, dryRun bool) error {
	path.Join(aptConfigPath, "sources.list.d")
	ppas, err := List(aptConfigPath)
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

		keyringFilePath := path.Join(aptConfigPath, "keyrings", path.Base(ppa.KeyringFile))
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

	PPA, err := NewFromShort(ppaShort)
	if err != nil {
		return err
	}

	for _, file := range sourceFilesToEdit {
		err = editSourceFile(file, PPA, dryRun)
	}

	err = deleteFiles(filesToRemove, dryRun)
	if err != nil {
		return err
	}

	return nil
}

func editSourceFile(path string, PPA *PPA, dryrun bool) error {
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
