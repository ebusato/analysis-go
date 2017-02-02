// Package computePedestal computes pedestals.
// It should be run before applyCorrCalib is used.
package selectCalib

import (
	"log"
	"os"
	"strings"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
)

func Which(s string) {
	chars := strings.Split(s, "")
	pedFile := os.Getenv("HOME") + "/godaq/calib/period" + chars[0] + "/v" + chars[1] + "/pedestals.csv"
	tdoFile := os.Getenv("HOME") + "/godaq/calib/period" + chars[0] + "/v" + chars[1] + "/timeDepOffsets.csv"
	enFile := os.Getenv("HOME") + "/godaq/calib/period" + chars[0] + "/v" + chars[1] + "/energy.csv"
	if _, err := os.Stat(pedFile); err == nil {
		dpgadetector.Det.ReadPedestalsFile(pedFile)
	} else {
		log.Printf("pedestal file %s does not exist\n", pedFile)
	}
	if _, err := os.Stat(tdoFile); err == nil {
		dpgadetector.Det.ReadTimeDepOffsetsFile(tdoFile)
	} else {
		log.Printf("tdo file %s does not exist\n", tdoFile)
	}
	if _, err := os.Stat(enFile); err == nil {
		dpgadetector.Det.ReadEnergyCalibFile(enFile)
	} else {
		log.Printf("en file %s does not exist\n", enFile)
	}
}
