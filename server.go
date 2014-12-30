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

// server
package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
)

var knownHosts = struct {
	sync.RWMutex
	hosts map[string]bool
}{
	hosts: make(map[string]bool),
}

func server(tcpAddress *net.TCPAddr, udpAddress *net.UDPAddr) {
	// Setup tunnel channels
	toTunnel := make(chan *TunnelPacket, TUNNEL_PACKET_QUEUE_SIZE)
	fromTunnel := make(chan *TunnelPacket, TUNNEL_PACKET_QUEUE_SIZE)

	// Setup UDP socket
	udp, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		log.Printf("Cannot listen on UDP address: %s \n", err)
		return
	}
	defer udp.Close()

	log.Printf("Listening on UDP %s \n", udp.LocalAddr())

	// Setup TCP socket
	tcp, err := net.ListenTCP("tcp", tcpAddress)
	if err != nil {
		log.Printf("Cannot listen on TCP address: %s \n", err)
		return
	}
	defer tcp.Close()

	log.Printf("Listening on TCP %s \n", tcp.Addr())

	// Accept client connection
	tunnel, err := tcp.Accept()
	if err != nil {
		log.Printf("Cannot accept client connection: %s \n", err)
		return
	}
	defer tunnel.Close()
	tcp.Close()

	log.Printf("Tunnel client connection from %s \n", tunnel.RemoteAddr())

	// Setup tunnel
	tunnelReadWriter, err := newTunnelReadWriter(tunnel)
	if err != nil {
		log.Printf("Error setting up tunnel: %s \n", err)
		return
	}

	// Start listening
	go serverUDPin(*udp, toTunnel)
	go serverTCPin(tunnelReadWriter, fromTunnel)
	go serverUDPout(*udp, fromTunnel)
	go serverTCPout(tunnelReadWriter, toTunnel)

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

func serverUDPin(udp net.UDPConn, toTunnel chan *TunnelPacket) {
	for {
		// Allocate new tunnel packet
		tunnelPacket := newTunnelPacket()

		// Listen for UDP packet
		n, clientAddr, err := udp.ReadFrom(tunnelPacket.tunneledPacket)
		if err != nil {
			fatalErr(err, "Error reading from UDP socket")
		}

		// Pack tunnel packet
		tunnelPacket.clientAddress = clientAddr.String()
		tunnelPacket.tunneledPacketLength = n

		// Update known hosts
		knownHosts.Lock()
		if !knownHosts.hosts[tunnelPacket.clientAddress] {
			log.Printf("New client detected on %s \n", tunnelPacket.clientAddress)
			knownHosts.hosts[tunnelPacket.clientAddress] = true
		}
		knownHosts.Unlock()

		// Send received UDP packet to tunnel
		toTunnel <- tunnelPacket
	}
}

func serverTCPin(tunnel *bufio.ReadWriter, fromTunnel chan *TunnelPacket) {
	for {
		// Allocate new tunnel packet
		tunnelPacket := newTunnelPacket()

		// Read tunnel packet
		err := tunnelPacket.readPacket(tunnel.Reader)
		if err != nil {
			if err == io.EOF {
				fatalErr(err, "Client closed tunnel")
			} else {
				fatalErr(err, "Error reading from tunnel")
			}
		}

		// Check host is known
		knownHosts.RLock()
		if knownHosts.hosts[tunnelPacket.clientAddress] {
			// Send receieved tunnel packet to get sent to client
			fromTunnel <- tunnelPacket
		} else {
			log.Fatalf("Unknown client (%s) detected in tunnel packet \n", tunnelPacket.clientAddress)
		}
		knownHosts.RUnlock()
	}
}

func serverUDPout(udp net.UDPConn, fromTunnel chan *TunnelPacket) {
	for {
		select {
		// Listen for tunnel packet on fromTunnel
		case tunnelPacket := <-fromTunnel:
			// Resolve client address
			clientAddr, err := net.ResolveUDPAddr("udp", tunnelPacket.clientAddress)
			if err != nil {
				fatalErr(err, "")
			}

			// Send tunnel packet to client
			_, err = udp.WriteToUDP(tunnelPacket.tunneledPacket, clientAddr)
			if err != nil {
				fatalErr(err, "Error writing to UDP socket")
			}
		}
	}
}

func serverTCPout(tunnel *bufio.ReadWriter, toTunnel chan *TunnelPacket) {
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
