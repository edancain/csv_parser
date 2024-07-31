package csvparser

import (
	"bufio"
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

// ParseGeometry reads a CSV-like file and extracts geographic coordinates.
// It handles various delimiter types and space-separated values.
// Returns a geometry object or an error if parsing fails.
func (p *CSVParser) ParseGeometry(file io.Reader) (*geom.Geometry, error) {
    reader := bufio.NewReader(file)
    
    // Read the first line as header
    header, err := reader.ReadString('\n')
    if err != nil {
        return nil, fmt.Errorf("error reading header: %v", err)
    }
    header = strings.TrimSpace(header)

    // Detect delimiter
    delimiter := detectDelimiter([]string{header})
    fmt.Printf("Detected delimiter: %q\n", delimiter)

    var headerFields []string
    if delimiter == ' ' {
        // For space-separated values, replace multiple spaces with a single comma.
        header = strings.Join(strings.Fields(header), ",")
        headerFields = strings.Split(header, ",")
    } else {
        headerFields = strings.Split(header, string(delimiter))
    }
    
    // Find latitude and longitude column indices.
	// Unsure of the naming of these columns so test for multiple spellings.
    latIndex, lonIndex := -1, -1
    for i, field := range headerFields {
        fieldLower := strings.ToLower(field)
        switch {
        case fieldLower == "lat" || fieldLower == "latitude" || fieldLower == "y":
            latIndex = i
        case fieldLower == "lon" || fieldLower == "longitude" || fieldLower == "lng" || fieldLower == "x":
            lonIndex = i
        }
        if latIndex != -1 && lonIndex != -1 {
            break
        }
    }

    if latIndex == -1 || lonIndex == -1 {
        return nil, fmt.Errorf("could not find latitude and longitude columns")
    }

    fmt.Printf("Latitude index: %d, Longitude index: %d\n", latIndex, lonIndex)

    var coords []float64
    lineCount := 0

    // Process the rest of the file
    for {
        line, err := reader.ReadString('\n')
        if err != nil && err != io.EOF {
            return nil, fmt.Errorf("error reading line: %v", err)
        }

        line = strings.TrimSpace(line)
        if line == "" {
            if err == io.EOF {
                break
            }
            continue
        }

        var fields []string
        if delimiter == ' ' {
			// The files made with DJI-GO 4, exporting flight paths to csv show an odd
			// delimiter strategy. They are using spaces, but not the same number of spaces
			// between values. So, replace multiple spaces with a single comma and split the 
			// string so that we definitely only have fields with values.
            line = strings.Join(strings.Fields(line), ",")
            fields = strings.Split(line, ",")
        } else {
            fields = strings.Split(line, string(delimiter))
        }

        lineCount++

        if len(fields) <= latIndex || len(fields) <= lonIndex {
            //fmt.Printf("Line %d: Insufficient fields. Expected at least %d, got %d\n", lineCount, max(latIndex, lonIndex)+1, len(fields))
            continue
        }

        lat, err := strconv.ParseFloat(strings.TrimSpace(fields[latIndex]), 64)
        if err != nil {
            //fmt.Printf("Line %d: Error parsing latitude '%s': %v\n", lineCount, fields[latIndex], err)
            continue
        }

        lon, err := strconv.ParseFloat(strings.TrimSpace(fields[lonIndex]), 64)
        if err != nil {
            //fmt.Printf("Line %d: Error parsing longitude '%s': %v\n", lineCount, fields[lonIndex], err)
            continue
        }

        coords = append(coords, lon, lat)
    }

    if len(coords) < 4 {
        return nil, fmt.Errorf("not enough valid coordinates to form a LineString. Processed %d points", len(coords)/2)
    }

    seq := geom.NewSequence(coords, geom.DimXY)
    lineString := geom.NewLineString(seq)

    p.lineString = &lineString
    geometry := lineString.AsGeometry()
    p.geometry = &geometry

    return p.geometry, nil
}


func detectDelimiter(sampleLines []string) rune {
    delimiters := []rune{',', '\t', ';', '|', ' '}
    counts := make(map[rune]int)

    for _, line := range sampleLines {
        for _, d := range delimiters {
            counts[d] += strings.Count(line, string(d))
        }
    }

    bestDelimiter := ','
    maxCount := 0
    for d, count := range counts {
        if count > maxCount {
            maxCount = count
            bestDelimiter = d
        }
    }

    return bestDelimiter
}