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

// mcudpt
package main

import (
	"flag"
	"log"
	"net"
	"runtime"
)

const TUNNEL_PACKET_QUEUE_SIZE uint = 64

func main() {
	// Allow usage of multiple OS threads
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Flags
	modeFlag := flag.String("mode", "server", "the run mode: client or server")
	tcpFlag := flag.String("tcp", ":1937", "the TCP address to listen on or connect to")
	udpFlag := flag.String("udp", ":1937", "the UDP address to listen on or send tunneled packets to")

	// Parse flags
	flag.Parse()

	// Resolve address flags
	tcpAddress, err := net.ResolveTCPAddr("tcp", *tcpFlag)
	if err != nil {
		log.Printf("Cannot resolve TCP address %s: %s \n", *tcpFlag, err)
		return
	}

	udpAddress, err := net.ResolveUDPAddr("udp", *udpFlag)
	if err != nil {
		log.Printf("Cannot resolve UCP address %s: %s \n", *udpFlag, err)
		return
	}

	switch *modeFlag {
	case "server":
		server(tcpAddress, udpAddress)
	case "client":
		client(tcpAddress, udpAddress)
	default:
		log.Printf("Unknown mode %s \n", *modeFlag)
	}
}

func fatalErr(err error, errorMessage string) {
	if errorMessage == "" {
		log.Fatalln(err)
	} else {
		log.Fatalf("%s : %s", errorMessage, err)
	}
}
