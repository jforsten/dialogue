package sysex

// Start byte of Sysex message
const Start byte = 0xF0

// End byte of Sysex message
const End byte = 0xF7

// KorgID is manufacturer ID in sysex messages
const KorgID byte = 0x42

// Request returns sysex message with proper data to be sent
func Request(familyID byte, deviceID byte, messageType byte, data []byte) []byte {
	channel := 0x30 + deviceID
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

	data = sysex[7 : len(data)-1]

	return familyID, messageType, data
}

func ProgramNumber(number int) [2]byte {
	number--
	return [2]byte {byte(number) & 0b01111111, byte(number>>7)} 
}

