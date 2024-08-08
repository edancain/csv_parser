package main

import (
	"os"
	"path/filepath"

	"csv_parser/csvparser"
	"csv_parser/kmlparser"
)

func main() {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	// Construct the path to the CSV file
	filePath := filepath.Join(cwd, "test_files", "ExportCSV_2018-12-09_[10-09-00].csv")

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	// Create a new CSVParser
	parser := &csvparser.CSVParser{}

	// Parse the geometry
	geometry, err := parser.ParseGeometry(file)
	if err != nil {
		return
	}

	if geometry == nil {
		return
	}

	filePath = filepath.Join(cwd, "test_files", "10.kmz")

	// Open the file
	kmlfile, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	// Create a new CSVParser
	kmlParser := &kmlparser.KMLKMZParser{}

	// Parse the geometry
	geometry, err = kmlParser.ParseGeometry(kmlfile)
	if err != nil {
		return
	}

	if geometry == nil {
		return
	}
}
