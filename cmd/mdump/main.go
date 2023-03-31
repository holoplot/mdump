package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"os"
	"syscall"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/ipv4"
)

func multicastOpen(bindAddr net.IP, port int, ifname string) (*ipv4.PacketConn, error) {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		return nil, err
	}

	if err := syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return nil, err
	}

	//syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
	if err := syscall.SetsockoptString(s, syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, ifname); err != nil {
		return nil, err
	}

	lsa := syscall.SockaddrInet4{Port: port}
	copy(lsa.Addr[:], bindAddr.To4())

	if err := syscall.Bind(s, &lsa); err != nil {
		syscall.Close(s)
		return nil, err
	}

	f := os.NewFile(uintptr(s), "")
	defer f.Close()

	c, err := net.FilePacketConn(f)
	if err != nil {
		return nil, err
	}

	p := ipv4.NewPacketConn(c)

	return p, nil
}

func main() {
	consoleWriter := zerolog.ConsoleWriter{
		Out: colorable.NewColorableStdout(),
	}

	log.Logger = log.Output(consoleWriter)

	interfaceFlag := flag.String("i", "", "interface name")
	groupFlag := flag.String("g", "", "Multicast group")
	portFlag := flag.Int("p", 0, "Sender port")
	protocolFlag := flag.String("proto", "udp", "protocol")
	senderAddressFlag := flag.String("sa", "0.0.0.0", "Sender address")
	dumpFlag := flag.Bool("x", false, "Dump packets in hex")
	flag.Parse()

	if *interfaceFlag == "" || *groupFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	ifi, err := net.InterfaceByName(*interfaceFlag)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("interface", *interfaceFlag).
			Msg("unable to get interface")
	}

	senderAddr := net.ParseIP(*senderAddressFlag)
	if senderAddr == nil {
		log.Fatal().
			Str("address", *senderAddressFlag).
			Msg("bad address")
	}

	groupAddr := net.ParseIP(*groupFlag)
	if groupAddr == nil {
		log.Fatal().
			Str("group", *groupFlag).
			Msg("bad group")
	}

	c, err := multicastOpen(groupAddr, *portFlag, *interfaceFlag)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("interface", *interfaceFlag).
			Int("port", *portFlag).
			Msg("unable to open multicast socket")
	}

	if err := c.JoinGroup(ifi, &net.UDPAddr{IP: groupAddr}); err != nil {
		log.Fatal().
			Err(err).
			Str("interface", *interfaceFlag).
			Str("group", *groupFlag).
			Msg("unable to join multicast group")

	}

	if err := c.SetControlMessage(ipv4.FlagTTL|ipv4.FlagSrc|ipv4.FlagDst|ipv4.FlagInterface, true); err != nil {
		log.Fatal().
			Err(err).
			Str("interface", *interfaceFlag).
			Str("group", *groupFlag).
			Msg("unable to set control message")
	}

	log.Info().
		Str("interface", ifi.Name).
		Str("protocol", *protocolFlag).
		Str("group", *groupFlag).
		Str("address", *senderAddressFlag).
		Int("port", *portFlag).
		Msg("Starting multicast listener")

	buf := make([]byte, 65535)

	for {
		n, cm, _, err := c.ReadFrom(buf)
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("error reading from socket")
		}

		log.Info().
			Str("source", cm.Src.String()).
			Str("destination", cm.Dst.String()).
			Int("ttl", cm.TTL).
			Int("bytes", n).
			Msgf("Received packet")

		if *dumpFlag {
			fmt.Printf("%s", hex.Dump(buf[:n]))
		}
	}
}
