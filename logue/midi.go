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
	"fmt"
	"time"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/sysex"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
	//driver "github.com/jforsten/rtmididrv"
)

type midiConnection struct {
	drv  *(driver.Driver)
	wr   *(writer.Writer)
	ins  []midi.In
	outs []midi.Out
	in   midi.In
	out  midi.Out
	ch   chan []byte
}

var midiConn midiConnection

func initializeMidi() error {
	var err error
	midiConn.drv, err = driver.New()

	midiConn.ins, err = midiConn.drv.Ins()
	checkError(err)

	midiConn.outs, err = midiConn.drv.Outs()
	checkError(err)

	midiConn.ch = make(chan []byte)

	return err
}

func getMidiPortNames() ([]string, []string) {
	var ins = []string{}
	var outs = []string{}

	for _, p := range midiConn.ins {
		ins = append(ins, p.String())
	}
	for _, p := range midiConn.outs {
		outs = append(outs, p.String())
	}
	return ins, outs
}

func closeMidi() {
	if midiConn.in != nil {
		midiConn.in.Close()
	}
	if midiConn.out != nil {
		midiConn.out.Close()
	}
	if midiConn.drv != nil {
		midiConn.drv.Close()
	}
}

func setMidi(inIdx int, outIdx int) error {

	if inIdx < 0 || inIdx > len(midiConn.ins)-1 {
		return fmt.Errorf("In port is out of range!")
	}

	if outIdx < 0 || outIdx > len(midiConn.outs)-1 {
		return fmt.Errorf("Out port is out of range!")
	}

	midiConn.in, midiConn.out = midiConn.ins[inIdx], midiConn.outs[outIdx]

	checkError(midiConn.in.Open())
	checkError(midiConn.out.Open())

	midiConn.wr = writer.New(midiConn.out)

	rd := reader.New(
		reader.NoLogger(),
		reader.IgnoreMIDIClock(),
		reader.SysEx(func(pos *reader.Position, data []byte) {
			midiConn.ch <- sysex.SysEx(data).Raw()
		}),
		// write every message to the out port
		//reader.Each(func(pos *reader.Position, msg midi.Message) {
		//	fmt.Printf("got %s\n", msg)
		//}),
	)

	// listen for MIDI
	err := rd.ListenTo(midiConn.in)
	checkError(err)

	return err
}

func sendSysexAsync(sysexData []byte) <-chan []byte {
	replyChan := make(chan []byte, 1)
	if midiConn.wr != nil {
		err := writer.SysEx(midiConn.wr, sysexData)
		checkError(err)
	} else {
		fmt.Printf("Out port is not writeable!")
		replyChan <- nil
		return replyChan
	}

	select {
	case reply := <-midiConn.ch:
		replyChan <- reply
		return replyChan
	case <-time.After(15 * time.Second):
		fmt.Printf("ERROR: Timeout!")
		replyChan <- nil
		return replyChan
	}
}

func sendControlChange(channel byte, controller byte, value byte) error {
	midiConn.wr.SetChannel(channel)
	return writer.ControlChange(midiConn.wr, controller, value)
}

func sendProgramChange(channel byte, program byte) error {
	midiConn.wr.SetChannel(channel)
	return writer.ProgramChange(midiConn.wr, program)
}

func sendNoteOn(channel byte, key byte, volume byte) error {
	midiConn.wr.SetChannel(channel)
	return writer.NoteOn(midiConn.wr, key, volume)
}

func sendNoteOff(channel byte, key byte) error {
	midiConn.wr.SetChannel(channel)
	return writer.NoteOff(midiConn.wr, key)
}
