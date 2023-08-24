package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"path"

	"github.com/gjolly/install-ppa/pkg/ppa"
)

func main() {
	config, err := parseArgs()
	sourceListDir := path.Join(config.APTConfigPath, "sources.list.d")
	ppas, err := ppa.ListPPAs(sourceListDir)
	if err != nil {
		log.Fatal("failed to list PPAs:", err)
	}

	if config.OutputFormat == "json" {
		ppaJson, _ := json.Marshal(ppas)
		fmt.Printf("%s\n", ppaJson)
	}

	if config.OutputFormat == "text" {
		for _, ppa := range ppas {
			fmt.Printf("ppa:%v/%v\n", ppa.Owner, ppa.Name)
		}
	}
}

type Config struct {
	OutputFormat  string
	APTConfigPath string
}

func parseArgs() (*Config, error) {
	var config Config

	format := flag.String("format", "text", "Output format (text, json)")
	aptConfigPath := flag.String("apt-config", "", "APT Config Path")

	flag.Parse()

	config.OutputFormat = *format
	if config.OutputFormat != "text" && config.OutputFormat != "json" {
		return &config, fmt.Errorf("output format unknown: %v", config.OutputFormat)
	}

	config.APTConfigPath = *aptConfigPath
	if *aptConfigPath == "" {
		config.APTConfigPath = "/etc/apt"
	}

	return &config, nil
}
