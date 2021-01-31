
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
	"time"

	sysex "logue/logue/sysex"
	sysexMessageType "logue/logue/sysex/message"
)

type ProgramRange struct {
	min int
	max int
}

func (p ProgramRange) has(programNumber int) bool {
	return programNumber >= logue.getDeviceSpecificInfo().programRange.min && programNumber <= logue.getDeviceSpecificInfo().programRange.max
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

func SetProgram(programNumber int, filename string) <-chan error {
	var msgType byte
	var header []byte

	data := getDataFromZipFile(logue.getDeviceSpecificInfo().programDataFileExtension, filename)

	if logue.getDeviceSpecificInfo().programRange.has(programNumber) {
		msgType = sysexMessageType.ProgramDataDump
		header = sysex.ProgramNumber(programNumber)
	} else {
		msgType = sysexMessageType.CurrentProgramDataDump
	}

	resp := <-getData(msgType, header, data)

	errChan := make(chan error, 1)
	errChan <- resp.err

	return errChan
}

func GetProgram(programNumber int, filename string) <-chan error {
	var msgType byte
	var header []byte

	if logue.getDeviceSpecificInfo().programRange.has(programNumber) {
		msgType = sysexMessageType.ProgramDataDumpRequest
		header = sysex.ProgramNumber(programNumber)
	} else {
		msgType = sysexMessageType.CurrentProgramDataDumpRequest
	}

	resp := <-getData(msgType, header, nil)

	errChan := make(chan error, 1)

	err := saveProgramDataToFile(resp.data, filename)

	if err != nil {
		err := fmt.Errorf("ERROR: Wrong data!")
		errChan <- err
		return errChan
	}

	errChan <- resp.err
	return errChan
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