package reconstruction

import "gitlab.in2p3.fr/avirm/analysis-go/detector"

// Minimal returns the coordinates of the beta+ decay according to the "minimal approach algorithm"
// So far, the coordinates of the scintillators front faces are used while one should rather use
// the coordinates of the scintillators center
//    -> THIS NEEDS TO BE CHANGED
func Minimal(ch1, ch2 *detector.Channel, xbeam, ybeam float64) (x, y, z float64) {
	coeff := ((ch2.CartCoord.X-ch1.CartCoord.X)*(ch1.CartCoord.X-xbeam) + (ch2.CartCoord.Y-ch1.CartCoord.Y)*(ch1.CartCoord.Y-ybeam)) / ((ch2.CartCoord.X-ch1.CartCoord.X)*(ch2.CartCoord.X-ch1.CartCoord.X) + (ch2.CartCoord.Y-ch1.CartCoord.Y)*(ch2.CartCoord.Y-ch1.CartCoord.Y))
	x = ch1.CartCoord.X - coeff*(ch2.CartCoord.X-ch1.CartCoord.X)
	y = ch1.CartCoord.Y - coeff*(ch2.CartCoord.Y-ch1.CartCoord.Y)
	z = ch1.CartCoord.Z - coeff*(ch2.CartCoord.Z-ch1.CartCoord.Z)
	return
}
