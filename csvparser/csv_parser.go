package csvparser

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/peterstace/simplefeatures/geom"
)

type CSVParser struct {
	lineString *geom.LineString
	geometry   *geom.Geometry
}

func (p *CSVParser) ParseGeometry(file io.Reader) (*geom.Geometry, error) {
	reader := bufio.NewReader(file)
	var csvContent strings.Builder
	captureCSV := false
	startMarker := "count(10HZ)"
	endMarker := "</document_content>"

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error reading file: %v", err)
		}

		if strings.Contains(line, startMarker) {
			captureCSV = true
			csvContent.WriteString(line) // Include header line
			continue
		}

		if strings.Contains(line, endMarker) {
			break
		}

		if captureCSV {
			csvContent.WriteString(line)
		}

		if err == io.EOF {
			break
		}
	}

	// Detect delimiter
	delimiter := detectDelimiter(csvContent.String())

	csvReader := csv.NewReader(strings.NewReader(csvContent.String()))
	csvReader.Comma = delimiter
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields

	// Read header
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV header: %v", err)
	}

	// Find latitude and longitude column indices
	latIndex, lonIndex := -1, -1
	for i, field := range header {
		fieldLower := strings.ToLower(field)
		switch {
		case strings.Contains(fieldLower, "lat") || fieldLower == "y":
			latIndex = i
		case strings.Contains(fieldLower, "lon") || strings.Contains(fieldLower, "lng") || fieldLower == "x":
			lonIndex = i
		}
	}

	if latIndex == -1 || lonIndex == -1 {
		return nil, fmt.Errorf("could not find latitude and longitude columns")
	}

	var coords []float64
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV record: %v", err)
		}

		if len(record) <= latIndex || len(record) <= lonIndex {
			continue // Skip rows with insufficient data
		}

		lat, err := strconv.ParseFloat(record[latIndex], 64)
		if err != nil {
			continue // Skip rows with invalid latitude
		}

		lon, err := strconv.ParseFloat(record[lonIndex], 64)
		if err != nil {
			continue // Skip rows with invalid longitude
		}

		coords = append(coords, lon, lat) 
	}

	if len(coords) < 4 { // Need at least 2 points (4 float64 values) for a valid LineString
		return nil, fmt.Errorf("not enough valid coordinates to form a LineString")
	}

	seq := geom.NewSequence(coords, geom.DimXY)
	lineString := geom.NewLineString(seq)

	p.lineString = &lineString
	geometry := lineString.AsGeometry()
	p.geometry = &geometry

	return p.geometry, nil
}

func detectDelimiter(content string) rune {
    // Potential delimiters to check
    delimiters := []rune{',', '\t', ';', '|'}
    
    lines := strings.Split(content, "\n")
    if len(lines) < 2 {
        return ',' // Default to comma if not enough lines
    }

    // Use the first few lines for detection
    sampleSize := min(5, len(lines))
    var bestDelimiter rune
    var maxConsistentColumns int

    for _, delimiter := range delimiters {
        consistentColumns := 0
        columnCount := -1

        for i := 0; i < sampleSize; i++ {
            fields := strings.Split(lines[i], string(delimiter))
            if len(fields) > 1 {
                if columnCount == -1 {
                    columnCount = len(fields)
                    consistentColumns++
                } else if len(fields) == columnCount {
                    consistentColumns++
                }
            }
        }

        if consistentColumns > maxConsistentColumns {
            maxConsistentColumns = consistentColumns
            bestDelimiter = delimiter
        }
    }

    if bestDelimiter == 0 {
        return ',' // Default to comma if no clear delimiter found
    }

    return bestDelimiter
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}


