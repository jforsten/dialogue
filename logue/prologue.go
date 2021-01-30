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
	. "logue/logue/sysex/messagetype"
)

// Prologue specific Logue interface implementation
type Prologue struct {
	DeviceID byte
}

var a = map[byte]byte {}

var prologueSysexMap = map[byte]SysexMessageMap {
	GlobalDataDumpRequest : {CurrentProgramDataDump,0},
	CurrentProgramDataDumpRequest : {CurrentProgramDataDump, 0},
	ProgramDataDumpRequest : {ProgramDataDump, 2},
	UserSlotDataRequest : {UserSlotData, 3},
	UserSlotData : {UserSlotDataRequest, 0},
	UserSlotStatusRequest : {UserSlotStatus, 3},
	UserModuleInfoRequest : {UserModuleInfo, 2},		
	ClearUserSlot : {DataLoadCompleted, 0},
}

func (p Prologue) getDeviceSpecificInfo() DeviceSpecificInfo {
	return DeviceSpecificInfo{
		deviceID:                 p.DeviceID,
		familyID:                 0x4B,		
		deviceName:               "prologue",
		programInfoName:		  "prologue",
		programFileExtension:     "prlgprog",
		programDataFileExtension: ".prog_bin",
		programFilesize:          336,
		midiNamePrefix:           "prologue",
		programRange:             ProgramRange{1, 500},
		sysexMap: 				  prologueSysexMap,		
	}
}

