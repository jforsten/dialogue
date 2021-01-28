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
	"encoding/hex"
	"fmt"
	"hash/crc32"
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

func (p ProgramRange) has(programNumber int) bool {
	return programNumber >= logue.getDeviceSpecificInfo().programRange.min && programNumber <= logue.getDeviceSpecificInfo().programRange.max
}

// Relation between sent sysex type and expected return type and data
type SysexMessageMap struct {
	responseType           byte
	responseDataHeaderSize int
}

type DeviceSpecificInfo struct {
	deviceID                 byte // Global MIDI channel (1-16)
	familyID                 byte
	deviceName               string
	programInfoName          string
	programFileExtension     string
	programDataFileExtension string
	programFilesize          int
	midiNamePrefix           string
	programRange             ProgramRange
	sysexMap                 map[byte]SysexMessageMap
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

type response struct {
	err  error
	data []byte
}

const (
	// as defined in https://github.com/01org/isa-l/blob/master/crc/crc_base.c#L145
	// which is different from what golang considers the IEEE_standard:
	// https://godoc.org/hash/crc32#pkg-constants
	polynomial_ieee uint32 = 0x04C11DB7
)

// based on isal's crc32 algo found at:
// https://github.com/01org/isa-l/blob/master/crc/crc_base.c#L138-L155
func crc32_ieee_base(seed uint32, data []byte) (crc uint32) {
	rem := uint64(^seed)

	var i, j int

	const (
		// defined in
		// https://github.com/01org/isa-l/blob/master/crc/crc_base.c#L33
		MAX_ITER = 8
	)

	for i = 0; i < len(data); i++ {
		rem = rem ^ (uint64(data[i]) << 24)
		for j = 0; j < MAX_ITER; j++ {
			rem = rem << 1
			if (rem & 0x100000000) != 0 {
				rem ^= uint64(polynomial_ieee)
			}
		}
	}

	crc = uint32(^rem)
	return
}

func getData(requestType byte, requestDataHeader []byte, requestData []byte) <-chan response {

	var binData []byte
	var err error
	ch := make(chan response, 1)
	var sysexMessage []byte

	sysexMessage = createSysex(requestType, requestDataHeader, requestData)

	fmt.Printf("\nSEND:\n%s\n", hex.Dump(sysexMessage))

	replyChan := sendSysexAsync(sysexMessage)
	reply := <-replyChan

	if reply == nil {
		err = fmt.Errorf("ERROR: Communication not working!")
		ch <- response{err, nil}
		return ch
	}

	_, _, responseData := sysex.Response(reply)

	if len(responseData) > 10 {

		responseDataHeaderSize := logue.getDeviceSpecificInfo().sysexMap[requestType].responseDataHeaderSize
		//fmt.Printf("responseData:\n%s\n", hex.Dump(responseData))
		dataSection := responseData[responseDataHeaderSize:]
		//fmt.Printf("dataSection:\n%s\n", hex.Dump(dataSection))
		binData = convertSysexDataToBinaryData(dataSection)
		d := binData[8:]

		crc := crc32.ChecksumIEEE(d)
		fmt.Printf("size:%x, crc32:\n%x\n", len(binData[8:]), crc)

		dd := []byte{0x04, 0x06, 0x61, 0x04, 0x0C, 0x00, 0x00, 0x58, 0x15, 0x04, 0x00, 0x1C, 0x04,
			0x01, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 00, 00, 00, 00, 00, 01}
		fmt.Printf("\nT:%s\n", hex.Dump(convertSysexDataToBinaryData(dd[2:])))

		//fmt.Printf("binData:\n%s\n", hex.Dump(binData))
		//binData2 := convertSysexDataToBinaryData_(dataSection)
		//fmt.Printf("binData2:\n%s\n", hex.Dump(binData2))

		if binData == nil {
			err = fmt.Errorf("ERROR: Received wrong data!")
			ch <- response{err, nil}
			return ch
		}

	}

	ch <- response{err, binData}
	return ch
}

func SetProgram(programNumber int, filename string) <-chan error {
	var msgType byte
	var header []byte

	data := getDataFromZipFile(logue.getDeviceSpecificInfo().programFileExtension, filename)

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

	//fmt.Printf("\nRECEIVED:\n%s\n", hex.Dump(resp.data))

	err := saveProgramDataToFile(resp.data, filename)

	if err != nil {
		err := fmt.Errorf("ERROR: Wrong data!")
		errChan <- err
		return errChan
	}

	errChan <- resp.err
	return errChan
}

func SetUserSlotData(moduleID byte, slotID byte) <-chan error {

	errChan := make(chan error, 1)

	m := getDataFromZipFile(".json", "pluck.prlgunit")
	fmt.Printf("FILE:%s\n", string(m))
	man := sysex.ToModuleManifest(string(m))
	b := getDataFromZipFile(".bin", "pluck.prlgunit")
	plat, modData := man.CreateModuleData(b)
	fmt.Println(plat)
	fmt.Println(hex.Dump(modData))

	modData = append(modData, make([]byte,7)...)

	resp := <-getData(
		sysexMessageType.UserSlotData,
		//sysexMessageType.UserSlotStatusRequest,
		sysex.UserSlotHeader(moduleID, slotID),
		modData,
	)

	errChan <- resp.err
	return errChan

}

func GetUserSlotData(moduleID byte, slotID byte, filename string) <-chan error {

	resp := <-getData(
		//sysexMessageType.UserModuleInfoRequest,
		//[]byte{0x01},
		sysexMessageType.UserSlotDataRequest,
		//sysexMessageType.UserSlotStatusRequest,
		sysex.UserSlotHeader(moduleID, slotID),
		nil,
	)

	errChan := make(chan error, 1)

	if resp.data != nil {
		fmt.Printf("\nRECEIVED:\n%s\n", hex.Dump(resp.data))
		mod := sysex.ToModule(resp.data)
		dat := mod.FromModule()
		fmt.Printf("\nModule:\n%s\n", hex.Dump(dat))

		fmt.Println()
		sysex.TestJSON()

		/*
			err := saveProgramDataToFile(resp.data, filename)

			if err != nil {
				err := fmt.Errorf("ERROR: Wrong data!")
				errChan <- err
				return errChan
			}
		*/
	}
	errChan <- resp.err
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
