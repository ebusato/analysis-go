package utils

import (
	"fmt"
	"log"
)

func FindIntersections(sine []float64, times []float64) []float64 {
	if len(sine) != len(times) {
		log.Fatalf("len(sine) != len(times)\n")
	}
	var min float64 = 1e6
	var max float64 = 0
	for i := 0; i < len(sine); i++ {
		ampl := sine[i]
		if ampl < min {
			min = ampl
		}
		if ampl > max {
			max = ampl
		}
		// 		fmt.Print(i, " ", sineWave[i])
		// 		for j := 0; j < int(math.Floor(sineWave[i])/10); j++ {
		// 			fmt.Print("*")
		// 		}
		// 		fmt.Print("\n")
	}
	av := (min + max) / 2.
	//fmt.Println("min, max, av =", min, max, av)
	var intersectionTimes []float64
	var amplPrev float64 = max
	// 	low := av - 0.2*av
	// 	high := av + 0.2*av
	// 	var timesInIntervals [5][]float64
	// 	var intervalIdx int = 0
	for i := 0; i < len(sine); i++ {
		//fmt.Println(times[i], sine[i])
		ampl := sine[i]

		// 		if ampl > low && ampl < high {
		// 			timesInIntervals[intervalIdx] = append(timesInIntervals[intervalIdx], times[i])
		// 		}
		// 		if ampl > high {
		// 			intervalIdx++
		// 		}
		if ampl > av && amplPrev <= av {
			isRisingFront := false
			if i+15 <= len(sine)-1 {
				if ampl < sine[i+15] { // consider only rising fronts
					isRisingFront = true
				}
			} else {
				if ampl > sine[i-15] { // consider only rising fronts
					isRisingFront = true
				}
			}
			if isRisingFront {
				intersectionTimes = append(intersectionTimes, times[i])
			}
		}
		amplPrev = ampl
	}

	t := CheckAndFix(intersectionTimes)
	return t
}

func CheckAndFix(intersectionTimes []float64) []float64 {
	// 	fmt.Println(len(intersectionTimes))
	// 	fmt.Println(intersectionTimes)

	var times []float64
	times = append(times, intersectionTimes[0])
	for i := 1; i < len(intersectionTimes); i++ {
		timeWrtPrevious := intersectionTimes[i] - intersectionTimes[i-1]
		if timeWrtPrevious > 1/24.85e6*1e9+5 { // 24.85 MHz is the HF frequency
			fmt.Println(intersectionTimes)
			log.Fatalf("Elapsed time since last intersection too large\n")
		} else if timeWrtPrevious < 1/24.85e6*1e9-5 {
			if timeWrtPrevious < 3 {
				continue
			} else {
				fmt.Println(intersectionTimes)
				log.Fatalf("Elapsed time since last intersection too small (ti, ti-1) = (%v, %v)\n", intersectionTimes[i], intersectionTimes[i-1])
			}
		} else {
			times = append(times, intersectionTimes[i])
		}
	}
	// 	fmt.Println(times)

	if len(times) > 5 {
		if times[len(times)-1]-times[0] > 197 {
			//fmt.Println("here", times)
			times = times[:len(times)-1]
		} else {
			fmt.Println(times)
			log.Fatalf("len(times) > 5\n")
		}
	}
	if len(times) < 5 {
		if len(times) <= 3 {
			fmt.Println(times)
			log.Fatalf("Warning: len(times) = %v\n", len(times))
		} else if len(times) == 4 {
			times = append([]float64{0}, times...)
			//fmt.Println(times)
		} else {
			log.Fatalf("Impossible !\n")
		}
	}
	return times
}
