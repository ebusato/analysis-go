package reconstruction

import (
	"fmt"
	"math"

	"gitlab.in2p3.fr/avirm/analysis-go/detector"
)

// Minimal returns the coordinates of the beta+ decay according to the "minimal approach algorithm"
// This minimal approach says that the coordinates of the beta+ decay are those of the point on the
// LOR which is closest to the beam axis.
// The formulas implemented below are demonstrated for example here:
//     http://geomalgorithms.com/a07-_distance.html
// with :
//    - line L1 = LOR
//    - line L2 = beam
//    - Q0 = (xbeam, ybeam, 0)
//    - u = (x2-x1, y2-y1, z2-z1)
//    - v = (0, 0, 1)
//    - w0 = (x1-x0, y1-y0, z1)
// The quantity called coeff below corresponds to the quantity called sc in the above web page
// (multiplied by -1).
//
// Important remark: so far, the coordinates of the scintillators front faces are used while one
// should rather use the coordinates of the scintillators center
//    -> THIS NEEDS TO BE CHANGED
func Minimal(crystCenter bool, ch1, ch2 *detector.Channel, xbeam, ybeam float64) (x, y, z float64) {
	var x1, x2, y1, y2, z1, z2 float64
	switch crystCenter {
	case true:
		x1 = ch1.CrystCenter.X
		x2 = ch2.CrystCenter.X
		y1 = ch1.CrystCenter.Y
		y2 = ch2.CrystCenter.Y
		z1 = ch1.CrystCenter.Z
		z2 = ch2.CrystCenter.Z
	case false:
		x1 = ch1.CartCoord.X
		x2 = ch2.CartCoord.X
		y1 = ch1.CartCoord.Y
		y2 = ch2.CartCoord.Y
		z1 = ch1.CartCoord.Z
		z2 = ch2.CartCoord.Z
	}
	coeff := ((x2-x1)*(x1-xbeam) + (y2-y1)*(y1-ybeam)) / ((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))
	x = x1 - coeff*(x2-x1)
	y = y1 - coeff*(y2-y1)
	z = z1 - coeff*(z2-z1)

	if math.IsNaN(x) || math.IsNaN(y) || math.IsNaN(z) {
		fmt.Printf("One is NaN: %v %v %v\n", math.IsNaN(x), math.IsNaN(y), math.IsNaN(z))
		fmt.Printf("  %v %v %v\n", x1, y1, z1)
		fmt.Printf("  %v %v %v\n", x2, y2, z2)
	}

	return
}
