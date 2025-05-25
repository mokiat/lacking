package game

import (
	"fmt"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game/asset/dto/meshdto"
	"github.com/mokiat/lacking/game/asset/dto/shadingdto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/render"
)

func (*ResourceSet) resolveWrapMode(wrap shadingdto.WrapMode) render.WrapMode {
	switch wrap {
	case shadingdto.WrapModeClamp:
		return render.WrapModeClamp
	case shadingdto.WrapModeRepeat:
		return render.WrapModeRepeat
	case shadingdto.WrapModeMirroredRepeat:
		return render.WrapModeMirroredRepeat
	default:
		panic(fmt.Errorf("unknown wrap mode: %v", wrap))
	}
}

func (*ResourceSet) resolveFiltering(filter shadingdto.FilterMode) render.FilterMode {
	switch filter {
	case shadingdto.FilterModeNearest:
		return render.FilterModeNearest
	case shadingdto.FilterModeLinear:
		return render.FilterModeLinear
	case shadingdto.FilterModeAnisotropic:
		return render.FilterModeAnisotropic
	default:
		panic(fmt.Errorf("unknown filter mode: %v", filter))
	}
}

func (*ResourceSet) resolveDataFormat(format shadingdto.TexelFormat) render.DataFormat {
	switch format {
	case shadingdto.TexelFormatRGBA8:
		return render.DataFormatRGBA8
	case shadingdto.TexelFormatRGBA16F:
		return render.DataFormatRGBA16F
	case shadingdto.TexelFormatRGBA32F:
		return render.DataFormatRGBA32F
	default:
		panic(fmt.Errorf("unknown format: %v", format))
	}
}

func (*ResourceSet) resolveCullMode(mode shadingdto.CullMode) render.CullMode {
	switch mode {
	case shadingdto.CullModeNone:
		return render.CullModeNone
	case shadingdto.CullModeFront:
		return render.CullModeFront
	case shadingdto.CullModeBack:
		return render.CullModeBack
	case shadingdto.CullModeFrontAndBack:
		return render.CullModeFrontAndBack
	default:
		panic(fmt.Errorf("unknown cull mode: %v", mode))
	}
}

func (*ResourceSet) resolveFaceOrientation(orientation shadingdto.FaceOrientation) render.FaceOrientation {
	switch orientation {
	case shadingdto.FaceOrientationCCW:
		return render.FaceOrientationCCW
	case shadingdto.FaceOrientationCW:
		return render.FaceOrientationCW
	default:
		panic(fmt.Errorf("unknown face orientation: %v", orientation))
	}
}

func (*ResourceSet) resolveComparison(comparison shadingdto.Comparison) render.Comparison {
	switch comparison {
	case shadingdto.ComparisonNever:
		return render.ComparisonNever
	case shadingdto.ComparisonLess:
		return render.ComparisonLess
	case shadingdto.ComparisonEqual:
		return render.ComparisonEqual
	case shadingdto.ComparisonLessOrEqual:
		return render.ComparisonLessOrEqual
	case shadingdto.ComparisonGreater:
		return render.ComparisonGreater
	case shadingdto.ComparisonNotEqual:
		return render.ComparisonNotEqual
	case shadingdto.ComparisonGreaterOrEqual:
		return render.ComparisonGreaterOrEqual
	case shadingdto.ComparisonAlways:
		return render.ComparisonAlways
	default:
		panic(fmt.Errorf("unknown comparison: %v", comparison))
	}
}

func (*ResourceSet) resolveTopology(primitive meshdto.Topology) render.Topology {
	switch primitive {
	case meshdto.TopologyPoints:
		return render.TopologyPoints
	case meshdto.TopologyLineList:
		return render.TopologyLineList
	case meshdto.TopologyLineStrip:
		return render.TopologyLineStrip
	case meshdto.TopologyTriangleList:
		return render.TopologyTriangleList
	case meshdto.TopologyTriangleStrip:
		return render.TopologyTriangleStrip
	default:
		panic(fmt.Errorf("unsupported primitive: %d", primitive))
	}
}

