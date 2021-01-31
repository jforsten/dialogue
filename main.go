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
	dlg "dialogue/dialogue"
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
		mode               = flag.String("m", "pw", "Operation mode: pw, pr, uw, ur, ui, ud.")
		moduleTypeSlot     = flag.String("s", "osc/0", "Module type & slot.")
		debug              = flag.Bool("d", false, "Enable extra debug prints.")
	)
	flag.Parse()

	if len(flag.Args()) > 1 {
		fmt.Printf("Only one file at a time!")
		os.Exit(-1)
	}

	filename := flag.Arg(0)

	if *debug {
		dlg.EnableDebugging()
	}

	err := dlg.Open()
	checkError(err)

	defer dlg.Close()

	if *enablePortListing {
		dlg.ListMidiPorts()
	}

	dlg.SetDevice(dlg.Prologue{DeviceID: byte(*deviceID)})

	var in, out int

	inFound, outFound := dlg.FindMidiIO()

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
		dlg.ListMidiPorts()
		fmt.Printf("\nNo supported devices found! Please try to set MIDI in & out ports explicitely.")
		os.Exit(-1)
	}

	if *debug {
		fmt.Printf("\nDEBUG: Using MIDI (in:%d / out:%d) - channel <%d>\n", in, out, *deviceID)
	}

	err = dlg.SetMidi(in, out)
	checkError(err)

	// Exit if no files to process...
	if filename == "" && !(*mode == "ud" || *mode == "ui") {
		// Select program if opted even no files to process
		if *patchNumber > 0 {
			fmt.Printf("Selecting program <%d>\n", *patchNumber)
			dlg.SelectProgram(*patchNumber)
		}
		os.Exit(0)
	}

	// Use patch number only if in valid range (1-500). Defaults to edit buffer...
	switch *mode {

	case "pr":
		err = <-dlg.GetProgram(*patchNumber, filename)
		checkError(err)
		if err == nil {
			fmt.Printf("\nProgram file '%s' saved to file!\n", filename)
		}

	case "pw":
		err = <-dlg.SetProgram(*patchNumber, filename)
		checkError(err)
		if err == nil {
			fmt.Printf("\nProgram file '%s' sent to device!\n", filename)
		}

	case "ur":
		err = <-dlg.GetUserSlotData(*moduleTypeSlot, filename)
		checkError(err)
		fmt.Printf("\nUser data read - %s!\n", *moduleTypeSlot)

	case "uw":
		err = <-dlg.SetUserSlotData(*moduleTypeSlot, filename)
		checkError(err)
		fmt.Printf("\nUser data sent to device!\n")

	case "ud":
		err = <-dlg.DeleteUserData(*moduleTypeSlot)
		checkError(err)
		fmt.Printf("\nUser data '%s' deleted!\n", *moduleTypeSlot)

	case "ui":
		err = <-dlg.GetUserDataInfo(*moduleTypeSlot)
		checkError(err)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Printf("\nERROR:%s", error.Error(err))
		os.Exit(-1)
	}
}
