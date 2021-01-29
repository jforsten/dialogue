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
	"archive/zip"
	"bytes"
	"strconv"
	"strings"

	//"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
)

func checkError(err error) {
	if err != nil {
		fmt.Printf("error:%s", error.Error(err))
		panic(err.Error())
	}
}

func ParseModuleSlot(str string) (string, byte, error) {
	res := strings.Split(str, "/")
	if res == nil || len(res) != 2 {
		return "", 0, fmt.Errorf("Wrong  module/slot option format!")
	}
	moduleType := res[0]
	moduleSlot, err := strconv.Atoi(res[1])
	return moduleType, byte(moduleSlot), err
}

func convertBinaryDataToSysexData_(data []byte) []byte {
	var outBuffer []byte
	datalen := len(data)
	//if datalen%7 != 0 {
	//	panic("ERROR: Data cannot be converted")
	//}
	outBufferLen := datalen / 7 * 8
	outBuffer = make([]byte, outBufferLen)

	for i := 0; i < datalen/7; i++ {
		outBuffer[i*8] =
			(data[i*7]&0b10000000)>>7 +
				(data[i*7+1]&0b10000000)>>6 +
				(data[i*7+2]&0b10000000)>>5 +
				(data[i*7+3]&0b10000000)>>4 +
				(data[i*7+4]&0b10000000)>>3 +
				(data[i*7+5]&0b10000000)>>2 +
				(data[i*7+6]&0b10000000)>>1

		outBuffer[i*8+1] = data[i*7] & 0b01111111
		outBuffer[i*8+2] = data[i*7+1] & 0b01111111
		outBuffer[i*8+3] = data[i*7+2] & 0b01111111
		outBuffer[i*8+4] = data[i*7+3] & 0b01111111
		outBuffer[i*8+5] = data[i*7+4] & 0b01111111
		outBuffer[i*8+6] = data[i*7+5] & 0b01111111
		outBuffer[i*8+7] = data[i*7+6] & 0b01111111
	}
	return outBuffer
}

func convertBinaryDataToSysexData(data []byte) []byte {
	var outBuffer []byte
	datalen := len(data)
	for i := 0; i < datalen/7; i++ {
		outBuffer = append(outBuffer, binDataToSysexData(data[i*7:i*7+7])...)
	}
	if datalen%7 != 0 {
		outBuffer = append(outBuffer, binDataToSysexData(data[len(data)-(datalen%7):])...)
	}
	return outBuffer
}

func binDataToSysexData(data []byte) []byte {
	var outBuffer []byte
	outBuffer = make([]byte, len(data)+1)

	for i := 0; i < len(data); i++ {
		outBuffer[0] += (data[i] & 0b10000000) >> (7 - i)
		outBuffer[i+1] = data[i] & 0b01111111
	}
	return outBuffer
}

func sysexDataToBinData(data []byte) []byte {
	var outBuffer []byte
	outBuffer = make([]byte, len(data)-1)

	for i := 0; i < len(data)-1; i++ {
		outBuffer[i] = data[i+1] + ((data[0] << (len(data) - 1 - i)) & 0b10000000)
	}
	return outBuffer
}

func convertSysexDataToBinaryData(sysexData []byte) []byte {
	var outBuffer []byte
	datalen := len(sysexData)
	for i := 0; i < datalen/8; i++ {
		outBuffer = append(outBuffer, sysexDataToBinData(sysexData[i*8:i*8+8])...)
	}
	if datalen%8 != 0 {
		outBuffer = append(outBuffer, sysexDataToBinData(sysexData[len(sysexData)-(datalen%8):])...)
	}
	return outBuffer
}

func convertSysexDataToBinaryData_(sysexData []byte) []byte {
	var outBuffer []byte
	datalen := len(sysexData)
	fmt.Printf("datalen=%d   mod8=%d\n", datalen, datalen%8)
	if datalen%8 != 0 {
		//panic("ERROR: Data cannot be converted")
	}
	outBufferLen := datalen / 8 * 7
	outBuffer = make([]byte, outBufferLen)

	for i := 0; i < datalen/8; i++ {
		outBuffer[i*7] = sysexData[i*8+1] + ((sysexData[i*8] << 7) & 0b10000000)
		outBuffer[i*7+1] = sysexData[i*8+2] + ((sysexData[i*8] << 6) & 0b10000000)
		outBuffer[i*7+2] = sysexData[i*8+3] + ((sysexData[i*8] << 5) & 0b10000000)
		outBuffer[i*7+3] = sysexData[i*8+4] + ((sysexData[i*8] << 4) & 0b10000000)
		outBuffer[i*7+4] = sysexData[i*8+5] + ((sysexData[i*8] << 3) & 0b10000000)
		outBuffer[i*7+5] = sysexData[i*8+6] + ((sysexData[i*8] << 2) & 0b10000000)
		outBuffer[i*7+6] = sysexData[i*8+7] + ((sysexData[i*8] << 1) & 0b10000000)
	}
	return outBuffer
}

func getDataFromZipFile(extension string, zipFile string) []byte {
	var buf []byte

	// Open a zip archive for reading.
	r, err := zip.OpenReader(zipFile)
	checkError(err)
	defer r.Close()

	for _, f := range r.File {
		fmt.Println(f.Name)
		rc, err := f.Open()
		checkError(err)

		fmt.Println(f.Name)

		if filepath.Ext(f.Name) == extension {

			fmt.Println("FOUND")

			buf = make([]byte, f.UncompressedSize)
			_, err := rc.Read(buf)

			if err != nil {
				if err != io.EOF {
					checkError(err)
				}
			}
			break
		}
		rc.Close()
	}
	return buf
}

func createZipFile(outname string, fileList map[string][]byte) error {

	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	zipWriter := zip.NewWriter(buf)

	for name, content := range fileList {
		zipFile, err := zipWriter.Create(name)
		if err != nil {
			return err
		}
		_, err = zipFile.Write(content)
		if err != nil {
			return err
		}
	}

	err := zipWriter.Close()
	if err != nil {
		return err
	}

	//write the zipped data to the disk
	err = ioutil.WriteFile(outname, buf.Bytes(), 0777)
	return err
}
