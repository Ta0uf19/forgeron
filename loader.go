package forgeron

import (
	"archive/zip"
	"bytes"
	"embed"
	"fmt"
	"io"
)

//go:embed data_points/*.json data_points/*.zip
var dataFiles embed.FS

// loadNetworkFromZip loads a Bayesian network from a zip file in the data_points directory
func loadNetworkFromZip(filename string) (*bayesianNetwork, error) {
	zipData, err := dataFiles.ReadFile("data_points/" + filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %v", filename, err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create zip reader: %v", err)
	}

	if len(zipReader.File) == 0 {
		return nil, fmt.Errorf("no files found in zip")
	}

	file, err := zipReader.File[0].Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file in zip: %v", err)
	}
	defer file.Close()

	networkData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file contents: %v", err)
	}

	network := newBayesianNetwork()
	if err := network.loadNetwork(networkData); err != nil {
		return nil, fmt.Errorf("failed to load network: %v", err)
	}

	return network, nil
}
