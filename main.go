package main

import (
	"fmt"
	"os"
	"path/filepath"

	"https://github.com/edancain/csv_parser"
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

	// Print some information about the parsed geometry
	fmt.Printf("Geometry type: %s\n", geometry.GeometryType())
	fmt.Printf("Number of points: %d\n", geometry.NumPoints())
	fmt.Printf("Bounding box: %v\n", geometry.Envelope())

	// If it's a LineString, we can get more specific information
	if ls, ok := geometry.AsLineString(); ok {
		fmt.Printf("Length: %.2f\n", ls.Length())
		startPoint := ls.StartPoint()
		endPoint := ls.EndPoint()
		fmt.Printf("Start point: (%.6f, %.6f)\n", startPoint.X(), startPoint.Y())
		fmt.Printf("End point: (%.6f, %.6f)\n", endPoint.X(), endPoint.Y())
	}
}