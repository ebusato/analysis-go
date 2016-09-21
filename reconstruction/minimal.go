package reconstruction

import (
	"gitlab.in2p3.fr/avirm/analysis-go/detector"
	"fmt"
	"math"
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
func Minimal(ch1, ch2 *detector.Channel, xbeam, ybeam float64) (x, y, z float64) {
	coeff := ((ch2.CartCoord.X-ch1.CartCoord.X)*(ch1.CartCoord.X-xbeam) + (ch2.CartCoord.Y-ch1.CartCoord.Y)*(ch1.CartCoord.Y-ybeam)) / ((ch2.CartCoord.X-ch1.CartCoord.X)*(ch2.CartCoord.X-ch1.CartCoord.X) + (ch2.CartCoord.Y-ch1.CartCoord.Y)*(ch2.CartCoord.Y-ch1.CartCoord.Y))
	x = ch1.CartCoord.X - coeff*(ch2.CartCoord.X-ch1.CartCoord.X)
	y = ch1.CartCoord.Y - coeff*(ch2.CartCoord.Y-ch1.CartCoord.Y)
	z = ch1.CartCoord.Z - coeff*(ch2.CartCoord.Z-ch1.CartCoord.Z)
	
	 if math.IsNaN(x) || math.IsNaN(y) || math.IsNaN(z) {
		fmt.Printf("One is NaN: %v %v %v\n", math.IsNaN(x), math.IsNaN(y), math.IsNaN(z))
		fmt.Printf("  %v %v %v\n", ch1.CartCoord.X,  ch1.CartCoord.Y, ch1.CartCoord.Z)
		fmt.Printf("  %v %v %v\n", ch2.CartCoord.X,  ch2.CartCoord.Y, ch2.CartCoord.Z)
	}
	
	return
}
