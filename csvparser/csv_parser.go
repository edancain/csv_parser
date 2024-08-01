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

const (
	MinCoordinatesForLineString = 4
	CoordinateComponents        = 2
)

// ParseGeometry reads a CSV-like file and extracts geographic coordinates.
// It handles various delimiter types and space-separated values.
// Returns a geometry object or an error if parsing fails.
func (p *CSVParser) ParseGeometry(file io.Reader) (*geom.Geometry, error) {
	reader := bufio.NewReader(file)

	header, err := p.readHeader(reader)
	if err != nil {
		return nil, err
	}

	delimiter := detectDelimiter([]string{header})
	headerFields := p.splitFields(header, delimiter)

	latIndex, lonIndex, err := p.findCoordinateIndices(headerFields)
	if err != nil {
		return nil, err
	}

	coords, err := p.processCoordinates(reader, delimiter, latIndex, lonIndex)
	if err != nil {
		return nil, err
	}

	return p.createGeometry(coords)
}

func (p *CSVParser) readHeader(reader *bufio.Reader) (string, error) {
	header, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading header: %v", err)
	}
	return strings.TrimSpace(header), nil
}

func (p *CSVParser) splitFields(line string, delimiter rune) []string {
	if delimiter == ' ' {
		return strings.Fields(line)
	}
	return strings.Split(line, string(delimiter))
}

func (p *CSVParser) findCoordinateIndices(headerFields []string) (int, int, error) {
	latIndex, lonIndex := -1, -1
	for i, field := range headerFields {
		fieldLower := strings.ToLower(field)
		switch {
		case fieldLower == "lat" || fieldLower == "latitude" || fieldLower == "y":
			latIndex = i
		case fieldLower == "lon" || fieldLower == "longitude" || fieldLower == "longtitude" || fieldLower == "lng" || fieldLower == "x":
			lonIndex = i
		}
		if latIndex != -1 && lonIndex != -1 {
			return latIndex, lonIndex, nil
		}
	}
	return -1, -1, fmt.Errorf("could not find latitude and longitude columns")
}

func (p *CSVParser) processCoordinates(reader *bufio.Reader, delimiter rune, latIndex, lonIndex int) ([]float64, error) {
	var coords []float64
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

		fields := p.splitFields(line, delimiter)
		if len(fields) <= latIndex || len(fields) <= lonIndex {
			continue
		}

		lat, lon, err := p.parseCoordinates(fields, latIndex, lonIndex)
		if err != nil {
			continue
		}

		coords = append(coords, lon, lat)
	}
	return coords, nil
}

func (p *CSVParser) parseCoordinates(fields []string, latIndex, lonIndex int) (float64, float64, error) {
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

func (p *CSVParser) createGeometry(coords []float64) (*geom.Geometry, error) {
	if len(coords) < MinCoordinatesForLineString {
		return nil, fmt.Errorf("not enough valid coordinates to form a LineString. Processed %d points", len(coords)/CoordinateComponents)
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
