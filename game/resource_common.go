package game

import (
	"fmt"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/render"
)

func (*ResourceSet) resolveWrapMode(wrap dto.WrapMode) render.WrapMode {
	switch wrap {
	case dto.WrapModeClamp:
		return render.WrapModeClamp
	case dto.WrapModeRepeat:
		return render.WrapModeRepeat
	case dto.WrapModeMirroredRepeat:
		return render.WrapModeMirroredRepeat
	default:
		panic(fmt.Errorf("unknown wrap mode: %v", wrap))
	}
}

func (*ResourceSet) resolveFiltering(filter dto.FilterMode) render.FilterMode {
	switch filter {
	case dto.FilterModeNearest:
		return render.FilterModeNearest
	case dto.FilterModeLinear:
		return render.FilterModeLinear
	case dto.FilterModeAnisotropic:
		return render.FilterModeAnisotropic
	default:
		panic(fmt.Errorf("unknown filter mode: %v", filter))
	}
}

func (*ResourceSet) resolveDataFormat(format dto.TexelFormat) render.DataFormat {
	switch format {
	case dto.TexelFormatRGBA8:
		return render.DataFormatRGBA8
	case dto.TexelFormatRGBA16F:
		return render.DataFormatRGBA16F
	case dto.TexelFormatRGBA32F:
		return render.DataFormatRGBA32F
	default:
		panic(fmt.Errorf("unknown format: %v", format))
	}
}

func (*ResourceSet) resolveCullMode(mode dto.CullMode) render.CullMode {
	switch mode {
	case dto.CullModeNone:
		return render.CullModeNone
	case dto.CullModeFront:
		return render.CullModeFront
	case dto.CullModeBack:
		return render.CullModeBack
	case dto.CullModeFrontAndBack:
		return render.CullModeFrontAndBack
	default:
		panic(fmt.Errorf("unknown cull mode: %v", mode))
	}
}

func (*ResourceSet) resolveFaceOrientation(orientation dto.FaceOrientation) render.FaceOrientation {
	switch orientation {
	case dto.FaceOrientationCCW:
		return render.FaceOrientationCCW
	case dto.FaceOrientationCW:
		return render.FaceOrientationCW
	default:
		panic(fmt.Errorf("unknown face orientation: %v", orientation))
	}
}

func (*ResourceSet) resolveComparison(comparison dto.Comparison) render.Comparison {
	switch comparison {
	case dto.ComparisonNever:
		return render.ComparisonNever
	case dto.ComparisonLess:
		return render.ComparisonLess
	case dto.ComparisonEqual:
		return render.ComparisonEqual
	case dto.ComparisonLessOrEqual:
		return render.ComparisonLessOrEqual
	case dto.ComparisonGreater:
		return render.ComparisonGreater
	case dto.ComparisonNotEqual:
		return render.ComparisonNotEqual
	case dto.ComparisonGreaterOrEqual:
		return render.ComparisonGreaterOrEqual
	case dto.ComparisonAlways:
		return render.ComparisonAlways
	default:
		panic(fmt.Errorf("unknown comparison: %v", comparison))
	}
}

func (*ResourceSet) resolveTopology(primitive dto.Topology) render.Topology {
	switch primitive {
	case dto.TopologyPoints:
		return render.TopologyPoints
	case dto.TopologyLineList:
		return render.TopologyLineList
	case dto.TopologyLineStrip:
		return render.TopologyLineStrip
	case dto.TopologyTriangleList:
		return render.TopologyTriangleList
	case dto.TopologyTriangleStrip:
		return render.TopologyTriangleStrip
	default:
		panic(fmt.Errorf("unsupported primitive: %d", primitive))
	}
}

