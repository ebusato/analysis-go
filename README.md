avirm/analysis-go [![GoDoc](https://godoc.org/github.com/ebusato/analysis-go?status.svg)](https://godoc.org/github.com/ebusato/analysis-go)
=================

`analysis-go` is a set of packages used for the acquisition and the analysis of avirm data. 

It can be used both for dpga and test bench data. 
It has been developped in order to be as general as possible.
It should be easy to implement the acquisition and analysis of future detectors (MCPPMT or any other) using this package.

Also, `analysis-go` is independent of the acquisition system. 
It is currently used for data acquired with the VME system but can also be used for data acquired with the microTCA system whenever it's ready.

`analysis-go` is developped using the go langage. It is based on the `gonum` and `go-hep` packages.

## Installation

```sh
$ go get gitlab.in2p3.fr/avirm/analysis-go
```

## Package structure

`analysis-go` contains several packages.

Packages specific to the dpga are in the `dpga` directory.

Packages specific to the test bench are in the `testbench` directory.

Packages `detector` and `pulse` are common to the dpga and test bench.

## DPGA packages

* The DAQ of the dpga detector is performed using the `dpga/godaq` package. 
  This package implements an executable that can be run to retrieve the data delivered by the CPU card and write them on disk. 
  While doing this, several cross checks are performed to check the data integrity.
  For testing and development purposes, it can retrieve data produced by the `dpga/rw/sendtosocket` executable.

* The pedestals of the dpga detector are computed using the `dpga/computePedestals` package.

* The analysis of the dpga data is performed using the `dpga/analysis` package.

## Test bench packages

* The pedestals of the test bench detector are computed using the `testbench/computePedestals` package.

* The analysis of the dpga data is performed using the `testbench/analysis` package.

## Common packages

* `pulse` implements structures and functions for analysing and plotting pulses.

* `detector` describes the detector compononents common to the dpga and test bench apparatus: capacitors, channels, quartets, drs, etc.
  Components specific to the dpga and test bench apparatus are defined in the `dpga/dpgadetector` and `testbench/tbdetector` packages respectively.



