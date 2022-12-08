package main

import (
	_ "embed"
	"fmt"
	"os"

	"text/template"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type TemplateYaml struct {
	Command     string `yaml:"cmd" validate:"required"`         // ExecStart in systemd
	Directory   string `yaml:"dir" validate:"required"`         // WorkingDirectory in systemd
	Target      string `yaml:"target" validate:"required"`      // PartOf in systemd
	Description string `yaml:"description" validate:"required"` // Description in systemd
}

//go:embed service.tmpl
var serviceTemplate string

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./service-gen <input file>")
		os.Exit(1)
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

	// Generate service file
	var serviceTemplate = template.Must(template.New("service").Parse(serviceTemplate))

	outFile := inpFile + ".service"

	out, err := os.Create(outFile)

	if err != nil {
		panic(err)
	}

	err = serviceTemplate.Execute(out, tmpl)

	if err != nil {
		panic(err)
	}
}