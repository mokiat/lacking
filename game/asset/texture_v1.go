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
	reader := storage.NewTypedReader(in)
	if width, err := reader.ReadUInt16(); err != nil {
		return err
	} else {
		t.Width = width
	}
	if height, err := reader.ReadUInt16(); err != nil {
		return err
	} else {
		t.Height = height
	}
	if wrapping, err := reader.ReadUInt8(); err != nil {
		return err
	} else {
		t.Wrapping = WrapMode(wrapping)
	}
	if filtering, err := reader.ReadUInt8(); err != nil {
		return err
	} else {
		t.Filtering = FilterMode(filtering)
	}
	if format, err := reader.ReadUInt8(); err != nil {
		return err
	} else {
		t.Format = TexelFormat(format)
	}
	if flags, err := reader.ReadUInt8(); err != nil {
		return err
	} else {
		t.Flags = TextureFlag(flags)
	}
	if data, err := reader.ReadBytesBlock(); err != nil {
		return err
	} else {
		t.Data = data
	}
	return nil
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
	reader := storage.NewTypedReader(in)
	if dimension, err := reader.ReadUInt16(); err != nil {
		return err
	} else {
		t.Dimension = dimension
	}
	if filtering, err := reader.ReadUInt8(); err != nil {
		return err
	} else {
		t.Filtering = FilterMode(filtering)
	}
	if format, err := reader.ReadUInt8(); err != nil {
		return err
	} else {
		t.Format = TexelFormat(format)
	}
	if flags, err := reader.ReadUInt8(); err != nil {
		return err
	} else {
		t.Flags = TextureFlag(flags)
	}
	if frontData, err := reader.ReadBytesBlock(); err != nil {
		return err
	} else {
		t.FrontSide.Data = frontData
	}
	if backData, err := reader.ReadBytesBlock(); err != nil {
		return err
	} else {
		t.BackSide.Data = backData
	}
	if leftData, err := reader.ReadBytesBlock(); err != nil {
		return err
	} else {
		t.LeftSide.Data = leftData
	}
	if rightData, err := reader.ReadBytesBlock(); err != nil {
		return err
	} else {
		t.RightSide.Data = rightData
	}
	if topData, err := reader.ReadBytesBlock(); err != nil {
		return err
	} else {
		t.TopSide.Data = topData
	}
	if bottomData, err := reader.ReadBytesBlock(); err != nil {
		return err
	} else {
		t.BottomSide.Data = bottomData
	}
	return nil
}