func (s *ResourceSet) resolveVertexFormat(layout dto.VertexLayout) graphics.MeshGeometryVertexFormat {
	var result graphics.MeshGeometryVertexFormat
	if attrib := layout.Coord; attrib.BufferIndex != dto.UnspecifiedBufferIndex {
		result.Coord = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Normal; attrib.BufferIndex != dto.UnspecifiedBufferIndex {
		result.Normal = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Tangent; attrib.BufferIndex != dto.UnspecifiedBufferIndex {
		result.Tangent = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.TexCoord; attrib.BufferIndex != dto.UnspecifiedBufferIndex {
		result.TexCoord = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Color; attrib.BufferIndex != dto.UnspecifiedBufferIndex {
		result.Color = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Weights; attrib.BufferIndex != dto.UnspecifiedBufferIndex {
		result.Weights = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Joints; attrib.BufferIndex != dto.UnspecifiedBufferIndex {
		result.Joints = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	return result
}

func (*ResourceSet) resolveVertexAttributeFormat(format dto.VertexAttributeFormat) render.VertexAttributeFormat {
	switch format {
	case dto.VertexAttributeFormatRGBA32F:
		return render.VertexAttributeFormatRGBA32F
	case dto.VertexAttributeFormatRGB32F:
		return render.VertexAttributeFormatRGB32F
	case dto.VertexAttributeFormatRG32F:
		return render.VertexAttributeFormatRG32F
	case dto.VertexAttributeFormatR32F:
		return render.VertexAttributeFormatR32F

	case dto.VertexAttributeFormatRGBA16F:
		return render.VertexAttributeFormatRGBA16F
	case dto.VertexAttributeFormatRGB16F:
		return render.VertexAttributeFormatRGB16F
	case dto.VertexAttributeFormatRG16F:
		return render.VertexAttributeFormatRG16F
	case dto.VertexAttributeFormatR16F:
		return render.VertexAttributeFormatR16F

	case dto.VertexAttributeFormatRGBA16S:
		return render.VertexAttributeFormatRGBA16S
	case dto.VertexAttributeFormatRGB16S:
		return render.VertexAttributeFormatRGB16S
	case dto.VertexAttributeFormatRG16S:
		return render.VertexAttributeFormatRG16S
	case dto.VertexAttributeFormatR16S:
		return render.VertexAttributeFormatR16S

	case dto.VertexAttributeFormatRGBA16SN:
		return render.VertexAttributeFormatRGBA16SN
	case dto.VertexAttributeFormatRGB16SN:
		return render.VertexAttributeFormatRGB16SN
	case dto.VertexAttributeFormatRG16SN:
		return render.VertexAttributeFormatRG16SN
	case dto.VertexAttributeFormatR16SN:
		return render.VertexAttributeFormatR16SN

	case dto.VertexAttributeFormatRGBA16U:
		return render.VertexAttributeFormatRGBA16U
	case dto.VertexAttributeFormatRGB16U:
		return render.VertexAttributeFormatRGB16U
	case dto.VertexAttributeFormatRG16U:
		return render.VertexAttributeFormatRG16U
	case dto.VertexAttributeFormatR16U:
		return render.VertexAttributeFormatR16U

	case dto.VertexAttributeFormatRGBA16UN:
		return render.VertexAttributeFormatRGBA16UN
	case dto.VertexAttributeFormatRGB16UN:
		return render.VertexAttributeFormatRGB16UN
	case dto.VertexAttributeFormatRG16UN:
		return render.VertexAttributeFormatRG16UN
	case dto.VertexAttributeFormatR16UN:
		return render.VertexAttributeFormatR16UN

	case dto.VertexAttributeFormatRGBA8S:
		return render.VertexAttributeFormatRGBA8S
	case dto.VertexAttributeFormatRGB8S:
		return render.VertexAttributeFormatRGB8S
	case dto.VertexAttributeFormatRG8S:
		return render.VertexAttributeFormatRG8S
	case dto.VertexAttributeFormatR8S:
		return render.VertexAttributeFormatR8S

	case dto.VertexAttributeFormatRGBA8SN:
		return render.VertexAttributeFormatRGBA8SN
	case dto.VertexAttributeFormatRGB8SN:
		return render.VertexAttributeFormatRGB8SN
	case dto.VertexAttributeFormatRG8SN:
		return render.VertexAttributeFormatRG8SN
	case dto.VertexAttributeFormatR8SN:
		return render.VertexAttributeFormatR8SN

	case dto.VertexAttributeFormatRGBA8U:
		return render.VertexAttributeFormatRGBA8U
	case dto.VertexAttributeFormatRGB8U:
		return render.VertexAttributeFormatRGB8U
	case dto.VertexAttributeFormatRG8U:
		return render.VertexAttributeFormatRG8U
	case dto.VertexAttributeFormatR8U:
		return render.VertexAttributeFormatR8U

	case dto.VertexAttributeFormatRGBA8UN:
		return render.VertexAttributeFormatRGBA8UN
	case dto.VertexAttributeFormatRGB8UN:
		return render.VertexAttributeFormatRGB8UN
	case dto.VertexAttributeFormatRG8UN:
		return render.VertexAttributeFormatRG8UN
	case dto.VertexAttributeFormatR8UN:
		return render.VertexAttributeFormatR8UN

	case dto.VertexAttributeFormatRGBA8IU:
		return render.VertexAttributeFormatRGBA8IU
	case dto.VertexAttributeFormatRGB8IU:
		return render.VertexAttributeFormatRGB8IU
	case dto.VertexAttributeFormatRG8IU:
		return render.VertexAttributeFormatRG8IU
	case dto.VertexAttributeFormatR8IU:
		return render.VertexAttributeFormatR8IU

	default:
		panic(fmt.Errorf("unsupported vertex attribute format: %d", format))
	}
}

func (*ResourceSet) resolveIndexFormat(layout dto.IndexLayout) render.IndexFormat {
	switch layout {
	case dto.IndexLayoutUint16:
		return render.IndexFormatUnsignedU16
	case dto.IndexLayoutUint32:
		return render.IndexFormatUnsignedU32
	default:
		panic(fmt.Errorf("unsupported index layout: %d", layout))
	}
}
