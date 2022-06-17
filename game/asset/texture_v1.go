package asset

import (
	"io"

	"github.com/mokiat/lacking/data/storage"
)

func (t *TwoDTexture) encodeV1(out io.Writer) error {
	writer := storage.NewTypedWriter(out)
	if err := writer.WriteUInt16(t.Width); err != nil {
		return err
	}
	if err := writer.WriteUInt16(t.Height); err != nil {
		return err
	}
	if err := writer.WriteUInt8(uint8(t.Wrapping)); err != nil {
		return err
	}
	if err := writer.WriteUInt8(uint8(t.Filtering)); err != nil {
		return err
	}
	if err := writer.WriteUInt8(uint8(t.Format)); err != nil {
		return err
	}
	if err := writer.WriteUInt8(uint8(t.Flags)); err != nil {
		return err
	}
	if err := writer.WriteByteBlock(t.Data); err != nil {
		return err
	}
	return nil
}

func (t *TwoDTexture) decodeV1(in io.Reader) error {
	return NewReflectDecoder(in).Decode(t)
}

func (t *CubeTexture) encodeV1(out io.Writer) error {
	writer := storage.NewTypedWriter(out)
	if err := writer.WriteUInt16(t.Dimension); err != nil {
		return err
	}
	if err := writer.WriteUInt8(uint8(t.Filtering)); err != nil {
		return err
	}
	if err := writer.WriteUInt8(uint8(t.Format)); err != nil {
		return err
	}
	if err := writer.WriteUInt8(uint8(t.Flags)); err != nil {
		return err
	}
	if err := writer.WriteByteBlock(t.FrontSide.Data); err != nil {
		return err
	}
	if err := writer.WriteByteBlock(t.BackSide.Data); err != nil {
		return err
	}
	if err := writer.WriteByteBlock(t.LeftSide.Data); err != nil {
		return err
	}
	if err := writer.WriteByteBlock(t.RightSide.Data); err != nil {
		return err
	}
	if err := writer.WriteByteBlock(t.TopSide.Data); err != nil {
		return err
	}
	if err := writer.WriteByteBlock(t.BottomSide.Data); err != nil {
		return err
	}
	return nil
}

func (t *CubeTexture) decodeV1(in io.Reader) error {
	return NewReflectDecoder(in).Decode(t)
}
