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

// client
package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

var udpChannels = struct {
	sync.RWMutex
	channels map[string]chan *TunnelPacket
}{
	channels: make(map[string]chan *TunnelPacket),
}

func client(tcpAddress *net.TCPAddr, udpAddress *net.UDPAddr) {
	// Setup tunnel channel
	toTunnel := make(chan *TunnelPacket, TUNNEL_PACKET_QUEUE_SIZE)

	// Connect to tunnel server
	tunnel, err := net.DialTCP("tcp", nil, tcpAddress)
	if err != nil {
		log.Printf("Cannot connect to tunnel server %s: %s \n", tcpAddress, err)
		return
	}
	defer tunnel.Close()

	log.Printf("Connected to tunnel server %s \n", tcpAddress)

	// Setup tunnel
	tunnelReadWriter, err := newTunnelReadWriter(tunnel)
	if err != nil {
		log.Printf("Error setting up tunnel: %s \n", err)
		return
	}

	// Start listening
	go clientTCPin(tunnelReadWriter, udpAddress, toTunnel)
	go clientTCPout(tunnelReadWriter, toTunnel)

	// Setup channel for OS Interrupt signals (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	select {
	// Wait for interrupt
	case <-interrupt:
		log.Println("Caught interrupt")
		return
	}
}

func clientTCPin(tunnel *bufio.ReadWriter, udpAddress *net.UDPAddr, toTunnel chan *TunnelPacket) {
	for {
		// Allocate new tunnel packet
		tunnelPacket := newTunnelPacket()

		// Read tunnel packet
		err := tunnelPacket.readPacket(tunnel.Reader)
		if err != nil {
			if err == io.EOF {
				fatalErr(err, "Server closed tunnel")
			} else {
				fatalErr(err, "Error reading from tunnel")
			}
		}

		// Get fromTunnel channel
		udpChannels.RLock()
		fromTunnel, ok := udpChannels.channels[tunnelPacket.clientAddress]
		udpChannels.RUnlock()

		if !ok {
			// Setup new UDP socket and channel
			fromTunnel = make(chan *TunnelPacket, TUNNEL_PACKET_QUEUE_SIZE)

			udp, err := net.DialUDP("udp", nil, udpAddress)
			if err != nil {
				fatalErr(err, "Cannot create new UDP socket")
			}

			// Set 5 minute deadline on UDP socket to allow it to auto-close after 5 minutes of inactivity
			err = udp.SetDeadline(time.Now().Add(5 * time.Minute))
			if err != nil {
				fatalErr(err, "")
			}

			// Start listening on new socket
			go clientUDPin(tunnelPacket.clientAddress, *udp, toTunnel)
			go clientUDPout(*udp, fromTunnel)

			udpChannels.Lock()
			udpChannels.channels[tunnelPacket.clientAddress] = fromTunnel
			udpChannels.Unlock()
		}

		// Send receieved tunnel packet to get sent to host
		fromTunnel <- tunnelPacket
	}
}

func clientUDPin(clientAddress string, udp net.UDPConn, toTunnel chan *TunnelPacket) {
	for {
		// Allocate new tunnel packet
		tunnelPacket := newTunnelPacket()

		// Listen for UDP packet
		n, err := udp.Read(tunnelPacket.tunneledPacket)
		if err != nil {
			switch err.(type) {
			case net.Error:
				// If UDP socket timed out, close quietly
				networkError := err.(net.Error)
				if networkError.Timeout() {
					udpChannels.Lock()
					delete(udpChannels.channels, clientAddress)
					udpChannels.Unlock()

					udp.Close()

					return
				} else {
					fatalErr(err, "Error reading from UDP socket")
				}
			default:
				fatalErr(err, "Error reading from UDP socket")
			}

		}

		// Pack tunnel packet
		tunnelPacket.clientAddress = clientAddress
		tunnelPacket.tunneledPacketLength = n

		// Send received UDP packet to tunnel
		toTunnel <- tunnelPacket
	}
}

func clientTCPout(tunnel *bufio.ReadWriter, toTunnel chan *TunnelPacket) {
	for {
		select {
		// Listen for tunnel packet on toTunnel
		case tunnelPacket := <-toTunnel:
			// Write tunnel packet to tunnel
			err := tunnelPacket.writePacket(tunnel.Writer)
			if err != nil {
				fatalErr(err, "Error writing to tunnel")
			}
		}
	}
}

func clientUDPout(udp net.UDPConn, fromTunnel chan *TunnelPacket) {
	for {
		select {
		// Listen for tunnel packet on fromTunnel
		case tunnelPacket := <-fromTunnel:
			// Send tunnel packet to host
			_, err := udp.Write(tunnelPacket.tunneledPacket)
			switch err.(type) {
			case net.Error:
				// If UDP socket timed out, close quietly
				networkError := err.(net.Error)
				if networkError.Timeout() {
					udpChannels.Lock()
					delete(udpChannels.channels, tunnelPacket.clientAddress)
					udpChannels.Unlock()

					udp.Close()

					return
				} else {
					fatalErr(err, "Error writing from UDP socket")
				}
			default:
				fatalErr(err, "Error writing from UDP socket")
			}
		}
	}
}
