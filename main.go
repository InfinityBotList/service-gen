package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"text/template"

	"github.com/go-playground/validator/v10"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

// Defines a template which is any FILENAME.yaml where FILENAME != _meta
type TemplateYaml struct {
	Command     string `yaml:"cmd" validate:"required"`         // ExecStart in systemd
	Directory   string `yaml:"dir" validate:"required"`         // WorkingDirectory in systemd
	Target      string `yaml:"target" validate:"required"`      // PartOf in systemd
	Description string `yaml:"description" validate:"required"` // Description in systemd
	After       string `yaml:"after" validate:"required"`       // After in systemd
	Broken      bool   `yaml:"broken"`                          // Does the service even work?
}

// Defines metadata which is _meta.yaml
type MetaYAML struct {
	Targets []MetaTarget `yaml:"targets" validate:"required"` // List of targets to generate
}

// Defines a target in _meta.yaml:targets
type MetaTarget struct {
	Name        string `yaml:"name" validate:"required"`        // Name of target file
	Description string `yaml:"description" validate:"required"` // Directory to place target file
}

var (
	//go:embed service.tmpl
	serviceTemplate string

	//go:embed target.tmpl
	targetTemplate string

	targetNames []string
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./service-gen <input folder>")
		os.Exit(1)
	}

	dirName := os.Args[1]

	if dirName == "" {
		fmt.Println("Usage: ./service-gen <input folder>")
		os.Exit(1)
	}

	// Read meta file
	fmt.Println("Creating metadata for services")

	inp, err := os.ReadFile(dirName + "/_meta.yaml")

	if err != nil {
		panic(err)
	}

	var meta MetaYAML

	err = yaml.Unmarshal(inp, &meta)

	if err != nil {
		panic(err)
	}

	// Validate input file
	validator := validator.New()

	err = validator.Struct(meta)

	if err != nil {
		panic(err)
	}

	// Generate target files
	for _, target := range meta.Targets {
		targetNames = append(targetNames, target.Name)

		var targetTemplate = template.Must(template.New("target").Parse(targetTemplate))

		// Output file is removal of suffix and addition of .target
		outFile := target.Name + ".target"

		if os.Getenv("OUTPUT_DIR") != "" {
			outFile = os.Getenv("OUTPUT_DIR") + "/" + outFile
		}

		// Create output file
		out, err := os.Create(outFile)

		if err != nil {
			panic(err)
		}

		err = targetTemplate.Execute(out, target)

		if err != nil {
			panic(err)
		}

		fmt.Println("Generated " + outFile)

		err = out.Close()

		if err != nil {
			panic(err)
		}
	}

	// Get dir listing of SERVICE_DIR
	dir, err := os.ReadDir(dirName)

	if err != nil {
		panic(err)
	}

	for _, file := range dir {
		if !file.IsDir() {
			if file.Name() == "_meta.yaml" {
				continue
			}

			// Generate service file by calling gen()
			fmt.Println("Generating service for " + file.Name())
			gen(dirName + "/" + file.Name())
		}
	}

	os.Exit(0)
}

func gen(inpFile string) {
	// Read input file
	inp, err := os.ReadFile(inpFile)

	if err != nil {
		panic(err)
	}

	if strings.HasSuffix(inpFile, ".service") {
		var outFile = inpFile
		// This is a service file, copy to OUTPUT_DIRECTORY directly
		if os.Getenv("OUTPUT_DIR") != "" {
			outFile = os.Getenv("OUTPUT_DIR") + "/" + outFile
		}

		// Create output file
		out, err := os.Create(outFile)

		if err != nil {
			panic(err)
		}

		_, err = out.Write(inp)

		if err != nil {
			panic(err)
		}

		fmt.Println("Copied "+inpFile, "to", outFile, "(already service)")

		err = out.Close()

		if err != nil {
			panic(err)
		}

		return
	}

	// Parse input file
	var tmpl TemplateYaml

	err = yaml.Unmarshal(inp, &tmpl)

	if err != nil {
		panic(err)
	}

	if tmpl.Broken {
		fmt.Println("Ignoring broken service:", inpFile)
		return
	}

	// Validate input file
	validator := validator.New()

	err = validator.Struct(tmpl)

	if err != nil {
		panic(err)
	}

	if strings.Contains(tmpl.Target, ".") {
		panic("Target cannot contain a period (.)")
	}

	if !slices.Contains(targetNames, tmpl.Target) {
		panic("Target " + tmpl.Target + " does not exist")
	}

	if strings.Contains(tmpl.After, ".") {
		panic("Target cannot contain a period (.)")
	}

	// Generate service file
	var serviceTemplate = template.Must(template.New("service").Parse(serviceTemplate))

	// Output file is removal of suffix and addition of .service
	outFile := strings.TrimSuffix(inpFile, ".yaml") + ".service"

	if os.Getenv("OUTPUT_DIR") != "" {
		outFile = os.Getenv("OUTPUT_DIR") + "/" + outFile
	}

	// Create output file
	out, err := os.Create(outFile)

	if err != nil {
		panic(err)
	}

	err = serviceTemplate.Execute(out, tmpl)

	if err != nil {
		panic(err)
	}

	fmt.Println("Generated " + outFile)

	err = out.Close()

	if err != nil {
		panic(err)
	}
}
