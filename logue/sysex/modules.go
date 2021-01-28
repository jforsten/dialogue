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

import (
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"fmt"
)

type Version struct {
	Patch		byte
	Minor		byte
	Major		byte
}

func ToVersion(versionData []byte) Version {
	v := Version{}
	v.Patch = versionData[0]
	v.Minor = versionData[1]
	v.Major = versionData[2]
	return v
}

func (v Version) FromVersion() []byte {
	var buf []byte
	buf = append(buf, v.Patch)
	buf = append(buf, v.Minor)
	buf = append(buf, v.Major)
	buf = append(buf, 0x00)
	return buf
}

func (v Version) VersionString() string {
	return fmt.Sprintf("%d.%d-%d", v.Major, v.Minor, v.Patch)
}

func FromVersionString(ver string) Version {
	v := Version{}
	fmt.Sscanf(ver, "%d.%d-%d", &v.Major, &v.Minor, &v.Patch)
	return v
}
 
type Parameter struct  {
	MinValue	  int8 // 2's complement
	MaxValue	  int8 // 2's complement
	ParameterType string // 0x00 = %, 0x02 = ""
	Name	      string
}

func ToParameter(parameterData []byte) Parameter {
	p := Parameter{}
	p.MinValue = int8(parameterData[0])
	p.MaxValue = int8(parameterData[1])
	if parameterData[2] == 0x02 {
		p.ParameterType = "%"
	} else {
		p.ParameterType = ""
	}
	p.Name = string(parameterData[3:13])
	return p
}

func (p Parameter) FromParameter() []byte {
	var buf []byte
	buf = append(buf, byte(p.MinValue))
	buf = append(buf, byte(p.MaxValue))
	var t byte
	if p.ParameterType == "%" {
		t = 0x02
	} else {
		t = 0x00
	}
	buf = append(buf, t)
	buf = append(buf, []byte(p.Name)...)
	padding := make([]byte, 3 + 10 - len(p.Name))
	buf = append(buf, padding...)
	return buf
}

type Header struct {
	TotalSize		uint32
	Crc32			uint32
	ModuleID		byte
	APIVersion      Version
	DeveloperID		uint32
	ProgramID		uint32
	Version			Version
	Name            string
	NumOfParams     byte
	Parameters      []Parameter
	PayloadSize   uint32
}

func ToHeader(headerData []byte) Header {
	h := Header{}
	h.TotalSize = binary.LittleEndian.Uint32(headerData[0:4])
	fmt.Printf("TotalSize:%d", h.TotalSize)
	h.Crc32 = binary.LittleEndian.Uint32(headerData[4:8])
	h.ModuleID = headerData[8]
	h.APIVersion = ToVersion(headerData[10:13])
	h.DeveloperID = binary.LittleEndian.Uint32(headerData[14:18])
	h.ProgramID = binary.LittleEndian.Uint32(headerData[18:22])
	h.Version = ToVersion(headerData[22:25])
	h.Name = string(headerData[26:39])
	h.NumOfParams = headerData[40]

	for i:=0; i<int(h.NumOfParams); i++ {
		p := ToParameter(headerData[(44+i*16):(57+i*16)])
		h.Parameters = append(h.Parameters, p)
	}
	h.PayloadSize = binary.LittleEndian.Uint32(headerData[1028:1032])
	return h
}

func (h Header) FromHeader() []byte {
	var buf []byte
	
	// Size
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, h.TotalSize)
	buf = append(buf, size...)

	// CRC
	crc := make([]byte, 4)
	binary.LittleEndian.PutUint32(crc, h.Crc32)
	buf = append(buf, crc...)

	buf = append(buf, h.ModuleID)
	buf = append(buf, byte(0x01))
	buf = append(buf, h.APIVersion.FromVersion()...)

	// Dev ID
	devID := make([]byte, 4)
	binary.LittleEndian.PutUint32(devID, h.DeveloperID)
	buf = append(buf, devID...)

	// Prog ID
	progID := make([]byte, 4)
	binary.LittleEndian.PutUint32(progID, h.ProgramID)
	buf = append(buf, progID...)
	
	buf = append(buf, h.Version.FromVersion()...)

	buf = append(buf, []byte(h.Name)...)

	// Fill to begining of "Parameters" section
	buf = append(buf, make([]byte, 44-len(buf))...)

	buf[40] = h.NumOfParams

	for i:=0; i<int(h.NumOfParams); i++ {
		buf = append(buf, h.Parameters[i].FromParameter()...)
	}

	// Fill header block just before the last size field..
	buf = append(buf, make([]byte, 1028-len(buf))...)

	// NextBlockSize
	next := make([]byte, 4)
	binary.LittleEndian.PutUint32(next, h.PayloadSize)
	buf = append(buf, next...)

	return buf
}

type Module struct {
	Header Header
	Payload []byte
}

func ToModule(data []byte) Module {
	m := Module{}
	m.Header = Header{}
	m.Header = ToHeader(data[0:1032])
	fmt.Printf("\n%x %x %x %x\n", data[0], data[1], data[2], data[3])
	m.Payload = data[1032: 1032 + m.Header.PayloadSize]	
	fmt.Printf("\n ToMod HdrSize: %d, TotalSize:%d\n", len(m.Header.FromHeader()), m.Header.TotalSize)
	return m
}

