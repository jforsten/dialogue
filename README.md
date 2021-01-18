# Dialogue

## Ultimate command-line tool for Korg Logue series synthesizer. 

---

Disclaimer: This is as much a Go Lang learning project as it is a proper tool for <i>logue</i> synths :-)  
## Usage

<b>NOTE:</b> Currently only Prologue is supported. <i>Dialogue</i> can send/receive patches in Korg Library format (*.prlgprog files).


<b>Future plan is to add support for other files and devices.</b>

Examples:


* <i>To send a program to current position (edit buffer):</i>
<code> dialogue MyPatch.prlgprog </code>

* <i>To receive the current program (edit buffer):</i>
<code> dialogue -W NewPatch.prlgprog </code>

* <i>To send and save a program to position 100:</i>
<code> dialogue -p 100 MyPatch.prlgprog </code>

* <i>To receive program from position 100:</i>
<code> dialogue -p 100 -W NewPatch.prlgprog </code>

* <i>List MIDI ports:</i>
<code> dialogue -l </code>

If using direct USB-connection to device, MIDI in/out is automatically detected. Otherwise you can explicitly set them. Use <code>-l</code> option to list all available ports. Use <code>-id \<midi channel\></code> to match the device MIDI channel (default is 1).

