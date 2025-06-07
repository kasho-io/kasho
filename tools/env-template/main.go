package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// EnvMap converts environment variables to a map[string]interface{}
func EnvMap() map[string]interface{} {
	envMap := make(map[string]interface{})
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		envMap[pair[0]] = pair[1]
	}
	return envMap
}

func processTemplate(templatePath string, envMap map[string]interface{}) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, envMap); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	// Write to output file (remove .template extension)
	outputPath := templatePath[:len(templatePath)-9] // Remove .template
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("error writing output file: %w", err)
	}

	fmt.Printf("Processed %s -> %s\n", templatePath, outputPath)
	return nil
}

func main() {
	// Define command line flags
	templateDirs := flag.String("dirs", "", "Comma-separated list of directories to process")
	flag.Parse()

	if *templateDirs == "" {
		fmt.Println("Error: --dirs flag is required")
		os.Exit(1)
	}

	// Get environment variables
	envMap := EnvMap()

	// Process each directory
	for _, dir := range strings.Split(*templateDirs, ",") {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}

		var templateFiles []string
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".template") {
				templateFiles = append(templateFiles, path)
			}
			return nil
		})

		if err != nil {
			fmt.Printf("Error walking directory %s: %v\n", dir, err)
			continue
		}

		if len(templateFiles) == 0 {
			fmt.Printf("No template files found in %s\n", dir)
			continue
		}

		// Process each template file
		for _, file := range templateFiles {
			if err := processTemplate(file, envMap); err != nil {
				fmt.Printf("Error processing %s: %v\n", file, err)
			}
		}
	}
}
