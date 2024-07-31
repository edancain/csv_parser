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

func (p *CSVParser) ParseGeometry(file io.Reader) (*geom.Geometry, error) {
    reader := bufio.NewReader(file)
    
    // Read the first line as header
    header, err := reader.ReadString('\n')
    if err != nil {
        return nil, fmt.Errorf("error reading header: %v", err)
    }
    header = strings.TrimSpace(header)

    // Replace multiple spaces with a single comma in the header
    header = strings.Join(strings.Fields(header), ",")

    // Split the header
    headerFields := strings.Split(header, ",")
    
    // Find latitude and longitude column indices
    latIndex, lonIndex := -1, -1
    for i, field := range headerFields {
        fieldLower := strings.ToLower(field)
        switch {
        case strings.Contains(fieldLower, "lat") || fieldLower == "y":
            latIndex = i
        case strings.Contains(fieldLower, "lon") || strings.Contains(fieldLower, "lng") || fieldLower == "x":
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

        // Replace multiple spaces with a single comma
        line = strings.Join(strings.Fields(line), ",")

        lineCount++
        fields := strings.Split(line, ",")

        if len(fields) <= latIndex || len(fields) <= lonIndex {
            fmt.Printf("Line %d: Insufficient fields. Expected at least %d, got %d\n", lineCount, max(latIndex, lonIndex)+1, len(fields))
            continue
        }

        lat, err := strconv.ParseFloat(strings.TrimSpace(fields[latIndex]), 64)
        if err != nil {
            fmt.Printf("Line %d: Error parsing latitude '%s': %v\n", lineCount, fields[latIndex], err)
            continue
        }

        lon, err := strconv.ParseFloat(strings.TrimSpace(fields[lonIndex]), 64)
        if err != nil {
            fmt.Printf("Line %d: Error parsing longitude '%s': %v\n", lineCount, fields[lonIndex], err)
            continue
        }

        coords = append(coords, lon, lat)

        if lineCount <= 5 || lineCount%1000 == 0 {
            fmt.Printf("Processed line %d: lat %v, lon %v\n", lineCount, lat, lon)
        }

        if err == io.EOF {
            break
        }
    }

    fmt.Printf("Total lines processed: %d\n", lineCount)
    fmt.Printf("Total coordinates processed: %d\n", len(coords)/2)

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