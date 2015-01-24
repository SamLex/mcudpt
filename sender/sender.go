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
package sender

import (
	"github.com/SamLex/mcudpt/common"
	"log"
	"net"
)

// Run sender component
func Init(tcp, udp string) {
	log.Println("Running sender")

	// Setup tunnel connection
	tunnelConn, err := net.Dial("tcp", tcp)
	common.CheckErr("Cannot connect to reciever", err)
	defer tunnelConn.Close()

	log.Printf("Connected to receiver %s \n", tunnelConn.RemoteAddr())

	// Create tunnel
	tunnel := common.NewTunnel(tunnelConn, tunnelConn)

	hostConnMap := make(map[string]net.Conn)

	// Setup tunnel receive handler
	tunnel.SetMessageReceiveFunc(func(tm *common.TunnelMessage) {
		pm := tm.PacketMessage

		// Get connection from map
		_, ok := hostConnMap[pm.Host]

		if !ok {
			// Create new UDP connection
			conn, err := net.Dial("udp", udp)
			common.CheckErr("Cannot create UDP connection", err)

			hostConnMap[pm.Host] = conn

			// Handle incoming UDP packets on the new UDP connection
			go func(host string) {
				for {
					// Create packet message
					pm := common.NewTunnelPacketMessage()

					// Wait for new UDP packet
					len, err := conn.Read(pm.Packet)
					common.CheckErr("Error reading from UDP connection", err)

					// Fill packet message
					pm.Host = host
					pm.Packet = pm.Packet[:len]

					// Write message to tunnel
					tunnel.Write(common.NewTunnelMessage(pm))
				}
			}(pm.Host)
		}

		conn := hostConnMap[pm.Host]

		_, err := conn.Write(pm.Packet)
		common.CheckErr("Error writing to UDP connection", err)
	})

	// Starts tunnel traffic handling
	tunnel.Start()

	// Wait for an interrupt
	common.WaitForInterrupt()
	log.Println("Caught interrupt")
}
