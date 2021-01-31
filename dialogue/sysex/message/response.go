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

package message

// ResponseEntry represent the response type and possible data header 
type ResponseEntry struct {
	Type       byte // MessageType
	HeaderSize int  // -1 = No data expected
}

// ResponseInfo defines relation between sent sysex type and expected return type and data header
var ResponseInfo = map[byte]ResponseEntry {
	GlobalDataDumpRequest : {GlobalDataDump,0},
	CurrentProgramDataDumpRequest : {CurrentProgramDataDump, 0},
	CurrentProgramDataDump : {DataLoadCompleted, -1},
	ProgramDataDumpRequest : {ProgramDataDump, 2},
	ProgramDataDump : {DataLoadCompleted, -1},
	UserSlotDataRequest : {UserSlotData, 3},
	UserSlotData : {DataLoadCompleted, -1},
	UserSlotStatusRequest : {UserSlotStatus, 3},
	UserModuleInfoRequest : {UserModuleInfo, 2},		
	ClearUserSlot : {DataLoadCompleted, -1},
}
