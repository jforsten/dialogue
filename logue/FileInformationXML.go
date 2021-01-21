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
	"encoding/xml"
	"fmt"
)

type ProgramData struct {
	Information   string `xml:"Information"`
	ProgramBinary string `xml:"ProgramBinary"`
}

type Contents struct {
	NumLivesetData       int         `xml:"NumLivesetData,attr"`
	NumProgramData       int         `xml:"NumProgramData,attr"`
	NumPresetInformation int         `xml:"NumPresetInformation,attr"`
	NumTuneScaleData     int         `xml:"NumTuneScaleData,attr"`
	NumTuneOctData       int         `xml:"NumTuneOctData,attr"`
	ProgramData          ProgramData `xml:"ProgramData"`
}

type Korg struct {
	XMLName  xml.Name `xml:"KorgMSLibrarian_Data"`
	Product  string   `xml:"Product"`
	Contents Contents `xml:"Contents"`
}

func createFileInformationXML(product string) string {
	korg := &Korg{
		Product: product,
		Contents: Contents{
			NumLivesetData:       0,
			NumProgramData:       1,
			NumPresetInformation: 0,
			NumTuneScaleData:     0,
			NumTuneOctData:       0,
			ProgramData: ProgramData{
				Information: "Prog_000.prog_info", ProgramBinary: "Prog_000.prog_bin",
			},
		},
	}

	out, _ := xml.MarshalIndent(korg, " ", "  ")
	xmlStr := xml.Header + string(out)
	return xmlStr
}

// Program information XML ("Prog_NNN.prog_info") inside *.XXXprog package
func createProgramInfoXML(device string, programmer string, comment string) string {
	var outXML string
	outXML =  xml.Header + 
			  fmt.Sprintf("<%s_ProgramInformation>\n", device) +
			  fmt.Sprintf("  <Programmer>%s</Programmer>\n", programmer) +
			  fmt.Sprintf("  <Comment>%s</Comment>\n", comment) +
			  fmt.Sprintf("</%s_ProgramInformation>\n", device)
	return outXML
}
