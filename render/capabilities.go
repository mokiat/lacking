package render

type Quality int

const (
	QualityLow Quality = iota
	QualityMedium
	QualityHigh
)

type Capabilities struct {
	Quality Quality
}