func (s *ResourceSet) resolveVertexFormat(layout meshdto.VertexLayout) graphics.MeshGeometryVertexFormat {
	var result graphics.MeshGeometryVertexFormat
	if attrib := layout.Coord; attrib.BufferIndex != meshdto.UnspecifiedBufferIndex {
		result.Coord = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Normal; attrib.BufferIndex != meshdto.UnspecifiedBufferIndex {
		result.Normal = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Tangent; attrib.BufferIndex != meshdto.UnspecifiedBufferIndex {
		result.Tangent = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.TexCoord; attrib.BufferIndex != meshdto.UnspecifiedBufferIndex {
		result.TexCoord = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Color; attrib.BufferIndex != meshdto.UnspecifiedBufferIndex {
		result.Color = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Weights; attrib.BufferIndex != meshdto.UnspecifiedBufferIndex {
		result.Weights = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	if attrib := layout.Joints; attrib.BufferIndex != meshdto.UnspecifiedBufferIndex {
		result.Joints = opt.V(graphics.MeshGeometryVertexAttribute{
			BufferIndex: uint32(attrib.BufferIndex),
			ByteOffset:  attrib.ByteOffset,
			Format:      s.resolveVertexAttributeFormat(attrib.Format),
		})
	}
	return result
}

func (*ResourceSet) resolveVertexAttributeFormat(format meshdto.VertexAttributeFormat) render.VertexAttributeFormat {
	switch format {
	case meshdto.VertexAttributeFormatRGBA32F:
		return render.VertexAttributeFormatRGBA32F
	case meshdto.VertexAttributeFormatRGB32F:
		return render.VertexAttributeFormatRGB32F
	case meshdto.VertexAttributeFormatRG32F:
		return render.VertexAttributeFormatRG32F
	case meshdto.VertexAttributeFormatR32F:
		return render.VertexAttributeFormatR32F

	case meshdto.VertexAttributeFormatRGBA16F:
		return render.VertexAttributeFormatRGBA16F
	case meshdto.VertexAttributeFormatRGB16F:
		return render.VertexAttributeFormatRGB16F
	case meshdto.VertexAttributeFormatRG16F:
		return render.VertexAttributeFormatRG16F
	case meshdto.VertexAttributeFormatR16F:
		return render.VertexAttributeFormatR16F

	case meshdto.VertexAttributeFormatRGBA16S:
		return render.VertexAttributeFormatRGBA16S
	case meshdto.VertexAttributeFormatRGB16S:
		return render.VertexAttributeFormatRGB16S
	case meshdto.VertexAttributeFormatRG16S:
		return render.VertexAttributeFormatRG16S
	case meshdto.VertexAttributeFormatR16S:
		return render.VertexAttributeFormatR16S

	case meshdto.VertexAttributeFormatRGBA16SN:
		return render.VertexAttributeFormatRGBA16SN
	case meshdto.VertexAttributeFormatRGB16SN:
		return render.VertexAttributeFormatRGB16SN
	case meshdto.VertexAttributeFormatRG16SN:
		return render.VertexAttributeFormatRG16SN
	case meshdto.VertexAttributeFormatR16SN:
		return render.VertexAttributeFormatR16SN

	case meshdto.VertexAttributeFormatRGBA16U:
		return render.VertexAttributeFormatRGBA16U
	case meshdto.VertexAttributeFormatRGB16U:
		return render.VertexAttributeFormatRGB16U
	case meshdto.VertexAttributeFormatRG16U:
		return render.VertexAttributeFormatRG16U
	case meshdto.VertexAttributeFormatR16U:
		return render.VertexAttributeFormatR16U

	case meshdto.VertexAttributeFormatRGBA16UN:
		return render.VertexAttributeFormatRGBA16UN
	case meshdto.VertexAttributeFormatRGB16UN:
		return render.VertexAttributeFormatRGB16UN
	case meshdto.VertexAttributeFormatRG16UN:
		return render.VertexAttributeFormatRG16UN
	case meshdto.VertexAttributeFormatR16UN:
		return render.VertexAttributeFormatR16UN

	case meshdto.VertexAttributeFormatRGBA8S:
		return render.VertexAttributeFormatRGBA8S
	case meshdto.VertexAttributeFormatRGB8S:
		return render.VertexAttributeFormatRGB8S
	case meshdto.VertexAttributeFormatRG8S:
		return render.VertexAttributeFormatRG8S
	case meshdto.VertexAttributeFormatR8S:
		return render.VertexAttributeFormatR8S

	case meshdto.VertexAttributeFormatRGBA8SN:
		return render.VertexAttributeFormatRGBA8SN
	case meshdto.VertexAttributeFormatRGB8SN:
		return render.VertexAttributeFormatRGB8SN
	case meshdto.VertexAttributeFormatRG8SN:
		return render.VertexAttributeFormatRG8SN
	case meshdto.VertexAttributeFormatR8SN:
		return render.VertexAttributeFormatR8SN

	case meshdto.VertexAttributeFormatRGBA8U:
		return render.VertexAttributeFormatRGBA8U
	case meshdto.VertexAttributeFormatRGB8U:
		return render.VertexAttributeFormatRGB8U
	case meshdto.VertexAttributeFormatRG8U:
		return render.VertexAttributeFormatRG8U
	case meshdto.VertexAttributeFormatR8U:
		return render.VertexAttributeFormatR8U

	case meshdto.VertexAttributeFormatRGBA8UN:
		return render.VertexAttributeFormatRGBA8UN
	case meshdto.VertexAttributeFormatRGB8UN:
		return render.VertexAttributeFormatRGB8UN
	case meshdto.VertexAttributeFormatRG8UN:
		return render.VertexAttributeFormatRG8UN
	case meshdto.VertexAttributeFormatR8UN:
		return render.VertexAttributeFormatR8UN

	case meshdto.VertexAttributeFormatRGBA8IU:
		return render.VertexAttributeFormatRGBA8IU
	case meshdto.VertexAttributeFormatRGB8IU:
		return render.VertexAttributeFormatRGB8IU
	case meshdto.VertexAttributeFormatRG8IU:
		return render.VertexAttributeFormatRG8IU
	case meshdto.VertexAttributeFormatR8IU:
		return render.VertexAttributeFormatR8IU

	default:
		panic(fmt.Errorf("unsupported vertex attribute format: %d", format))
	}
}

func (*ResourceSet) resolveIndexFormat(layout meshdto.IndexLayout) render.IndexFormat {
	switch layout {
	case meshdto.IndexLayoutUint16:
		return render.IndexFormatUnsignedU16
	case meshdto.IndexLayoutUint32:
		return render.IndexFormatUnsignedU32
	default:
		panic(fmt.Errorf("unsupported index layout: %d", layout))
	}
}
