package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: icmp-ping <target>")
		os.Exit(1)
	}

	target := os.Args[1]

	// Resolve the target IP address
	ipAddr, err := net.ResolveIPAddr("ip4:icmp", target)
	if err != nil {
		fmt.Println("Error resolving target:", err)
		os.Exit(1)
	}

	// Create a new ICMP connection
	conn, err := net.DialIP("ip4:icmp", nil, ipAddr)
	if err != nil {
		fmt.Println("Error creating connection:", err)
		os.Exit(1)
	}

	// Set up the ICMP echo request packet
	icmpType := 8                  // ICMP echo request
	icmpCode := 0                  // Must be zero
	icmpId := os.Getpid() & 0xffff // Use the current process ID as the ICMP ID
	icmpSeq := 0                   // Sequence number, starts at 0
	icmpData := "Hello, world!"    // Payload data
	icmpBody := []byte(icmpData)
	icmpLen := len(icmpBody) + 8 // ICMP header is 8 bytes long
	icmpBuf := make([]byte, icmpLen)

	icmpBuf[0] = byte(icmpType)
	icmpBuf[1] = byte(icmpCode)
	icmpBuf[2] = byte(0) // ICMP checksum will be calculated automatically
	icmpBuf[3] = byte(0)
	icmpBuf[4] = byte(icmpId >> 8)
	icmpBuf[5] = byte(icmpId & 0xff)
	icmpBuf[6] = byte(icmpSeq >> 8)
	icmpBuf[7] = byte(icmpSeq & 0xff)
	copy(icmpBuf[8:], icmpBody)

	// Ping the target until we get a response
	fmt.Printf("Pinging %s...\n", target)
	for {
		_, err := conn.Write(icmpBuf)
		if err != nil {
			fmt.Println("Error sending ICMP packet:", err)
		}

		// Wait for a response
		replyBuf := make([]byte, 1500)
		conn.SetReadDeadline(time.Now().Add(time.Second))
		_, err = conn.Read(replyBuf)
		if err != nil {
			if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
				// Timeout, try again
				continue
			} else {
				fmt.Println("Error receiving ICMP packet:", err)
			}
		}

		// Check if the response is an ICMP echo reply
		replyType := int(replyBuf[0])
		replyCode := int(replyBuf[1])
		replyId := int(replyBuf[4])<<8 | int(replyBuf[5])
		replySeq := int(replyBuf[6])<<8 | int(replyBuf[7])
		replyData := string(replyBuf[8:])
		if replyType == 0 && replyCode == 0 && replyId == icmpId && replySeq == icmpSeq && replyData == icmpData {
			fmt.Println("Ping successful!")
			break
		} else {
			fmt.Printf("Received ICMP packet: type=%d code=%d id=%d seq=%d data=%q\n", replyType, replyCode, replyId, replySeq, replyData)
		}

		time.Sleep(time.Second * 5) // Wait 5 seconds before trying again
	}

	conn.Close()
}
