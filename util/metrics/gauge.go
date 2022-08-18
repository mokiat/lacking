package metrics

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
)

var (
	gauges    = make(map[string]*Gauge)
	gaugeList []*Gauge
)

func RegisterGauge(name string) *Gauge {
	if _, ok := gauges[name]; ok {
		panic(fmt.Errorf("gauge with name %q already registered", name))
	}
	result := &Gauge{
		name: name,
	}
	gauges[name] = result
	return result
}

func GetGauge(name string) *Gauge {
	gauge, ok := gauges[name]
	if !ok {
		return NopGauge
	}
	return gauge
}

func Gauges() []*Gauge {
	if len(gauges) == len(gaugeList) {
		return gaugeList
	}
	gaugeList = make([]*Gauge, 0, len(gauges))
	for _, gauge := range gauges {
		gaugeList = append(gaugeList, gauge)
	}
	return gaugeList
}

var NopGauge = &Gauge{}

type Gauge struct {
	name        string
	value       float64
	smoothValue float64
}

func (g *Gauge) Name() string {
	return g.name
}

func (g *Gauge) Set(value float64) {
	g.value = value
	g.smoothValue = dprec.Mix(value, g.smoothValue, 0.1)
}

func (g *Gauge) Get() float64 {
	return g.value
}

func (g *Gauge) GetSmooth() float64 {
	return g.smoothValue
}
