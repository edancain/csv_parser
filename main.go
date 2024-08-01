package main

import (
	"fmt"
	"os"
	"path/filepath"

	"csv_parser/csvparser"
	"csv_parser/kmlparser"
)

func main() {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current working directory: %v\n", err)
		return
	}

	// Construct the path to the CSV file
	filePath := filepath.Join(cwd, "test_files", "ExportCSV_2018-12-09_[10-09-00].csv")

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Create a new CSVParser
	parser := &csvparser.CSVParser{}

	// Parse the geometry
	geometry, err := parser.ParseGeometry(file)
	if err != nil {
		fmt.Printf("Error parsing geometry: %v\n", err)
		return
	}

	if geometry != nil {
		fmt.Println("Success")
	}

	filePath = filepath.Join(cwd, "test_files", "10.kmz")

	// Open the file
	kmlfile, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Create a new CSVParser
	kmlParser := &kmlparser.KMLKMZParser{}

	// Parse the geometry
	geometry, err = kmlParser.ParseGeometry(kmlfile)
	if err != nil {
		fmt.Printf("Error parsing geometry: %v\n", err)
		return
	}

	if geometry != nil {
		fmt.Println("Success")
	}
}
