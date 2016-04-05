package utils

import "math"

// CartCoord describes cartesian coordinates.
type CartCoord struct {
	X, Y, Z float64
}

// CylCoord describes cylindrical coordinates.
// R: distance in the (x,y) plane
// Phi: angle in radian with the x axis in the (x,y) plane
type CylCoord struct {
	R, Phi, Z float64
}

func CartesianToCylindical(cart CartCoord) (cyl CylCoord) {
	cyl.R = math.Hypot(cart.X, cart.Y)
	cyl.Phi = math.Atan2(cart.Y, cart.X)
	cyl.Z = cart.Z
	return
}

func CylindricalToCartesian(cyl CylCoord) (cart CartCoord) {
	s, c := math.Sincos(float64(cyl.Phi))
	cart.X = cyl.R * float64(c)
	cart.Y = cyl.R * float64(s)
	cart.Z = cyl.Z
	return
}
