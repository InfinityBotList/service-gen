package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"text/template"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type TemplateYaml struct {
	Command     string `yaml:"cmd" validate:"required"`         // ExecStart in systemd
	Directory   string `yaml:"dir" validate:"required"`         // WorkingDirectory in systemd
	Target      string `yaml:"target" validate:"required"`      // PartOf in systemd
	Description string `yaml:"description" validate:"required"` // Description in systemd
	After       string `yaml:"after" validate:"required"`       // After in systemd
}

//go:embed service.tmpl
var serviceTemplate string

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./service-gen <input file>")
		os.Exit(1)
	}

	if os.Args[1] == "all" {
		dirName := os.Getenv("SERVICE_DIR")

		if dirName == "" {
			panic("SERVICE_DIR env var not set")
		}

		// Get dir listing of SERVICE_DIR

		dir, err := os.ReadDir(dirName)

		if err != nil {
			panic(err)
		}

		for _, file := range dir {
			if !file.IsDir() {
				// Generate service file
				fmt.Println("Generating service for " + file.Name())
				os.Args[1] = dirName + "/" + file.Name()

				main()
			}
		}

		os.Exit(0)
	}

	inpFile := os.Args[1]

	// Read input file
	inp, err := os.ReadFile(inpFile)

	if err != nil {
		panic(err)
	}

	// Parse input file
	var tmpl TemplateYaml

	err = yaml.Unmarshal(inp, &tmpl)

	if err != nil {
		panic(err)
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

	if strings.Contains(tmpl.After, ".") {
		panic("Target cannot contain a period (.)")
	}

	// Generate service file
	var serviceTemplate = template.Must(template.New("service").Parse(serviceTemplate))

	// Output file is removal of suffix and addition of .service
	outFile := strings.TrimSuffix(inpFile, ".yaml") + ".service"

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
