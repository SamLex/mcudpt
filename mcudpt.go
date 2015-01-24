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
package main

import (
	"flag"
	"fmt"
	"github.com/SamLex/mcudpt/receiver"
	"github.com/SamLex/mcudpt/sender"
	"log"
	"os"
)

// Flags
var mode *string = flag.String("mode", "receiver", "the run mode: receiver or sender")
var tcp *string = flag.String("tcp", ":1937", "the TCP address listen for a sender connection or connect to a receiver")
var udp *string = flag.String("udp", ":1937", "the UDP address to listen for packets to forward or where to send forwarded packets")

func main() {
	// Use custom usage message
	flag.Usage = printUsage

	// Parse flags
	flag.Parse()

	// Branch on mode
	switch *mode {
	case "receiver":
		receiver.Init(*tcp, *udp)
	case "sender":
		sender.Init(*tcp, *udp)
	default:
		log.Printf("Unknown mode %s \n", *mode)
	}
}

// Print custom usage
func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage of mcudpt:\n")
	flag.PrintDefaults()
}
