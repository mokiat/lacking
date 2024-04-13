package json

import (
	"encoding/json"
	"fmt"
	"io"
)

type Level struct {
	StaticEntities []Entity `json:"static_entities"`
}

type Entity struct {
	Model  string      `json:"model"`
	Matrix [16]float32 `json:"matrix"`
}

func NewLevelDecoder() *LevelDecoder {
	return &LevelDecoder{}
}

type LevelDecoder struct{}

func (d *LevelDecoder) Decode(in io.Reader) (*Level, error) {
	var level Level
	if err := json.NewDecoder(in).Decode(&level); err != nil {
		return nil, fmt.Errorf("failed to decode json: %w", err)
	}
	return &level, nil
}

func NewLevelEncoder() *LevelEncoder {
	return &LevelEncoder{}
}

type LevelEncoder struct{}

func (e *LevelEncoder) Encode(out io.Writer, level *Level) error {
	if err := json.NewEncoder(out).Encode(level); err != nil {
		return fmt.Errorf("failed to encode json: %w", err)
	}
	return nil
}
