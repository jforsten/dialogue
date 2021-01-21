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

package logue

import (
	"fmt"
	"strings"
	"time"

	sysex "logue/logue/sysex"
	sysexMessageType "logue/logue/sysex/messagetype"
)

// Logue is generic interface for communicating Korg's logue series
type Logue interface {
	getDeviceSpecificInfo() DeviceSpecificInfo
}

type ProgramRange struct {
	min int
	max int
}

type DeviceSpecificInfo struct {
	deviceID                 byte // Global MIDI channel (1-16)
	familyID                 byte
	deviceName               string
	programInfoName          string
	programFileExtension     string
	programDataFileExtension string
	programFilesize          int

	midiNamePrefix string

	programRange ProgramRange
}

var logue Logue

func Open() error {
	return initializeMidi()
}

func SetDevice(lg Logue) {
	logue = lg
}

func findMidiPort(ports []string, prefix string, postfix string) int {
	for idx, in := range ports {
		if strings.Contains(in, prefix) && strings.Contains(in, postfix) {
			return idx
		}
	}
	return -1
}

func FindMidiIO() (int, int) {
	ins, outs := getMidiPortNames()

	return findMidiPort(ins, logue.getDeviceSpecificInfo().midiNamePrefix, "KBD/KNOB"),
		findMidiPort(outs, logue.getDeviceSpecificInfo().midiNamePrefix, "SOUND")
}

func ListMidiPorts() {

	ins, outs := getMidiPortNames()

	fmt.Println("  Available MIDI inputs:")
	for i, temp := range ins {
		fmt.Printf("    in %2d: %s\n", i, temp)
	}

	fmt.Println("\n  Available MIDI outputs:")
	for i, temp := range outs {
		fmt.Printf("    out%2d: %s\n", i, temp)
	}

}

func SetMidi(inIdx int, outIdx int) error {
	return setMidi(inIdx, outIdx)
}

// Prologue way of selecting program..
func SelectProgram(number int) error {
	if number < logue.getDeviceSpecificInfo().programRange.min || number > logue.getDeviceSpecificInfo().programRange.max {
		return fmt.Errorf("ERROR: Program number out of range!")
	}
	number--
	bankMsb := byte(0)
	bankLsb := byte(number / 100)
	num := byte(number % 100)

	sendNoteOn(logue.getDeviceSpecificInfo().deviceID-1, 1, 1)
	//time.Sleep(2 * time.Millisecond)
	sendNoteOff(logue.getDeviceSpecificInfo().deviceID-1, 1)
	sendControlChange(logue.getDeviceSpecificInfo().deviceID-1, 0x78, 0)
	time.Sleep(1 * time.Millisecond)

	sendControlChange(logue.getDeviceSpecificInfo().deviceID-1, 0x00, bankMsb)
	sendControlChange(logue.getDeviceSpecificInfo().deviceID-1, 0x20, bankLsb)
	sendProgramChange(logue.getDeviceSpecificInfo().deviceID-1, num)
	time.Sleep(1 * time.Millisecond)
	return nil
}

func createSysex(messageType byte, header []byte, data []byte) []byte {
	var buf []byte
	buf = append(header, convertBinaryDataToSysexData(data)...)
	return sysex.Request(
		logue.getDeviceSpecificInfo().familyID,
		logue.getDeviceSpecificInfo().deviceID,
		messageType,
		buf,
	)
}

func LoadProgramFile(programNumber int, filename string) <-chan error {
	var err error
	var sysexMessage []byte

	data := getDataFromZipFile(logue.getDeviceSpecificInfo(), filename)
	if programNumber < logue.getDeviceSpecificInfo().programRange.min || programNumber > logue.getDeviceSpecificInfo().programRange.max {
		sysexMessage = createSysex(sysexMessageType.CurrentProgramDataDump, nil, data)
	} else {
		sysexMessage = createSysex(sysexMessageType.ProgramDataDump, sysex.ProgramNumber(programNumber), data)
	}

	replyChan := sendSysexAsync(sysexMessage)
	reply := <-replyChan

	if reply == nil {
		err = fmt.Errorf("ERROR: Communication not working!")
	}

	errChan := make(chan error, 1)
	errChan <- err

	return errChan
}

func SaveProgramData(programNumber int, filename string) <-chan error {
	var err error
	errChan := make(chan error, 1)
	var sysexMessage []byte

	if programNumber < logue.getDeviceSpecificInfo().programRange.min || programNumber > logue.getDeviceSpecificInfo().programRange.max {
		sysexMessage = createSysex(
			sysexMessageType.CurrentProgramDataDumpRequest,
			nil,
			nil,
		)
	} else {
		sysexMessage = createSysex(
			sysexMessageType.CurrentProgramDataDumpRequest,
			sysex.ProgramNumber(programNumber),
			nil,
		)

	}

	replyChan := sendSysexAsync(sysexMessage)
	reply := <-replyChan

	if reply == nil {
		err = fmt.Errorf("ERROR: Communication not working!")
		errChan <- err
		return errChan
	}

	_, _, responseData := sysex.Response(reply)
	binData := convertSysexDataToBinaryData(responseData)

	if binData == nil {
		err = fmt.Errorf("ERROR: Received wrong data!")
		errChan <- err
		return errChan
	}

	err = saveProgramDataToFile(binData, filename)

	errChan <- err
	return errChan
}

func Close() {
	closeMidi()
}

func saveProgramDataToFile(data []byte, filename string) error {
	deviceName := logue.getDeviceSpecificInfo().deviceName
	fileInfoXML := createFileInformationXML(deviceName)

	programInfoXML := createProgramInfoXML(logue.getDeviceSpecificInfo().programInfoName, "", "")

	files := map[string][]byte{
		"FileInformation.xml": []byte(fileInfoXML),
		"Prog_000.prog_info":  []byte(programInfoXML),
		"Prog_000.prog_bin":   data,
	}

	err := createZipFile(filename, files)
	return err
}
