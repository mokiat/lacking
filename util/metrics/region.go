package metrics

import (
	"time"

	"golang.org/x/exp/slices"
)

const NilParentID = -1

var (
	regionFreeID  int
	regions       = make(map[string]*Region)
	currentRegion *Region
	regionStats   []RegionStat
)

type RegionStat struct {
	ID       int
	ParentID int
	Depth    int

	Name     string
	Duration time.Duration
	Samples  int
}

func BeginRegion(name string) *Region {
	region, ok := regions[name]
	if !ok {
		region = &Region{
			id:   regionFreeID,
			name: name,
		}
		regionFreeID++
		regions[name] = region
	}
	region.parent = currentRegion
	region.startTime = time.Now()
	region.samples++
	currentRegion = region
	return region
}

func RegionStats() []RegionStat {
	regionStats = slices.Grow(regionStats, len(regions))
	regionStats = regionStats[:len(regions)]
	for _, region := range regions {
		stat := RegionStat{
			ID:       region.id,
			ParentID: NilParentID,
			Depth:    region.Depth(),
			Name:     region.name,
			Duration: region.duration,
			Samples:  region.samples,
		}
		region.duration = 0
		region.samples = 0
		if parent := region.parent; parent != nil {
			stat.ParentID = parent.id
		}
		regionStats[region.id] = stat
	}
	return regionStats
}

type Region struct {
	parent *Region

	id   int
	name string

	startTime time.Time
	duration  time.Duration
	samples   int
}

func (r *Region) Depth() int {
	if r.parent == nil {
		return 1
	}
	return 1 + r.parent.Depth()
}

func (r *Region) Name() string {
	return r.name
}

func (r *Region) Duration() time.Duration {
	return r.duration
}

func (r *Region) End() {
	elapsedTime := time.Since(r.startTime)
	r.duration += elapsedTime
	currentRegion = r.parent
}
