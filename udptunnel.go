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

// udptunnel
package main

import (
	"bufio"
	"math"
	"net"
	"strconv"
	"strings"
)

const MAX_UDP_PACKET_SIZE = math.MaxUint16 - 8 - 20

type TunnelPacket struct {
	clientAddress        string
	tunneledPacketLength int
	tunneledPacket       []byte
}

func newTunnelPacket() (tp *TunnelPacket) {
	tp = new(TunnelPacket)
	tp.tunneledPacket = make([]byte, MAX_UDP_PACKET_SIZE)
	return
}

func (tp *TunnelPacket) readPacket(tunnel *bufio.Reader) (err error) {
	// Read client address from header
	tp.clientAddress, err = tunnel.ReadString('\n')
	if err != nil {
		return
	}

	// Read packet length from header
	tunneledPacketLengthString, err := tunnel.ReadString('\n')
	if err != nil {
		return
	}

	// Strip newlines from strings
	tp.clientAddress = strings.Trim(tp.clientAddress, "\n")
	tunneledPacketLengthString = strings.Trim(tunneledPacketLengthString, "\n")

	// Parse packet length into an int
	parsedInt, err := strconv.ParseInt(tunneledPacketLengthString, 36, 0)
	if err != nil {
		return
	}

	tp.tunneledPacketLength = int(parsedInt)

	// Read the UDP packet
	tp.tunneledPacket = tp.tunneledPacket[:tp.tunneledPacketLength]
	tp.tunneledPacketLength, err = tunnel.Read(tp.tunneledPacket)
	if err != nil {
		return
	}

	return
}

func (tp *TunnelPacket) writePacket(tunnel *bufio.Writer) (err error) {
	defer tunnel.Flush()

	// Write client address to header
	_, err = tunnel.WriteString(tp.clientAddress + "\n")
	if err != nil {
		return
	}

	// Write packet length to header
	_, err = tunnel.WriteString(strconv.FormatInt(int64(tp.tunneledPacketLength), 36) + "\n")
	if err != nil {
		return
	}

	// Write the UDP packet
	_, err = tunnel.Write(tp.tunneledPacket[:tp.tunneledPacketLength])
	if err != nil {
		return
	}

	return
}

func newTunnelReadWriter(conn net.Conn) (tunnelReadWriter *bufio.ReadWriter, err error) {
	// Setup buffered reader and writer
	tunnelReader := bufio.NewReaderSize(conn, MAX_UDP_PACKET_SIZE)
	tunnelWriter := bufio.NewWriterSize(conn, MAX_UDP_PACKET_SIZE)

	// Merge reader and writer
	tunnelReadWriter = bufio.NewReadWriter(tunnelReader, tunnelWriter)

	return
}
