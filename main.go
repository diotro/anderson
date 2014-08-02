package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fraenkel/candiedyaml"
	"github.com/mitchellh/colorstring"
	"github.com/ryanuber/go-license"
)

type Config struct {
	Whitelist  []string `yaml:"whitelist"`
	Greylist   []string `yaml:"greylist"`
	Blacklist  []string `yaml:"blacklist"`
	Exceptions []string `yaml:"exceptions"`
}

type Godeps struct {
	Deps []Dependency
}

type Dependency struct {
	ImportPath string
}

func main() {
	license.DefaultLicenseFiles = []string{
		"LICENSE", "LICENSE.txt", "LICENSE.md", "license.txt",
		"COPYING", "COPYING.txt", "COPYING.md", "copying.txt",
		"MIT.LICENSE",
	}

	say("[blue]> Hold still citizen, scanning dependencies for contraband...")

	configFile, err := os.Open(".anderson.yml")
	if err != nil {
		fatal("You seem to be missing your .anderson.yml...")
	}

	var config Config
	if err := candiedyaml.NewDecoder(configFile).Decode(&config); err != nil {
		panic(err)
	}

	godepsFile, err := os.Open("Godeps/Godeps.json")
	if err != nil {
		fatal("Couldn't find your Godeps.json file!")
	}

	var godep Godeps
	if err := json.NewDecoder(godepsFile).Decode(&godep); err != nil {
		fatal("Your Godeps file wasn't valid JSON!")
	}

	failed := false

	for _, dependency := range godep.Deps {
		importPath := dependency.ImportPath
		path, err := LookGopath(importPath)
		if err != nil {
			fatal(fmt.Sprintf("Could not find %s in your GOPATH...", importPath))
		}

		l, err := license.NewFromDir(path)
		whitespace := strings.Repeat(" ", 80-10-len(importPath))
		if err != nil {
			if err.Error() == "license: unable to find any license file" {
				say(fmt.Sprintf("[white]%s%s[magenta]NO LICENSE", importPath, whitespace))
				failed = true
			} else if err.Error() == "license: could not guess license type" {
				say(fmt.Sprintf("[white]%s%s   [cyan]UNKNOWN", importPath, whitespace))
			} else {
				panic(err)
			}
			failed = true

			continue
		}

		if contains(config.Blacklist, l.Type) {
			say(fmt.Sprintf("[white]%s%s[red]CONTRABAND", importPath, whitespace))
			failed = true
			continue
		}

		if contains(config.Whitelist, l.Type) {
			say(fmt.Sprintf("[white]%s%s[green]CHECKS OUT", importPath, whitespace))
			continue
		}

		if contains(config.Greylist, l.Type) {
			if contains(config.Exceptions, importPath) {
				say(fmt.Sprintf("[white]%s%s[green]CHECKS OUT", importPath, whitespace))
			} else {
				say(fmt.Sprintf("[white]%s%s[yellow]BORDERLINE", importPath, whitespace))
				failed = true
			}
			continue
		}
	}

	if failed {
		os.Exit(1)
	}
}

func fatal(message string) {
	say(fmt.Sprintf("[red]> %s", message))
	os.Exit(1)
}

func say(message string) {
	fmt.Println(colorstring.Color(message))
}

func contains(haystack []string, needle string) bool {
	for _, element := range haystack {
		if element == needle {
			return true
		}
	}
	return false
}
