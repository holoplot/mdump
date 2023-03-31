# Simple multicast packet dump

Written in pure Go.

## Run it

Dump all mDNS multicast packets on the Wifi interface:

```
$ go run ./cmd/mdump/... -i wlp2s0 -g 224.0.0.251 -p 5353 -x
03:45AM INF Received packet bytes=40 destination=224.0.0.251 source=192.168.178.116 ttl=255
00000000  00 00 00 00 00 01 00 00  00 00 00 00 0b 5f 67 6f  |............._go|
00000010  6f 67 6c 65 7a 6f 6e 65  04 5f 74 63 70 05 6c 6f  |oglezone._tcp.lo|
00000020  63 61 6c 00 00 0c 00 01                           |cal.....|
```

## License
MIT