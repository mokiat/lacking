package game

import (
	"fmt"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game/graphics"
	asset "github.com/mokiat/lacking/game/newasset"
	"github.com/mokiat/lacking/render"
)

func (*ResourceSet) resolveWrapMode(wrap asset.WrapMode) render.WrapMode {
	switch wrap {
	case asset.WrapModeClamp:
		return render.WrapModeClamp
	case asset.WrapModeRepeat:
		return render.WrapModeRepeat
	case asset.WrapModeMirroredRepeat:
		return render.WrapModeMirroredRepeat
	default:
		panic(fmt.Errorf("unknown wrap mode: %v", wrap))
	}
}

func (*ResourceSet) resolveFiltering(filter asset.FilterMode) render.FilterMode {
	switch filter {
	case asset.FilterModeNearest:
		return render.FilterModeNearest
	case asset.FilterModeLinear:
		return render.FilterModeLinear
	case asset.FilterModeAnisotropic:
		return render.FilterModeAnisotropic
	default:
		panic(fmt.Errorf("unknown filter mode: %v", filter))
	}
}

func (*ResourceSet) resolveDataFormat(format asset.TexelFormat) render.DataFormat {
	switch format {
	case asset.TexelFormatRGBA8:
		return render.DataFormatRGBA8
	case asset.TexelFormatRGBA16F:
		return render.DataFormatRGBA16F
	case asset.TexelFormatRGBA32F:
		return render.DataFormatRGBA32F
	default:
		panic(fmt.Errorf("unknown format: %v", format))
	}
}

func (*ResourceSet) resolveCullMode(mode asset.CullMode) render.CullMode {
	switch mode {
	case asset.CullModeNone:
		return render.CullModeNone
	case asset.CullModeFront:
		return render.CullModeFront
	case asset.CullModeBack:
		return render.CullModeBack
	case asset.CullModeFrontAndBack:
		return render.CullModeFrontAndBack
	default:
		panic(fmt.Errorf("unknown cull mode: %v", mode))
	}
}

func (*ResourceSet) resolveFaceOrientation(orientation asset.FaceOrientation) render.FaceOrientation {
	switch orientation {
	case asset.FaceOrientationCCW:
		return render.FaceOrientationCCW
	case asset.FaceOrientationCW:
		return render.FaceOrientationCW
	default:
		panic(fmt.Errorf("unknown face orientation: %v", orientation))
	}
}

func (*ResourceSet) resolveComparison(comparison asset.Comparison) render.Comparison {
	switch comparison {
	case asset.ComparisonNever:
		return render.ComparisonNever
	case asset.ComparisonLess:
		return render.ComparisonLess
	case asset.ComparisonEqual:
		return render.ComparisonEqual
	case asset.ComparisonLessOrEqual:
		return render.ComparisonLessOrEqual
	case asset.ComparisonGreater:
		return render.ComparisonGreater
	case asset.ComparisonNotEqual:
		return render.ComparisonNotEqual
	case asset.ComparisonGreaterOrEqual:
		return render.ComparisonGreaterOrEqual
	case asset.ComparisonAlways:
		return render.ComparisonAlways
	default:
		panic(fmt.Errorf("unknown comparison: %v", comparison))
	}
}

func (*ResourceSet) resolveTopology(primitive asset.Topology) render.Topology {
	switch primitive {
	case asset.TopologyPoints:
		return render.TopologyPoints
	case asset.TopologyLineList:
		return render.TopologyLineList
	case asset.TopologyLineStrip:
		return render.TopologyLineStrip
	case asset.TopologyTriangleList:
		return render.TopologyTriangleList
	case asset.TopologyTriangleStrip:
		return render.TopologyTriangleStrip
	default:
		panic(fmt.Errorf("unsupported primitive: %d", primitive))
	}
}

