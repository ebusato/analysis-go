// Temporary until H2D is implemented in hbook
// H2D implements the plotter.GridXYZ interface (used for making heatmap plots)
package utils

import (
	"math"
	"fmt"
)

type H2D struct {
	NbinsX      int
	LowX, HighX float64
	NbinsY      int
	LowY, HighY float64
	Vals        [][]float64
	Nentries    int
}

func NewH2D(nx int, lowx, highx float64, ny int, lowy, highy float64) *H2D {
	h := &H2D{
		NbinsX: nx,
		LowX:   lowx,
		HighX:  highx,
		NbinsY: ny,
		LowY:   lowy,
		HighY:  highy,
	}
	h.Vals = make([][]float64, ny)
	for i := range h.Vals {
		h.Vals[i] = make([]float64, nx)
	}
	return h
}

func (h *H2D) XCoordToIndex(coord float64) int {
	size := (h.HighX - h.LowX) / float64(h.NbinsX)
	switch {
	case coord < h.LowX:
		panic("coord < h.LowX")
	case coord > h.HighX:
		panic("coord > h.HighX")
	default:
		return int(math.Floor((coord - h.LowX) / float64(size)))
	}
}

func (h *H2D) YCoordToIndex(coord float64) int {
	size := (h.HighY - h.LowY) / float64(h.NbinsY)
	switch {
	case coord < h.LowY:
		panic("coord < h.LowX")
	case coord > h.HighY:
		panic("coord > h.HighX")
	default:
		return int(math.Floor((coord - h.LowY) / float64(size)))
	}
}

func (h *H2D) IndexToCoordX(i int) float64 {
	size := (h.HighX - h.LowX) / float64(h.NbinsX)
	switch {
	case i > h.NbinsX-1:
		panic("i > h.NbinsX - 1")
	case i < 0:
		panic("i < 0")
	default:
		return float64(i)*size + h.LowX + size/2
	}
}

func (h *H2D) IndexToCoordY(i int) float64 {
	size := (h.HighY - h.LowY) / float64(h.NbinsY)
	switch {
	case i > h.NbinsY-1:
		panic("i > h.NbinsX - 1")
	case i < 0:
		panic("i < 0")
	default:
		return float64(i)*size + h.LowY + size/2
	}
}

func (h *H2D) Fill(x, y float64) {
	inRange := true
	if x < h.LowX || x > h.HighX {
		fmt.Printf("In H2D.Fill method: x out of range (x=%v, LowX=%v, HighX=%v)\n", x, h.LowX, h.HighX)
		inRange = false
	}
	if y < h.LowY || y > h.HighY {
		fmt.Printf("In H2D.Fill method: y out of range (y=%v, LowY=%v, HighY=%v)\n", y, h.LowY, h.HighY)
		inRange = false
	}
	if inRange {
		idxX := h.XCoordToIndex(x)
		idxY := h.YCoordToIndex(y)
		//fmt.Println("idxs: ", idxX, idxY)
		h.Vals[idxX][idxY]++
		h.Nentries++
	}
}

func (h H2D) Dims() (c, r int) {
	c = h.NbinsX
	r = h.NbinsY
	return
}

func (h H2D) Z(c, r int) float64 {
	return -1 * h.Vals[r][c]
}

func (h H2D) X(c int) float64 {
	return h.IndexToCoordX(c)
}

func (h H2D) Y(r int) float64 {
	//fmt.Println("test", r, h.IndexToCoordY(r))
	return h.IndexToCoordY(r)
}
