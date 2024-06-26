package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"io"

	"most-active-github-users-counter/output"
	"most-active-github-users-counter/top"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var locations arrayFlags
var excludeLocations arrayFlags
var presetTitle string
var presetChecksum string

func main() {
	token := flag.String("token", LookupEnvOrString("GITHUB_TOKEN", ""), "Github auth token")
	amount := flag.Int("amount", 256, "Amount of users to show")
	considerNum := flag.Int("consider", 1000, "Amount of users to consider")
	outputOpt := flag.String("output", "plain", "Output format: plain, csv, json") // Include "json" as an option
	fileName := flag.String("file", "", "Output file (optional, defaults to stdout)")
	presetName := flag.String("preset", "", "Preset (optional)")
	listPresets := flag.Bool("list-presets", false, "List all available presets as CSV and exit immediately")

	flag.Var(&locations, "location", "Location to query")
	flag.Parse()

	if *listPresets {
		fmt.Println("preset,title,definition_checksum")
		for name, _ := range PRESETS {
			fmt.Printf("%v,\"%v\",%v\n", name, PresetTitle(name), PresetChecksum(name))
		}
		return
	}

	if *presetName != "" {
		preset := Preset(*presetName)
		locations = preset.include
		excludeLocations = preset.exclude
		presetTitle = PresetTitle(*presetName)
		presetChecksum = PresetChecksum(*presetName)
	}

	var format output.Format

	if *outputOpt == "plain" {
		format = output.PlainOutput
	} else if *outputOpt == "yaml" {
		format = output.YamlOutput
	} else if *outputOpt == "csv" {
		format = output.CsvOutput
	} else if *outputOpt == "json" {
		format = output.JsonOutput
	} else {
		log.Fatal("Unrecognized output format: ", *outputOpt)
	}

	opts := top.Options{Token: *token, Locations: locations, ExcludeLocations: excludeLocations, Amount: *amount, ConsiderNum: *considerNum, PresetTitle: presetTitle, PresetChecksum: presetChecksum}
	data, err := top.GithubTop(opts)

	if err != nil {
		log.Fatal(err)
	}

	var writer io.Writer
	if *fileName != "" {
		if *outputOpt == "json" {
			// If the output format is JSON, append .json to the filename.
			*fileName += ".json"
		}
		f, err := os.Create(*fileName)
		if err != nil {
			log.Fatal(err)
		}
		writer = bufio.NewWriter(f)
		defer f.Close()
	} else {
		writer = bufio.NewWriter(os.Stdout)
	}

	err = format(data, writer, opts)
	if err != nil {
		log.Fatal(err)
	}
	if bufferedWriter, ok := writer.(*bufio.Writer); ok {
		bufferedWriter.Flush()
	}	
}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
