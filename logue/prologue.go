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

package logue

import (
	"encoding/xml"
)

// Prologue specific Logue interface implementation
type Prologue struct {
	DeviceID byte
}

func (p Prologue) getDeviceSpecificInfo() DeviceSpecificInfo {
	return DeviceSpecificInfo{
		deviceID: p.DeviceID,
		deviceName: "prologue",
		programFileExtension: "prlgprog",
		programDataFileExtension: ".prog_bin",
		programFilesize: 336,
		midiNamePrefix: "prologue",
		programRange: [2]int{1,500},
	}
}

// Sysex
func (p Prologue) getCurrentProgramSysexMessage() []byte {
	channel := 0x30 + p.DeviceID - 1
	return []byte{0xF0, 0x42, channel, 0x00, 0x01, 0x4B, 0x10, 0xF7}
}

func (p Prologue) setCurrentProgramSysexMessage(binaryData []byte) []byte {
	var outBuf []byte

	channel := 0x30 + p.DeviceID - 1
	outBuf = []byte {0xF0, 0x42, channel, 0x00, 0x01, 0x4B, 0x40}
	outBuf = append(outBuf, convertBinaryDataToSysexData(binaryData)...)
	outBuf = append(outBuf, 0XF7)

	return outBuf
}

func (p Prologue) getProgramSysexMessage(number int) []byte {
	number--
	channel := 0x30 + p.DeviceID - 1
	return []byte{0xF0, 0x42, channel, 0x00, 0x01, 0x4B, 0x1C, byte(number) & 0b01111111, byte(number >> 7) & 0b01111111, 0xF7}
}

func (p Prologue) setProgramSysexMessage(number int, binaryData []byte) []byte {
	var outBuf []byte
	number-- // range 1-500 -> 0-499 for message 
	channel := 0x30 + p.DeviceID - 1
	outBuf = []byte {0xF0, 0x42, channel, 0x00, 0x01, 0x4B, 0x4C, byte(number) & 0b01111111, byte(number >> 7) & 0b01111111}
	outBuf = append(outBuf, convertBinaryDataToSysexData(binaryData)...)
	outBuf = append(outBuf, 0XF7)

	return outBuf
}

func (p Prologue) extractBinaryDataFromDump(sysexMessage []byte) []byte {	
	size := len(sysexMessage)

	if len(sysexMessage) < 7 {
		return nil
	}

	switch {
	// Current program
	case sysexMessage[6] == 0x40 && size == 392:
		return convertSysexDataToBinaryData(sysexMessage[7:size-1])
	// Program
	case sysexMessage[6] == 0x4C && size == 394:
		return convertSysexDataToBinaryData(sysexMessage[9:size-1])
	}

	return nil
}
 
// Program information XML ("Prog_000.prog_info") inside *.prlgprog package
type PrologueProgramInformation struct {
	XMLName xml.Name `xml:"prologue_ProgramInformation"`
	Programmer string `xml:"Programmer"`
	Comment string `xml:"Comment"`
}

func (p Prologue) createProgramInfoXML(programmer string, comment string) string {
	info := PrologueProgramInformation{Programmer: programmer, Comment: comment}
	out, _ := xml.MarshalIndent(info, " ", "  ")
	xmlStr := xml.Header + string(out)
	return xmlStr
}