func (s *ResourceSet) resolveVertexFormat(layout asset.VertexLayout) graphics.MeshGeometryVertexFormat {
	var result graphics.MeshGeometryVertexFormat
	if attrib := layout.Coord; attrib.BufferIndex != asset.UnspecifiedBufferIndex {
		result.Coord = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Normal; attrib.BufferIndex != asset.UnspecifiedBufferIndex {
		result.Normal = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Tangent; attrib.BufferIndex != asset.UnspecifiedBufferIndex {
		result.Tangent = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.TexCoord; attrib.BufferIndex != asset.UnspecifiedBufferIndex {
		result.TexCoord = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Color; attrib.BufferIndex != asset.UnspecifiedBufferIndex {
		result.Color = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Weights; attrib.BufferIndex != asset.UnspecifiedBufferIndex {
		result.Weights = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Joints; attrib.BufferIndex != asset.UnspecifiedBufferIndex {
		result.Joints = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	return result
}

func (*ResourceSet) resolveVertexAttributeFormat(format asset.VertexAttributeFormat) render.VertexAttributeFormat {
	switch format {
	case asset.VertexAttributeFormatRGBA32F:
		return render.VertexAttributeFormatRGBA32F
	case asset.VertexAttributeFormatRGB32F:
		return render.VertexAttributeFormatRGB32F
	case asset.VertexAttributeFormatRG32F:
		return render.VertexAttributeFormatRG32F
	case asset.VertexAttributeFormatR32F:
		return render.VertexAttributeFormatR32F

	case asset.VertexAttributeFormatRGBA16F:
		return render.VertexAttributeFormatRGBA16F
	case asset.VertexAttributeFormatRGB16F:
		return render.VertexAttributeFormatRGB16F
	case asset.VertexAttributeFormatRG16F:
		return render.VertexAttributeFormatRG16F
	case asset.VertexAttributeFormatR16F:
		return render.VertexAttributeFormatR16F

	case asset.VertexAttributeFormatRGBA16S:
		return render.VertexAttributeFormatRGBA16S
	case asset.VertexAttributeFormatRGB16S:
		return render.VertexAttributeFormatRGB16S
	case asset.VertexAttributeFormatRG16S:
		return render.VertexAttributeFormatRG16S
	case asset.VertexAttributeFormatR16S:
		return render.VertexAttributeFormatR16S

	case asset.VertexAttributeFormatRGBA16SN:
		return render.VertexAttributeFormatRGBA16SN
	case asset.VertexAttributeFormatRGB16SN:
		return render.VertexAttributeFormatRGB16SN
	case asset.VertexAttributeFormatRG16SN:
		return render.VertexAttributeFormatRG16SN
	case asset.VertexAttributeFormatR16SN:
		return render.VertexAttributeFormatR16SN

	case asset.VertexAttributeFormatRGBA16U:
		return render.VertexAttributeFormatRGBA16U
	case asset.VertexAttributeFormatRGB16U:
		return render.VertexAttributeFormatRGB16U
	case asset.VertexAttributeFormatRG16U:
		return render.VertexAttributeFormatRG16U
	case asset.VertexAttributeFormatR16U:
		return render.VertexAttributeFormatR16U

	case asset.VertexAttributeFormatRGBA16UN:
		return render.VertexAttributeFormatRGBA16UN
	case asset.VertexAttributeFormatRGB16UN:
		return render.VertexAttributeFormatRGB16UN
	case asset.VertexAttributeFormatRG16UN:
		return render.VertexAttributeFormatRG16UN
	case asset.VertexAttributeFormatR16UN:
		return render.VertexAttributeFormatR16UN

	case asset.VertexAttributeFormatRGBA8S:
		return render.VertexAttributeFormatRGBA8S
	case asset.VertexAttributeFormatRGB8S:
		return render.VertexAttributeFormatRGB8S
	case asset.VertexAttributeFormatRG8S:
		return render.VertexAttributeFormatRG8S
	case asset.VertexAttributeFormatR8S:
		return render.VertexAttributeFormatR8S

	case asset.VertexAttributeFormatRGBA8SN:
		return render.VertexAttributeFormatRGBA8SN
	case asset.VertexAttributeFormatRGB8SN:
		return render.VertexAttributeFormatRGB8SN
	case asset.VertexAttributeFormatRG8SN:
		return render.VertexAttributeFormatRG8SN
	case asset.VertexAttributeFormatR8SN:
		return render.VertexAttributeFormatR8SN

	case asset.VertexAttributeFormatRGBA8U:
		return render.VertexAttributeFormatRGBA8U
	case asset.VertexAttributeFormatRGB8U:
		return render.VertexAttributeFormatRGB8U
	case asset.VertexAttributeFormatRG8U:
		return render.VertexAttributeFormatRG8U
	case asset.VertexAttributeFormatR8U:
		return render.VertexAttributeFormatR8U

	case asset.VertexAttributeFormatRGBA8UN:
		return render.VertexAttributeFormatRGBA8UN
	case asset.VertexAttributeFormatRGB8UN:
		return render.VertexAttributeFormatRGB8UN
	case asset.VertexAttributeFormatRG8UN:
		return render.VertexAttributeFormatRG8UN
	case asset.VertexAttributeFormatR8UN:
		return render.VertexAttributeFormatR8UN

	case asset.VertexAttributeFormatRGBA8IU:
		return render.VertexAttributeFormatRGBA8IU
	case asset.VertexAttributeFormatRGB8IU:
		return render.VertexAttributeFormatRGB8IU
	case asset.VertexAttributeFormatRG8IU:
		return render.VertexAttributeFormatRG8IU
	case asset.VertexAttributeFormatR8IU:
		return render.VertexAttributeFormatR8IU

	default:
		panic(fmt.Errorf("unsupported vertex attribute format: %d", format))
	}
}

func (*ResourceSet) resolveIndexFormat(layout asset.IndexLayout) render.IndexFormat {
	switch layout {
	case asset.IndexLayoutUint16:
		return render.IndexFormatUnsignedU16
	case asset.IndexLayoutUint32:
		return render.IndexFormatUnsignedU32
	default:
		panic(fmt.Errorf("unsupported index layout: %d", layout))
	}
}
