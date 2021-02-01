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

	"fmt"

	sysex "dialogue/internal/pkg/dialogue/sysex"
	sysexMessageType "dialogue/internal/pkg/dialogue/sysex/message"
)

func SetUserSlotData(moduleTypeSlot string, filename string) <-chan error {

	moduleID, slotID, isOnlyModule, err := ParseModuleSlot(moduleTypeSlot)

	errChan := make(chan error, 1)

	if err != nil || isOnlyModule {
		err := fmt.Errorf("Wrong module slot definition. Please use 'module/slot' format!")
		errChan <- err
		return errChan
	}

	m := getDataFromZipFile(".json", filename)
	man := sysex.ToModuleManifest(m)
	b := getDataFromZipFile(".bin", filename)
	_, modData := man.CreateModuleData(b)

	resp := <-getData(
		sysexMessageType.UserSlotData,
		sysex.UserSlotHeader(moduleID, slotID),
		modData,
	)

	errChan <- resp.err
	return errChan
}

func GetUserSlotData(moduleTypeSlot string, filename string) <-chan error {

	moduleID, slotID, isOnlyModule, err := ParseModuleSlot(moduleTypeSlot)

	errChan := make(chan error, 1)

	if err != nil || isOnlyModule {
		err := fmt.Errorf("Wrong module slot definition. Please use 'module/slot' format!")
		errChan <- err
		return errChan
	}

	resp := <-getData(
		sysexMessageType.UserSlotDataRequest,
		sysex.UserSlotHeader(moduleID, slotID),
		nil,
	)

	if resp.data != nil {
		mod := sysex.ToModule(resp.data)
		files := map[string][]byte{
			mod.Header.Name + "/" + "manifest.json": []byte(mod.Header.CreateManifestJSON("prologue")),
			mod.Header.Name + "/" + "payload.bin":   mod.Payload,
		}

		err := createZipFile(filename, files)

		if err != nil {
			err := fmt.Errorf("ERROR:Cannot create file!")
			errChan <- err
			return errChan
		}

	}
	errChan <- resp.err
	return errChan
}

func DeleteUserData(moduleTypeSlot string) <-chan error {

	moduleID, slotID, isOnlyModule, err := ParseModuleSlot(moduleTypeSlot)

	errChan := make(chan error, 1)

	if err != nil {
		err := fmt.Errorf(err.Error())
		errChan <- err
		return errChan
	}

	var msgType byte
	var hdr []byte

	if isOnlyModule {
		msgType = sysexMessageType.ClearUserModule
		hdr = []byte{moduleID}
	} else {
		msgType = sysexMessageType.ClearUserSlot
		hdr = sysex.UserSlotHeader(moduleID, slotID)
	}

	resp := <-getData(msgType, hdr, nil)

	errChan <- resp.err
	return errChan
}

func GetUserDataInfo(moduleTypeSlot string) <-chan error {
	moduleID, slotID, isOnlyModule, err := ParseModuleSlot(moduleTypeSlot)

	errChan := make(chan error, 1)

	if err != nil {
		err := fmt.Errorf(err.Error())
		errChan <- err
		return errChan
	}

	var msgType byte
	var hdr []byte

	if isOnlyModule {
		msgType = sysexMessageType.UserModuleInfoRequest
		hdr = []byte{moduleID}
	} else {
		msgType = sysexMessageType.UserSlotStatusRequest
		hdr = sysex.UserSlotHeader(moduleID, slotID)
	}

	resp := <-getData(msgType, hdr, nil)

	if isOnlyModule {
		if len(resp.data) == 9 {
			mi := sysex.ToModuleInfo(resp.data)
			fmt.Printf("\nSlot:'%s' - Max slot size:%d, Max program size:%d, Slot count:%d\n\n",
				moduleTypeSlot,
				mi.MaxSlotSize,
				mi.MaxProgramSize,
				mi.SlotCount,
			)
		}
	} else {
		if resp.err == nil && len(resp.data) == 0 {
			fmt.Printf("\nSlot '%s' is empty!\n\n", moduleTypeSlot)
		} else {
			buf := make([]byte, 8)
			buf = append(buf, resp.data...)
			fill := make([]byte, 1032-len(buf))
			buf = append(buf, fill...)
			hdr := sysex.ToHeader(buf)
			fmt.Printf("\nSlot:'%s' - Name:%s, Ver:%s, API:%s, DevID:%d, ProgID:%d\n\n",
				moduleTypeSlot,
				hdr.Name,
				hdr.Version.VersionString(),
				hdr.APIVersion.VersionString(),
				hdr.DeveloperID,
				hdr.ProgramID,
			)
		}
	}

	errChan <- resp.err
	return errChan
}
