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

package sysex

// Start byte of Sysex message
const Start byte = 0xF0

// End byte of Sysex message
const End byte = 0xF7

// KorgID is manufacturer ID in sysex messages
const KorgID byte = 0x42

// Request returns sysex message with proper data to be sent
func Request(familyID byte, deviceID byte, messageType byte, data []byte) []byte {
	// deviceID = global MIDI channel -> '1'-based to '0'-based
	channel := 0x30 + deviceID - 1
	message := []byte{Start, KorgID, channel, 0x00, 0x01, familyID, messageType}
	message = append(message, data...)
	message = append(message, End)

	return message
}

// Response parses received sysex message
func Response(sysex []byte) (familyID byte, messageType byte, data []byte) {

	if len(sysex) < 8 || sysex[0] != Start || sysex[1] != KorgID {
		return 0, 0, nil
	}

	familyID = sysex[5]
	messageType = sysex[6]

	if len(sysex) == 8 {
		if sysex[7] == End {
			return familyID, messageType, nil
		} else {
			return 0, 0, nil
		}
	}

	data = sysex[7 : len(sysex)-1]

	return familyID, messageType, data
}

func ProgramNumber(number int) []byte {
	number--
	return []byte{byte(number) & 0b01111111, byte(number >> 7)}
}

func UserSlotHeader(moduleID byte, slotID byte) []byte {
	return []byte{moduleID, slotID}
}
