# dialogue

Ultimate command-line tool for Korg Logue series synthesizer.

## Usage

<b>NOTE:</b> Currently only Prologue is supported. Dialogue can send/receive patches in Korg Library format (*.prlgprog files).

<b>Future plan is to add support for other files and devices.</b>

Examples:


<i>To send a program to current position (edit buffer):</i>
<code> dialog MyPatch.prlgprog </code>

<i>To receive the current program (edit buffer):</i>
<code> dialog -W NewPatch.prlgprog </code>

<i>To send and save a program to position 100:</i>
<code> dialog -p 100 MyPatch.prlgprog </code>

<i>To receive program from position 100:</i>
<code> dialog -p 100 -W NewPatch.prlgprog </code>

<i>List MIDI ports:</i>
<code> dialog -l </code>

Is using direct USB-connection to device, MIDI in/out is automatically detected. Otherwise you can explicitly set them. Use <code>-l</code> option to list all available ports. Use <code>-id \<midi channel\></code> to match the device MIDI channel (default is 1).