func (m Module) FromModule() []byte {
	var buf []byte
	buf = append(buf, m.Header.FromHeader()...)
	buf = append(buf, m.Payload...)
	fmt.Printf("\nFromMod HdrSize: %d, TotalSize:%d\n", len(m.Header.FromHeader()), m.Header.TotalSize)
	buf = append(buf, make([]byte, int(m.Header.TotalSize) + 8 - len(buf))...)
	return buf
}

const TestManifest string = `
{
    "header" : 
    {
        "platform" : "prologue",
        "module" : "osc",
        "api" : "1.2-3",
        "dev_id" : 8,
        "prg_id" : 9,
        "version" : "5.6-7",
        "name" : "waves",
        "num_param" : 6,
        "params" : [
            ["Wave A",      0,  45,  ""],
            ["Wave B",      0,  43,  ""],
            ["Sub Wave",    0,  15,  ""],
            ["Sub Mix",     0, 100, "%"],
            ["Ring Mix",    0, 100, "%"],
            ["Bit Crush",   0, 100, "%"]
          ]
    }
}
`

func ModuleName(moduleID byte) string {
	switch moduleID {
	case 1: return "modfx"
	case 2: return "delfx"
	case 3: return "revfx"
	case 4: return "osc"
	default: return "unknown"
	}
}

func ModuleID(module string) byte {
	switch module {
	case "modfx": return 1
	case "delfx": return 2
	case "revfx": return 3
	case "osc":   return 4
	default: return 0
	}
}

type ModuleManifest struct {
	Header struct {
		Platform string          `json:"platform"`
		Module   string          `json:"module"`
		API      string          `json:"api"`
		DevID    int             `json:"dev_id"`
		PrgID    int             `json:"prg_id"`
		Version  string          `json:"version"`
		Name     string          `json:"name"`
		NumParam int             `json:"num_param"`
		Params   [][]interface{} `json:"params"`
	} `json:"header"`
}

func ToModuleManifest(jsonStr string) ModuleManifest {
	man := ModuleManifest{}
	json.Unmarshal([]byte(jsonStr), &man)
	return man
}

func (h Header) ToModuleManifest(platform string) ModuleManifest {
	man := ModuleManifest{}	
	man.Header.Platform = platform
	man.Header.Module = ModuleName(h.ModuleID)
	man.Header.API = h.APIVersion.VersionString()
	man.Header.DevID = int(h.DeveloperID)
	man.Header.PrgID = int(h.ProgramID)
	man.Header.Version = h.Version.VersionString()
	man.Header.Name = h.Name
	man.Header.NumParam = int(h.NumOfParams)
	for i := 0; i < man.Header.NumParam; i++ {
		man.Header.Params[0][0] = h.Parameters[i].Name
		man.Header.Params[0][1] = h.Parameters[i].MinValue
		man.Header.Params[0][2] = h.Parameters[i].MaxValue
		man.Header.Params[0][3] = h.Parameters[i].ParameterType
	}
	return man
}

func (mf ModuleManifest) FromModuleManifest() (string, Header) {
	var platform string
	h := Header{}
	h.ModuleID = ModuleID(mf.Header.Module)
	h.APIVersion = FromVersionString(mf.Header.API)
	h.DeveloperID = uint32(mf.Header.DevID)
	h.ProgramID = uint32(mf.Header.PrgID)
	h.Version = FromVersionString(mf.Header.Version)
	h.Name = mf.Header.Name
	h.NumOfParams = byte(mf.Header.NumParam)
	h.Parameters = []Parameter{}
	for i := 0; i < int(h.NumOfParams); i++ {
		p := Parameter{}
		p.Name = fmt.Sprintf("%v",mf.Header.Params[i][0])
		p.MinValue = int8((mf.Header.Params[i][1]).(float64))
		p.MaxValue = int8((mf.Header.Params[i][2]).(float64))
		p.ParameterType = fmt.Sprintf("%v",mf.Header.Params[i][3])
		h.Parameters = append(h.Parameters, p)
	}
	return platform, h
}

func TestJSON() {
	man := ModuleManifest{}
	json.Unmarshal([]byte(TestManifest), &man)
	fmt.Println(man)
	man.Header.Params[3][0] = "NONE"
	man.Header.Params[3][2] = 44
	
	fmt.Printf("Platform:%s, NumOfPar:%d Par3:%s\n", man.Header.Platform, man.Header.NumParam, man.Header.Params[3])

	s, _ := json.Marshal(&man)
	fmt.Println(string(s))

}

func (h Header) CreateManifestJSON(platform string) string {
	man := h.ToModuleManifest(platform)
	jsonBytes, _ := json.Marshal(man)
	return string(jsonBytes)
}

func (mf ModuleManifest) CreateModuleData(payload []byte) (string, []byte) {
	var platform string
	mod := Module{}
	platform, mod.Header = mf.FromModuleManifest()
	mod.Payload = payload
	mod.Header.PayloadSize = uint32(len(payload))
	mod.Header.TotalSize = 0xc84
	mod.Header.Crc32 =  crc32.ChecksumIEEE(mod.FromModule()[8:])
	
	return platform, mod.FromModule()
}