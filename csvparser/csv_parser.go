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
}

const (
	min_linestring_values = 4
	coordinate_components = 2
	defaultDelimiter      = ','
	maxSampleLines        = 6
	minRequiredLines      = 2
	headerLineIndex       = 0
	firstDataLineIndex    = 1
)

// ParseGeometry reads a CSV-like file and extracts geographic coordinates.
// It handles various delimiter types and space-separated values.
// Returns a geometry object or an error if parsing fails.
func (p *CSVParser) ParseGeometry(r io.Reader) (*geom.Geometry, error) {
	reader := bufio.NewReader(r)

	// Read header and up to 5 data lines for delimiter detection
	sampleLines := make([]string, 0, maxSampleLines)
	for i := 0; i < maxSampleLines; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading sample line: %v", err)
		}
		sampleLines = append(sampleLines, strings.TrimSpace(line))
	}

	if len(sampleLines) < minRequiredLines {
		return nil, fmt.Errorf("not enough data to parse")
	}

	delimiter := detectDelimiter(sampleLines...)
	headerFields := splitFields(sampleLines[0], delimiter)

	latIndex, lonIndex, err := findCoordinateIndices(headerFields)
	if err != nil {
		return nil, err
	}

	coords, err := processCoordinates(reader, delimiter, latIndex, lonIndex, len(headerFields))
	if err != nil {
		return nil, err
	}

	return createGeometry(coords)
}

func splitFields(line string, delimiter rune) []string {
	if delimiter == ' ' {
		return strings.Fields(line)
	}
	return strings.Split(line, string(delimiter))
}

func findCoordinateIndices(headerFields []string) (int, int, error) {
	latIndex, lonIndex := -1, -1

	for i, field := range headerFields {
		switch strings.ToLower(field) {
		case "lat", "latitude", "y":
			latIndex = i
		case "lon", "longitude", "longtitude", "lng", "x":
			lonIndex = i
		default:
			continue
		}

		// We have located lat/lng and do not need to process any further.
		if latIndex != -1 && lonIndex != -1 {
			return latIndex, lonIndex, nil
		}
	}

	return -1, -1, fmt.Errorf("could not find latitude and longitude columns")
}

func processCoordinates(reader *bufio.Reader, delimiter rune, latIndex, lonIndex int, headerFieldCount int) ([]float64, error) {
	var coords []float64
	lineNumber := 0
	for {
		lineNumber++
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				//  Successfully parsed to the end of the file.
				break
			}

			return nil, fmt.Errorf("error reading line: %v", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := splitFields(line, delimiter)
		if len(fields) != headerFieldCount {
			// further check of bad data. If the data row elements are different to that of the header row element count, bail
			// and return an error saying as much along with the offending line.
			return nil, fmt.Errorf("data row %d has %d fields, expected %d fields", lineNumber, len(fields), headerFieldCount)
		}

		if len(fields) <= latIndex || len(fields) <= lonIndex {
			continue
		}

		lat, lon, err := parseCoordinates(fields, latIndex, lonIndex)
		if err != nil {
			continue
		}

		coords = append(coords, lon, lat)
	}
	return coords, nil
}

func parseCoordinates(fields []string, latIndex, lonIndex int) (float64, float64, error) {
	lat, err := strconv.ParseFloat(strings.TrimSpace(fields[latIndex]), 64)
	if err != nil {
		return 0, 0, err
	}

	lon, err := strconv.ParseFloat(strings.TrimSpace(fields[lonIndex]), 64)
	if err != nil {
		return 0, 0, err
	}

	return lat, lon, nil
}

func createGeometry(coords []float64) (*geom.Geometry, error) {
	if len(coords) < min_linestring_values {
		return nil, fmt.Errorf("not enough valid coordinates to form a LineString. Processed %d points", len(coords)/coordinate_components)
	}

	seq := geom.NewSequence(coords, geom.DimXY)
	lineString := geom.NewLineString(seq)
	geometry := lineString.AsGeometry()

	return &geometry, nil
}

// detectDelimiter analyzes sample lines from a CSV file to work out the delimiter used.
// It checks for common delimiters and handles space-separated values as a special case
// for DJI created files.
func detectDelimiter(sampleLines ...string) rune {
	if len(sampleLines) < minRequiredLines {
		return defaultDelimiter
	}

	header := sampleLines[headerLineIndex]

	// First, try to detect non-space delimiters
	delimiters := []rune{',', '\t', ';', '|'}
	for _, d := range delimiters {
		if strings.Contains(header, string(d)) {
			return d
		}
	}

	// If no standard delimiter is found, lets check for space-separated fields
	// as used in DJI-GO 4 app generated csv files.
	headerFields := strings.Fields(header)
	expectedFieldCount := len(headerFields)
	potentialFieldCounts := make(map[int]int)

	for _, line := range sampleLines[firstDataLineIndex:] {
		fields := strings.Fields(line)
		potentialFieldCounts[len(fields)]++
	}

	// Find the most common field count
	mostCommonFieldCount := 0
	maxOccurrences := 0
	for count, occurrences := range potentialFieldCounts {
		if occurrences > maxOccurrences {
			maxOccurrences = occurrences
			mostCommonFieldCount = count
		}
	}

	// Check if the most common field count is consistent with the header
	// or if it's slightly less (accounting for potential merged fields)
	if mostCommonFieldCount == expectedFieldCount ||
		(mostCommonFieldCount >= expectedFieldCount-2 && mostCommonFieldCount <= expectedFieldCount) {
		return ' '
	}

	// If we can't determine a reliable delimiter, return the default
	return defaultDelimiter
}
