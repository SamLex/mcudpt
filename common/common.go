/*
Copyright (c) 2014, SamLex (Euan James Hunter)
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

* Neither the name of mcudpt nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/
package common

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"os"
	"os/signal"
	"runtime"
)

// Tunnel used to pass UDP packets between reciever and sender
type Tunnel struct {
	reader         *json.Decoder
	writer         *json.Encoder
	receiveHandler func(tm *TunnelMessage)
}

// Create new Tunnel
func NewTunnel(reader io.Reader, writer io.Writer) (t *Tunnel) {
	t = &Tunnel{
		reader: json.NewDecoder(reader),
		writer: json.NewEncoder(writer),
	}
	return
}

// Write given message to tunnel
func (t *Tunnel) Write(tm *TunnelMessage) (err error) {
	err = t.writer.Encode(tm)
	return
}

// Read message from tunnel
func (t *Tunnel) Read() (tm *TunnelMessage, err error) {
	tm = &TunnelMessage{}
	err = t.reader.Decode(tm)
	return
}

// Set a handler to handle incoming messages
func (t *Tunnel) SetMessageReceiveFunc(f func(tm *TunnelMessage)) {
	t.receiveHandler = f
}

// Starts handling tunnel traffic. Returns instantly
func (t *Tunnel) Start() {
	go func() {
		for {
			// Read message from tunnel
			tm, err := t.Read()
			CheckErr("Error reading from tunnel", err)

			// Send message to handler
			t.receiveHandler(tm)
		}
	}()
}

// Meta message
type TunnelMessage struct {
	PacketMessage *TunnelPacketMessage
}

// Creates new tunnel message, with embedded packet message
func NewTunnelMessage(pm *TunnelPacketMessage) (tm *TunnelMessage) {
	tm = &TunnelMessage{
		PacketMessage: pm,
	}
	return
}

// UDP tunnling message
type TunnelPacketMessage struct {
	Host   string
	Packet []byte
}

// Creates new packet message
func NewTunnelPacketMessage() (pm *TunnelPacketMessage) {
	pm = &TunnelPacketMessage{
		Packet: make([]byte, math.MaxUint16-8-20),
	}
	return
}

// Check for an error, log and die if there is one
func CheckErr(message string, err error) {
	if err != nil {
		_, filename, line, _ := runtime.Caller(1)

		log.Fatalf("[%s:%d] [ERROR] %s (%s)", filename, line, message, err)
	}
}

// Waits for an interrupt (Ctrl+C) then returns
func WaitForInterrupt() {
	// Handle interrupt (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	select {
	case <-interrupt:
		return
	}
}
