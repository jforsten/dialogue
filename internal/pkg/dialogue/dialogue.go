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

package dialogue

import (
	//"encoding/hex"
	"encoding/hex"
	"fmt"
	"strings"

	message "dialogue/internal/pkg/dialogue/sysex/message"
	sysex "dialogue/internal/pkg/dialogue/sysex"
)

// Logue is generic interface for communicating Korg's logue series
type Dialogue interface {
	getDeviceSpecificInfo() DeviceSpecificInfo
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
}

var dlg Dialogue

var isDebug bool

func EnableDebugging() { isDebug = true }

func Open() error {
	return initializeMidi()
}

func SetDevice(d Dialogue) {
	dlg = d
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

	return findMidiPort(ins, dlg.getDeviceSpecificInfo().midiNamePrefix, "KBD/KNOB"),
		findMidiPort(outs, dlg.getDeviceSpecificInfo().midiNamePrefix, "SOUND")
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

func createSysex(messageType byte, header []byte, data []byte) []byte {
	var buf []byte
	buf = append(header, convertBinaryDataToSysexData(data)...)
	return sysex.Request(
		dlg.getDeviceSpecificInfo().familyID,
		dlg.getDeviceSpecificInfo().deviceID,
		messageType,
		buf,
	)
}

type response struct {
	err      error
	familyID byte
	msgType  byte
	data     []byte
}

func getData(requestType byte, requestDataHeader []byte, requestData []byte) <-chan response {

	var binData []byte
	var err error
	ch := make(chan response, 1)
	var sysexMessage []byte

	sysexMessage = createSysex(requestType, requestDataHeader, requestData)

	if isDebug {
		fmt.Printf("\nDEBUG: Sent SysEx:\n%s\n", hex.Dump(sysexMessage))
	}

	replyChan := sendSysexAsync(sysexMessage)
	reply := <-replyChan

	if reply == nil {
		err = fmt.Errorf("ERROR: Communication not working!")
		ch <- response{err, 0, 0, nil}
		return ch
	}

	if isDebug {
		fmt.Printf("\nDEBUG: Received SyEx:\n%s\n", hex.Dump(reply))
	}

	_, _, responseData := sysex.Response(reply)

	if len(responseData) > 10 {
		responseDataHeaderSize := message.ResponseInfo[requestType].HeaderSize
		dataSection := responseData[responseDataHeaderSize:]
		binData = convertSysexDataToBinaryData(dataSection)

		if binData == nil {
			err = fmt.Errorf("ERROR: Received wrong data!")
			ch <- response{err, reply[5], reply[6], nil}
			return ch
		}

	}

	ch <- response{err, reply[5], reply[6], binData}
	return ch
}

func Close() {
	closeMidi()
}
