package locations

import (
	"sort"
	"strings"
)

type Points []Point

// ConvexHull returns the set of points that define the
// convex hull of p in CCW order starting from the left most.
func (p Points) ConvexHull() Points {
	// From https://en.wikibooks.org/wiki/Algorithm_Implementation/Geometry/Convex_hull/Monotone_chain
	// with only minor deviations.
	sort.Sort(p)
	var h Points

	// Lower hull
	for _, pt := range p {
		for len(h) >= 2 && !ccw(h[len(h)-2], h[len(h)-1], pt) {
			h = h[:len(h)-1]
		}
		h = append(h, pt)
	}

	// Upper hull
	for i, t := len(p)-2, len(h)+1; i >= 0; i-- {
		pt := p[i]
		for len(h) >= t && !ccw(h[len(h)-2], h[len(h)-1], pt) {
			h = h[:len(h)-1]
		}
		h = append(h, pt)
	}

	return h[:len(h)-1]
}

// ccw returns true if the three points make a counter-clockwise turn
func ccw(a, b, c Point) bool {
	return ((b.X - a.X) * (c.Y - a.Y)) > ((b.Y - a.Y) * (c.X - a.X))
}

func (p Points) Len() int      { return len(p) }
func (p Points) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p Points) Less(i, j int) bool {
	if p[i].X == p[j].X {
		return p[i].Y < p[i].Y
	}
	return p[i].X < p[j].X
}

func (p *Points) ParseFromStr(str string) {
	str = strings.TrimLeft(str, `{"`)
	str = strings.TrimRight(str, `"}`)
	tempPointsStr := strings.Split(str, `","`)
	for _, tp := range tempPointsStr {
		var point Point
		point.StrToFloat(tp)
		*p = append(*p, point)
	}
}
