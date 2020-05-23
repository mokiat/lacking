package asset

import (
	"encoding/gob"
	"fmt"
	"io"
)

type TwoDTexture struct {
	Width  uint16
	Height uint16
	Data   []byte
}

type TextureSide int

const (
	TextureSideFront TextureSide = iota
	TextureSideBack
	TextureSideLeft
	TextureSideRight
	TextureSideTop
	TextureSideBottom
)

type CubeTexture struct {
	Dimension uint16
	Sides     [6]CubeTextureSide
}

type CubeTextureSide struct {
	Data []byte
}

func EncodeTwoDTexture(out io.Writer, texture *TwoDTexture) error {
	return WriteCompressed(out, func(compOut io.Writer) error {
		if err := gob.NewEncoder(compOut).Encode(texture); err != nil {
			return fmt.Errorf("failed to encode gob stream: %w", err)
		}
		return nil
	})
}

func DecodeTwoDTexture(in io.Reader, texture *TwoDTexture) error {
	return ReadCompressed(in, func(compIn io.Reader) error {
		if err := gob.NewDecoder(compIn).Decode(texture); err != nil {
			return fmt.Errorf("failed to decode gob stream: %w", err)
		}
		return nil
	})
}

func EncodeCubeTexture(out io.Writer, texture *CubeTexture) error {
	return WriteCompressed(out, func(compOut io.Writer) error {
		if err := gob.NewEncoder(compOut).Encode(texture); err != nil {
			return fmt.Errorf("failed to encode gob stream: %w", err)
		}
		return nil
	})
}

func DecodeCubeTexture(in io.Reader, texture *CubeTexture) error {
	return ReadCompressed(in, func(compIn io.Reader) error {
		if err := gob.NewDecoder(compIn).Decode(texture); err != nil {
			return fmt.Errorf("failed to decode gob stream: %w", err)
		}
		return nil
	})
}
