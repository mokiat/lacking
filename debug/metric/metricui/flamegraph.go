package metricui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/std"
)

const (
	flamegraphIdealWidth = 800
	flamegraphRowHeigth  = 45
)

var FlameGraph = co.Define(&flamegraphComponent{})

type FlameGraphData struct {
	UpdateInterval   time.Duration
	AggregationRatio float64
	FocusPath        []string
}

type flamegraphComponent struct {
	co.BaseComponent

	font     *ui.Font
	fontSize float32

	element *ui.Element

	interval         time.Duration
	aggregationRatio float64
	path             []string

	tree metric.Span
}

func (c *flamegraphComponent) OnCreate() {
	c.font = co.OpenFont(c.Scope(), "ui:///roboto-bold.ttf")
	c.fontSize = 20.0

	data := co.GetOptionalData[FlameGraphData](c.Properties(), FlameGraphData{})
	c.interval = data.UpdateInterval
	c.aggregationRatio = data.AggregationRatio
	c.path = data.FocusPath
	if len(c.path) == 0 {
		if value, ok := os.LookupEnv("FLAME_FOCUS"); ok {
			c.path = strings.Split(value, ",")
		}
	}

	co.After(c.Scope(), c.interval, c.onRefresh)
}

func (c *flamegraphComponent) OnUpsert() {
	data := co.GetOptionalData[FlameGraphData](c.Properties(), FlameGraphData{})
	c.interval = data.UpdateInterval
	c.aggregationRatio = data.AggregationRatio
}

func (c *flamegraphComponent) Render() co.Instance {
	return co.New(std.Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(std.ElementData{
			Reference: &c.element,
			Essence:   c,
			IdealSize: opt.V(ui.Size{
				Width:  flamegraphIdealWidth,
				Height: flamegraphRowHeigth,
			}),
		})
	})
}

func (c *flamegraphComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	bounds := canvas.DrawBounds(element, true)
	canvas.Push()
	canvas.Translate(bounds.Position)
	c.renderSpan(canvas, c.tree, 1.0, sprec.ZeroVec2(), sprec.NewVec2(bounds.Width(), flamegraphRowHeigth))
	canvas.Pop()
}

func (c *flamegraphComponent) renderSpan(canvas *ui.Canvas, span metric.Span, ratio float32, position, size sprec.Vec2) {
	canvas.Push()
	canvas.ClipRect(position, size)
	canvas.Translate(position)

	fillColor := ui.MixColors(ui.Navy(), ui.Red(), ratio/2.0)
	canvas.Reset()
	canvas.SetStrokeColor(ui.Gray())
	canvas.SetStrokeSizeSeparate(1.0, 0.0)
	canvas.Rectangle(sprec.ZeroVec2(), size)
	canvas.Fill(ui.Fill{
		Color: ui.ColorWithAlpha(fillColor, 196),
	})
	canvas.Stroke()

	header := span.Name
	footer := fmt.Sprintf("%s | %.2f%%",
		span.Duration.Truncate(10*time.Microsecond),
		ratio*100,
	)
	headerWidth := c.font.LineWidth([]rune(header), c.fontSize)
	footerWidth := c.font.LineWidth([]rune(footer), c.fontSize)
	lineHeight := c.font.LineHeight(c.fontSize)
	headerPosition := sprec.Vec2{
		X: (size.X - headerWidth) / 2.0,
		Y: (size.Y - 2*lineHeight) / 2.0,
	}
	footerPosition := sprec.Vec2{
		X: (size.X - footerWidth) / 2.0,
		Y: headerPosition.Y + lineHeight,
	}
	canvas.FillTextLine([]rune(header), headerPosition, ui.Typography{
		Font:  c.font,
		Size:  c.fontSize,
		Color: ui.White(),
	})
	canvas.FillTextLine([]rune(footer), footerPosition, ui.Typography{
		Font:  c.font,
		Size:  c.fontSize,
		Color: ui.White(),
	})
	canvas.Pop()

	offset := float32(0.0)
	for _, child := range span.Children {
		ratio := float32(child.Duration.Seconds() / span.Duration.Seconds())
		childPosition := sprec.Vec2{
			X: position.X + offset,
			Y: position.Y + float32(flamegraphRowHeigth),
		}
		childSize := sprec.Vec2{
			X: ratio * size.X,
			Y: float32(flamegraphRowHeigth),
		}
		c.renderSpan(canvas, child, ratio, childPosition, childSize)
		offset += childSize.X
	}
}

func (c *flamegraphComponent) onRefresh() {
	tree, iterations := metric.FrameTree()
	iterations = max(1, iterations)
	if focusedNode, ok := c.findSpan(tree, c.path); ok {
		tree = focusedNode
	}
	c.updateTree(&c.tree, &tree, iterations)

	treeDepth := c.spanTreeDepth(tree)
	c.element.SetIdealSize(ui.Size{
		Width:  flamegraphIdealWidth,
		Height: treeDepth * flamegraphRowHeigth,
	})

	co.After(c.Scope(), c.interval, c.onRefresh)
}

func (c *flamegraphComponent) spanTreeDepth(node metric.Span) int {
	subTreeDepth := 0
	for _, child := range node.Children {
		subTreeDepth = max(subTreeDepth, c.spanTreeDepth(child))
	}
	return 1 + subTreeDepth
}

func (c *flamegraphComponent) findSpan(node metric.Span, path []string) (metric.Span, bool) {
	if len(path) == 0 {
		return metric.Span{}, false
	}
	if !strings.EqualFold(path[0], node.Name) {
		return metric.Span{}, false
	}
	if len(path) == 1 {
		return node, true
	}
	for _, child := range node.Children {
		if target, ok := c.findSpan(child, path[1:]); ok {
			return target, true
		}
	}
	return metric.Span{}, false
}

func (c *flamegraphComponent) updateTree(target, source *metric.Span, iterations int) {
	target.Name = source.Name
	target.Duration = time.Duration(dprec.Mix(float64(source.Duration), float64(target.Duration), c.aggregationRatio) / float64(iterations))
	if missing := len(source.Children) - len(target.Children); missing > 0 {
		target.Children = append(target.Children, make([]metric.Span, missing)...)
	}
	for i := range source.Children {
		c.updateTree(&target.Children[i], &source.Children[i], iterations)
	}
}
