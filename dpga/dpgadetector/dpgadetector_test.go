package dpgadetector_test

import (
	"fmt"
	"os"
	"testing"
	"text/tabwriter"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
)

func TestQuartetAbsIdxToRelIdx(t *testing.T) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintln(w, "iQuartetAbs\tiHemi\tiASM\tiDRS\tiQuartet")
	for iQuartetAbs := uint8(0); iQuartetAbs < 60; iQuartetAbs++ {
		iHemi, iASM, iDRS, iQuartet := dpgadetector.QuartetAbsIdxToRelIdx(iQuartetAbs)
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n", iQuartetAbs, iHemi, iASM, iDRS, iQuartet)
	}

	fmt.Fprintln(w)
	w.Flush()
}

func TestChannelAbsIdxToRelIdx(t *testing.T) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintln(w, "iChannelAbs\tiHemi\tiASM\tiDRS\tiQuartet\tiChannel")
	for iChannelAbs := uint16(0); iChannelAbs < 240; iChannelAbs++ {
		iHemi, iASM, iDRS, iQuartet, iChannel := dpgadetector.ChannelAbsIdx240ToRelIdx(iChannelAbs)
		// closure test
		_, iChannelAbs2 := dpgadetector.RelIdxToAbsIdx240(iHemi, iASM, iDRS, iQuartet, iChannel)
		if iChannelAbs2 != iChannelAbs {
			t.Errorf("closure test fails")
		}
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\n", iChannelAbs, iHemi, iASM, iDRS, iQuartet, iChannel)
	}

	fmt.Fprintln(w)
	w.Flush()
}
