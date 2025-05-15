package utils

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"unsafe"

	"golang.org/x/sys/windows"
)

func IpFromUint32(ip uint32) string {
	return net.IPv4(
		byte(ip),
		byte(ip>>8),
		byte(ip>>16),
		byte(ip>>24),
	).String()
}

func PortFromDWORD(dword uint32) uint16 {
	// ports are stored in network byte order
	return binary.BigEndian.Uint16((*[2]byte)(unsafe.Pointer(&dword))[:])
}

func TcpStateToStr(state uint32) string {
	fmt.Println("parsing tcp state", state)
	switch state {
	case 1:
		return "CLOSED"
	case 2:
		return "LISTEN"
	case 3:
		return "SYN_SENT"
	case 4:
		return "SYN_RECEIVED"
	case 5:
		return "ESTABLISHED"
	case 6:
		return "FIN_WAIT_1"
	case 7:
		return "FIN_WAIT_2"
	case 8:
		return "CLOSE_WAIT"
	case 9:
		return "CLOSING"
	case 10:
		return "LAST_ACK"
	case 11:
		return "TIME_WAIT"
	case 12:
		return "DELETE_TCB"
	default:
		return "UNKNOWN"
	}
}

func ResolvePID(pid uint32) string {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(h)

	var buf [windows.MAX_PATH]uint16
	size := uint32(len(buf))
	err = windows.QueryFullProcessImageName(h, 0, &buf[0], &size)
	if err != nil {
		return ""
	}
	return windows.UTF16ToString(buf[:])
}

func ClearScreen(logger *log.Logger) {
	logger.Print("\033[H\033[2J")
}
