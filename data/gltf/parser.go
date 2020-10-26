package gltf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func Parse(source Source) (*Document, error) {
	in, err := source.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open source: %w", err)
	}
	defer in.Close()

	document := new(Document)
	if err := json.NewDecoder(in).Decode(document); err != nil {
		return nil, fmt.Errorf("failed to decode json: %w", err)
	}
	for i, buffer := range document.Buffers {
		if buffer.URI != "" {
			data, err := readBuffer(source, buffer.URI)
			if err != nil {
				return nil, fmt.Errorf("failed to ")
			}
			document.Buffers[i].Data = data
		}
	}
	return document, nil
}

func readBuffer(source Source, uri string) ([]byte, error) {
	in, err := source.OpenRelative(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to open relative source %q: %w", uri, err)
	}
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read buffer data: %w", err)
	}
	return data, nil
}
