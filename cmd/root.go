package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/gjolly/ppa-utils/pkg/ppa"
	"github.com/spf13/cobra"
)

var (
	aptConfig    string
	outputFormat string
	keyID        string
	ppaName      string
	distro       string
	dryRun       bool

	rootCmd = &cobra.Command{
		Use:   "ppa",
		Short: "CLI tool to help you manage and install PPAs",
	}
	listPPACmd = &cobra.Command{
		Use:   "list",
		Short: "list PPAs installed on the system",
		RunE:  list,
	}
	removePPACmd = &cobra.Command{
		Use:   "remove ppa:<owner>/<name>",
		Short: "remove a PPA installed on the system",
		Args:  cobra.ExactArgs(1),
		RunE:  remove,
	}
	installPPACmd = &cobra.Command{
		Use:   "install ppa:<owner>/<name>",
		Short: "install a new PPA on the system",
		Args:  cobra.ExactArgs(1),
		RunE:  install,
	}
)

func list(cmd *cobra.Command, args []string) error {
	sourceListDir := path.Join(aptConfig, "sources.list.d")
	ppas, err := ppa.List(sourceListDir)
	if err != nil {
		return err
	}

	if outputFormat == "json" {
		ppaJSON, _ := json.Marshal(ppas)
		fmt.Printf("%s\n", ppaJSON)
	}

	if outputFormat == "text" {
		for _, ppa := range ppas {
			fmt.Printf("ppa:%v/%v\n", ppa.Owner, ppa.Name)
		}
	}

	return nil
}

func install(cmd *cobra.Command, args []string) error {
	return ppa.Install(aptConfig, args[0], distro, keyID)
}

func remove(cmd *cobra.Command, args []string) error {
	return ppa.Remove(aptConfig, args[0], dryRun)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&aptConfig, "apt-config", "c", "/etc/apt", "path to the APT config")

	installPPACmd.Flags().StringVarP(&keyID, "key-id", "k", "", "Fingerprint of the key used to signed packages installed form this PPA")
	installPPACmd.Flags().StringVarP(&distro, "distro", "d", "", "Targeted distro")
	installPPACmd.MarkFlagRequired("ppa")

	listPPACmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format")

	removePPACmd.Flags().BoolVarP(&dryRun, "dryrun", "d", false, "dryrun")
	removePPACmd.MarkFlagRequired("ppa")

	rootCmd.AddCommand(installPPACmd, listPPACmd, removePPACmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
