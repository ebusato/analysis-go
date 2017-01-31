// Package computePedestal computes pedestals.
// It should be run before applyCorrCalib is used.
package selectCalib

import (
	"os"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
)

func Which(s string) {
	switch s {
	case "A1":
		dpgadetector.Det.ReadPedestalsFile(os.Getenv("HOME") + "/godaq/calib/periodA/v1/pedestals.csv")
		dpgadetector.Det.ReadTimeDepOffsetsFile(os.Getenv("HOME") + "/godaq/calib/periodA/v1/timeDepOffsets.csv")
		dpgadetector.Det.ReadEnergyCalibFile(os.Getenv("HOME") + "/godaq/calib/periodA/v1/energy.csv")
	}
}
