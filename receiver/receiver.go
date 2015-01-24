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
package receiver

import (
	"fmt"
	"github.com/SamLex/mcudpt/common"
	"log"
	"net"
	"strings"
)

// Run receiver component
func Init(tcp, udp string) {
	log.Println("Running receiver")

	// Setup TCP listener
	senderConn, err := net.Listen("tcp", tcp)
	common.CheckErr("Cannot listen on TCP port", err)
	defer senderConn.Close()

	log.Printf("Listening on TCP address (%s) for sender connection \n", tcp)

	// Setup UDP listener
	recvConn, err := net.ListenPacket("udp", udp)
	common.CheckErr("Cannot listen on UDP port", err)
	defer recvConn.Close()

	log.Printf("Listening on UDP address (%s) for packets to forward \n", udp)

	// Accept sender connection
	tunnelConn, err := senderConn.Accept()
	common.CheckErr("Cannot accept sender connection", err)
	defer tunnelConn.Close()

	// Close TCP listen as only allowed one sender connection
	senderConn.Close()

	// Give user option to accept sender connection
	var ans string
	log.Printf("Sender connection from %s. Is this as expected? [yes/no] ", tunnelConn.RemoteAddr().String())
	_, err = fmt.Scan(&ans)
	common.CheckErr("Error reading user input", err)

	if strings.ToLower(ans)[0] == 'y' {
		log.Printf("Successfully accepted sender connnection from %s \n", tunnelConn.RemoteAddr().String())
	} else {
		log.Printf("Negative response received, closing \n")
		return
	}

	// Create tunnel
	tunnel := common.NewTunnel(tunnelConn, tunnelConn)

	// Setup tunnel receive handler
	tunnel.SetMessageReceiveFunc(func(tm *common.TunnelMessage) {
		pm := tm.PacketMessage

		// Resolve UDP address
		addr, err := net.ResolveUDPAddr("udp", pm.Host)
		common.CheckErr("Error resolving UDP address", err)

		// Send UDP packet
		_, err = recvConn.WriteTo(pm.Packet, addr)
		common.CheckErr("Error sending UDP packet", err)
	})

	// Starts tunnel traffic handling
	tunnel.Start()

	// Handle incoming UDP packets
	go func() {
		for {
			// Create packet message
			pm := common.NewTunnelPacketMessage()

			// Wait for new UDP packet
			len, addr, err := recvConn.ReadFrom(pm.Packet)
			common.CheckErr("Error reading from UDP connection", err)

			// Fill packet message
			pm.Host = addr.String()
			pm.Packet = pm.Packet[:len]

			// Write message to tunnel
			tunnel.Write(common.NewTunnelMessage(pm))
		}
	}()

	// Wait for an interrupt
	common.WaitForInterrupt()
	log.Println("Caught interrupt")
}
