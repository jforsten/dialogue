//
// Dialogue is a tool for Korg Logue series of synths
// Copyright (C) 2021 Juha Forsten
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
//
// INFO (Windows):
//
//  build static exe: go build -ldflags="-extldflags=-static" dialogue.go
//  (for size optimized build: go build -ldflags="-extldflags=-static" -ldflags="-s -w" dialogue.go)
//  (..or even further: upx --brute dialogue.exe)

package main

import "C"

import (
	"flag"
	"fmt"
	"logue/logue"
	"os"
)

var deviceID int
var filename string
var enablePortListing bool
var explicitMidiInIdx int
var explicitMidiOutIdx int
var patchNumber int
var receiveMode bool

func main() {
	
	// Get cmd line options
	flag.IntVar(&deviceID, "id", 1, "Midi channel of the device (DeviceID).")
	flag.IntVar(&explicitMidiInIdx, "in", -1, "Set Midi input (index) explicitely. -1 = Auto detect.")
	flag.IntVar(&explicitMidiOutIdx, "out", -1, "Set Midi output (index) explicitely. -1 = Auto detect.")
	flag.BoolVar(&enablePortListing, "l", false, "Show available MIDI ports.")
	flag.IntVar(&patchNumber, "p", -1, "Program number. -1 = Edit buffer.")
	flag.BoolVar(&receiveMode, "R", false, "READ patch from device and save to file.")
	
	flag.Parse()

	if len(flag.Args()) > 1 {
		fmt.Printf("Only one file at a time!")
		os.Exit(-1)
	}

	filename = flag.Arg(0)

	err := logue.Open()
	checkError(err)

	defer logue.Close()

	if enablePortListing { logue.ListMidiPorts() }

	logue.SetDevice(logue.Prologue{DeviceID: byte(deviceID)})

	var in, out int

	inFound, outFound := logue.FindMidiIO()

	if explicitMidiInIdx >= 0 { in = explicitMidiInIdx} else { in = inFound }
	if explicitMidiOutIdx >= 0 { out = explicitMidiOutIdx} else { out = outFound }

	if in < 0 || out < 0 {
		logue.ListMidiPorts()
		fmt.Printf("\nNo supported devices found! Please try to set MIDI in & out ports explicitely.")
		os.Exit(-1)
	}

	//fmt.Printf("Using MIDI (in:%d / out:%d) - channel <%d>\n", in, out, deviceID)
	
	err = logue.SetMidi(in,out)
	checkError(err)

	// Exit if no files to process...
	if filename == "" { 
		// Select program if opted even no files to process 
		if patchNumber > 0 { 
			fmt.Printf("Selecting program <%d>\n", patchNumber)
			logue.SelectProgram(patchNumber)
		}
		os.Exit(0) 
	}

	// Use patch number only if in valid range (1-500). Defaults to edit buffer...

	if receiveMode {
		err = <- logue.SaveProgramData(patchNumber, filename)
		checkError(err)
		if err == nil {
			fmt.Printf("\nProgram file '%s' saved to file!\n", filename) 
		}		
	} else {
		err = <- logue.LoadProgramFile(patchNumber, filename)
		checkError(err)
		if err == nil {
			fmt.Printf("\nProgram file '%s' sent to device!\n", filename) 
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Printf("\nERROR:%s", error.Error(err))
		os.Exit(-1)
	}
}
