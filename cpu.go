package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

// use 3 sec sample rate like htop
const SAMPLE_RATE = 3

func getCPUSample() (idle, total uint64) {
	// Get sample from /proc/stat
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}

	line := strings.Split(string(contents), "\n")[0]
	fields := strings.Fields(line)

	for i := 1; i < len(fields); i++ {
		val, err := strconv.ParseUint(fields[i], 10, 64)
		if err != nil {
			fmt.Println("Error: ", i, fields[i], err)
		}
		total += val
		if i == 4 { // idle is the 5th field in the cpu line
			idle = val
		}
	}
	return
}

func CPUPercentage() float64 {
	idle0, total0 := getCPUSample()
	time.Sleep(SAMPLE_RATE * time.Second)
	idle1, total1 := getCPUSample()
	idleTicks := float64(idle1 - idle0)
	totalTicks := float64(total1 - total0)
	cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks
	// fmt.Printf("CPU usage is %f%% [busy: %f, total: %f]\n", cpuUsage, totalTicks-idleTicks, totalTicks)
	return cpuUsage
}
