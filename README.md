# Dialogue

## Ultimate command-line tool for Korg Logue series synthesizer. 

---

Disclaimer: This is as much a Go Lang learning project as it is a proper tool for <i>logue</i> synths :-)  
## Usage

<b>NOTE:</b> Currently only Prologue is supported. <i>Dialogue</i> can transfer both patches (*.prlgprog) and user modules (*.prlgunit) to/from device. 


<b>Future plan is to add support for other files and devices.</b>

### Examples:<p>

* <i>List MIDI ports:</i><br>
<code> dialogue -l </code>

* <i>To send a program to <b>current position</b> (edit buffer):</i><br>
<code> dialogue MyPatch.prlgprog </code>

* <i>To read the <b>current program</b> (mode program-read: '-m pr'):</i><br>
<code> dialogue -m pr NewPatch.prlgprog </code>

* <i>To send and save a <b>program to position 100</b> (default mode: '-m pw'):</i><br>
<code> dialogue -p 100 MyPatch.prlgprog </code>

* <i>To receive <b>program from position 100</b>:</i><br>
<code> dialogue -m pr -p 100 NewPatch.prlgprog </code>

* <i>To get <b>user module info</b> of type ModultaionFX:</i><br>
<code> dialogue -m ui -s modfx </code>

* <i>To get <b>user module info</b> of DelayFX <b>from slot</b> 2:</i><br>
<code> dialogue -m ui -s delfx/2 </code>

* <i>To <b>send user module</b> ReverbFX to slot 2:</i><br>
<code> dialogue -m uw -s revfx/2 MyReverb.prlgunit </code>

* <i>To <b>receive user module</b> OSC from slot 5:</i><br>
<code> dialogue -m ur -s osc/5 NewOsc.prlgunit </code>

<br>
If using direct USB-connection to device, MIDI in/out is automatically detected. Otherwise you can explicitly set them (<code>-in</code> / <code>-out</code>). Use <code>-l</code> option to list all available ports. Use <code>-id \<midi channel\></code> to match the device MIDI channel (default is 1).

