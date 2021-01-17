//
// Dialogue is a tool for Korg Logue series of synths
// Copyright (C) 2021 Juha Forsten
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
//	"encoding/hex"
	"fmt"
	"time"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/sysex"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
)

type MidiPort struct {
	drv *(driver.Driver)
	wr *(writer.Writer)
	ins []midi.In
	outs []midi.Out
	in midi.In
	out midi.Out
	ch chan []byte
}

var midiPort MidiPort

func initializeMidi() error {
	var err error
	midiPort.drv, err = driver.New()

	midiPort.ins, err = midiPort.drv.Ins()
	checkError(err)

	midiPort.outs, err = midiPort.drv.Outs()
	checkError(err)

	midiPort.ch = make(chan []byte)

	return err
}

func getMidiPortNames() ([]string, []string) {
	var ins = []string{}
	var outs = []string{}

	for _, p := range midiPort.ins {
		ins = append(ins, p.String())
	}
	for _, p := range midiPort.outs {
		outs = append(outs, p.String())
	}
	return ins, outs
}

func closeMidi() {
	if midiPort.in != nil {
		midiPort.in.Close()
	}
	if midiPort.out != nil {
		midiPort.out.Close()
	}
	if midiPort.drv != nil {
		midiPort.drv.Close()
	}
}

func setMidi(inIdx int, outIdx int) error {

	if inIdx<0 || inIdx>len(midiPort.ins)-1 {
		return fmt.Errorf("In port is out of range!")
	}

	if outIdx<0 || outIdx> len(midiPort.outs)-1 {
		return fmt.Errorf("Out port is out of range!")
	}

	midiPort.in, midiPort.out = midiPort.ins[inIdx], midiPort.outs[outIdx]

	checkError(midiPort.in.Open())
	checkError(midiPort.out.Open())
	
	midiPort.wr = writer.New(midiPort.out)
	
	rd := reader.New(
		reader.NoLogger(),
		reader.IgnoreMIDIClock(),
		reader.SysEx(func(pos *reader.Position, data []byte) {
			//fmt.Printf("%s", hex.Dump(sysex.SysEx(data).Raw()))
			midiPort.ch<-sysex.SysEx(data).Raw()
		}),
		// write every message to the out port
		//reader.Each(func(pos *reader.Position, msg midi.Message) {
		//	fmt.Printf("got %s\n", msg)
		//}),
	)


	// listen for MIDI
	err := rd.ListenTo(midiPort.in)
	checkError(err)

	return err
}

func sendSysexAsync(sysexData []byte) <- chan []byte {
	replyChan := make(chan []byte, 1)
	if midiPort.wr != nil {
		err := writer.SysEx(midiPort.wr, sysexData)
		checkError(err)
	} else {
		fmt.Printf("Out port is not writeable!")
		replyChan <- nil
		return replyChan
	}
	
	select {
    case reply := <-midiPort.ch:
		replyChan <- reply
		return replyChan
    case <-time.After(2 * time.Second):
		fmt.Printf("ERROR: Timeout!")
		replyChan <- nil
		return replyChan
    }	
}

func sendControlChange(channel byte, controller byte, value byte) error {
	midiPort.wr.SetChannel(channel)
	return writer.ControlChange(midiPort.wr, controller, value)
}

func sendProgramChange(channel byte, program byte) error {
	midiPort.wr.SetChannel(channel)
	return writer.ProgramChange(midiPort.wr, program)
}

func sendNoteOn(channel byte, key byte, volume byte) error {
	midiPort.wr.SetChannel(channel)
	return writer.NoteOn(midiPort.wr, key, volume)
}

func sendNoteOff(channel byte, key byte) error {
	midiPort.wr.SetChannel(channel)
	return writer.NoteOff(midiPort.wr, key)
}