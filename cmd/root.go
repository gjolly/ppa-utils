package cmd

import (
	"encoding/json"
	"fmt"
	"log"
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
		Run:   list,
	}
	removePPACmd = &cobra.Command{
		Use:   "remove",
		Short: "remove a PPA installed on the system",
		Run:   remove,
	}
	installPPACmd = &cobra.Command{
		Use:   "install",
		Short: "install a new PPA on the system",
		Run:   install,
	}
)

func list(cmd *cobra.Command, args []string) {
	sourceListDir := path.Join(aptConfig, "sources.list.d")
	ppas, err := ppa.List(sourceListDir)
	if err != nil {
		log.Fatal("failed to list PPAs:", err)
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
}

func install(cmd *cobra.Command, args []string) {
	ppa.Install(aptConfig, ppaName, distro, keyID)
}

func remove(cmd *cobra.Command, args []string) {
	ppa.Remove(aptConfig, ppaName, dryRun)
}

func init() {
	rootCmd.Flags().StringP("apt-config", "c", "/etc/apt", "path to the APT config")

	installPPACmd.Flags().StringVarP(&ppaName, "ppa", "p", "", "PPA name formated as followed: ppa:<owner>/<name>")
	installPPACmd.Flags().StringVarP(&keyID, "key-id", "k", "", "Fingerprint of the key used to signed packages installed form this PPA")
	installPPACmd.Flags().StringVarP(&distro, "distro", "d", "", "Targeted distro")
	installPPACmd.MarkFlagRequired("ppa")

	listPPACmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format")

	removePPACmd.Flags().StringVarP(&ppaName, "ppa", "p", "", "PPA name formated as followed: ppa:<owner>/<name>")
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
