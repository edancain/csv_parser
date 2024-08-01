package kmlparser

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/peterstace/simplefeatures/geom"
)

type KML struct {
	XMLName  xml.Name `xml:"kml"`
	Document Document `xml:"Document"`
}

type Document struct {
	Folder Folder `xml:"Folder"`
}

type Folder struct {
	Placemarks []Placemark `xml:"Placemark"`
}

type Placemark struct {
	Name        string     `xml:"name"`
	Description string     `xml:"description"`
	LineString  LineString `xml:"LineString"`
}

type LineString struct {
	Coordinates string `xml:"coordinates"`
}

type KMLKMZParser struct {
	lineString *geom.LineString
	geometry   *geom.Geometry
}

const (
	CoordinateComponents = 2
)

func (p *KMLKMZParser) ParseGeometry(file io.Reader) (*geom.Geometry, error) {
	// Read the entire file content
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Check if it's a KMZ file
	if isKMZ(data) {
		data, err = extractKMLFromKMZ(data)
		if err != nil {
			return nil, err
		}
	}

	// Parse the KML data
	var kml KML
	err = xml.Unmarshal(data, &kml)
	if err != nil {
		return nil, err
	}

	// Create a slice to hold all coordinates
	var allcoords []float64
	// Extract and concatenate LineStrings
	for _, placemark := range kml.Document.Folder.Placemarks {
		if strings.Contains(placemark.Name, "Flight Mode") {
			coords := parseCoordinates(placemark.LineString.Coordinates)
			allcoords = append(allcoords, coords...)
		}
	}

	// Create a LineString from all coordinates
	seq := geom.NewSequence(allcoords, geom.DimXY)
	lineString := geom.NewLineString(seq)

	p.lineString = &lineString
	geometry := lineString.AsGeometry()
	p.geometry = &geometry
	return &geometry, nil
}

func isKMZ(data []byte) bool {
	return len(data) > 2 && data[0] == 0x50 && data[1] == 0x4B
}

func parseCoordinates(coordStr string) []float64 {
	var coords []float64
	points := strings.Fields(coordStr)
	for _, point := range points {
		parts := strings.Split(point, ",")
		if len(parts) >= CoordinateComponents {
			lon, _ := strconv.ParseFloat(parts[0], 64)
			lat, _ := strconv.ParseFloat(parts[1], 64)
			coords = append(coords, lon, lat)
		}
	}
	return coords
}

func extractKMLFromKMZ(data []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	for _, file := range reader.File {
		if filepath.Ext(file.Name) == ".kml" {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			return io.ReadAll(rc)
		}
	}

	return nil, errors.New("no KML file found in KMZ archive")
}
