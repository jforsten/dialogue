//
// Dialogue is a tool for Korg Logue series of synths
// Copyright (C) 2021 Juha Forstén
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

import (
	"flag"
	"fmt"
	"logue/logue"
	"os"
)

func main() {

	// Cmd line options
	var (
		deviceID           = flag.Int("id", 1, "Midi channel of the device (DeviceID).")
		explicitMidiInIdx  = flag.Int("in", -1, "Set Midi input (index) explicitely. -1 = Auto detect.")
		explicitMidiOutIdx = flag.Int("out", -1, "Set Midi output (index) explicitely. -1 = Auto detect.")
		enablePortListing  = flag.Bool("l", false, "Show available MIDI ports.")
		patchNumber        = flag.Int("p", -1, "Program number. -1 = Edit buffer.")
		mode               = flag.String("m", "pw", "Operation mode")
		moduleTypeSlot     = flag.String("s", "osc/0", "Module type & slot")
		debug              = flag.Bool("d", false, "Enable extra debug info")
	)
	flag.Parse()

	if len(flag.Args()) > 1 {
		fmt.Printf("Only one file at a time!")
		os.Exit(-1)
	}

	filename := flag.Arg(0)

	if *debug {
		logue.EnableDebugging()
	}

	err := logue.Open()
	checkError(err)

	defer logue.Close()

	if *enablePortListing {
		logue.ListMidiPorts()
	}

	logue.SetDevice(logue.Prologue{DeviceID: byte(*deviceID)})

	var in, out int

	inFound, outFound := logue.FindMidiIO()

	if *explicitMidiInIdx >= 0 {
		in = *explicitMidiInIdx
	} else {
		in = inFound
	}
	if *explicitMidiOutIdx >= 0 {
		out = *explicitMidiOutIdx
	} else {
		out = outFound
	}

	if in < 0 || out < 0 {
		logue.ListMidiPorts()
		fmt.Printf("\nNo supported devices found! Please try to set MIDI in & out ports explicitely.")
		os.Exit(-1)
	}

	//fmt.Printf("Using MIDI (in:%d / out:%d) - channel <%d>\n", in, out, deviceID)

	err = logue.SetMidi(in, out)
	checkError(err)

	// Exit if no files to process...
	if filename == "" && !(*mode == "ud" || *mode == "ui") {
		// Select program if opted even no files to process
		if *patchNumber > 0 {
			fmt.Printf("Selecting program <%d>\n", *patchNumber)
			logue.SelectProgram(*patchNumber)
		}
		os.Exit(0)
	}

	// Use patch number only if in valid range (1-500). Defaults to edit buffer...
	switch *mode {

	case "pr":
		err = <-logue.GetProgram(*patchNumber, filename)
		checkError(err)
		if err == nil {
			fmt.Printf("\nProgram file '%s' saved to file!\n", filename)
		}

	case "pw":
		err = <-logue.SetProgram(*patchNumber, filename)
		checkError(err)
		if err == nil {
			fmt.Printf("\nProgram file '%s' sent to device!\n", filename)
		}

	case "ur":
		err = <-logue.GetUserSlotData(*moduleTypeSlot, filename)
		checkError(err)
		fmt.Printf("\nUser data read - %s!\n", *moduleTypeSlot)

	case "uw":
		err = <-logue.SetUserSlotData(*moduleTypeSlot, filename)
		checkError(err)
		fmt.Printf("\nUser data sent to device!\n")

	case "ud":
		err = <-logue.DeleteUserData(*moduleTypeSlot)
		checkError(err)
		fmt.Printf("\nUser data '%s' deleted!\n", *moduleTypeSlot)

	case "ui":
		err = <-logue.GetUserDataInfo(*moduleTypeSlot)
		checkError(err)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Printf("\nERROR:%s", error.Error(err))
		os.Exit(-1)
	}
}
