package network

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/gur22-09/net-spec/internals/constants"
	"github.com/gur22-09/net-spec/internals/utils"
	"golang.org/x/sys/windows"
)

type MIB_TCPROW_OWNER_PID struct {
	State      uint32
	LocalAddr  uint32
	LocalPort  uint32
	RemoteAddr uint32
	RemotePort uint32
	OwningPid  uint32
}

type MIB_TCPTABLE_OWNER_PID struct {
	NumEntries uint32
	Table      []MIB_TCPROW_OWNER_PID
}

func ParseTCPv4Connections(buffer *[]byte) ([]Connection, error) {
	buff := *buffer
	if len(buff) < 4 {
		return nil, errors.New("Insufficient buffer size")
	}

	count := *(*uint32)(unsafe.Pointer(&buff[0]))

	connections := make([]Connection, 0, count)

	entrySize := int(unsafe.Sizeof(MIB_TCPROW_OWNER_PID{}))

	base := unsafe.Pointer(&buff[4])

	for i := 0; i < int(count); i++ {
		entryPtr := uintptr(base) + uintptr(i*entrySize)
		row := (*MIB_TCPROW_OWNER_PID)(unsafe.Pointer(entryPtr))

		if len(buff) < 4+(int(count)*entrySize) {
			return nil, errors.New("buffer too small for expected entry count")
		}

		connection := Connection{
			LocalAddress:  utils.IpFromUint32(row.LocalAddr),
			LocalPort:     utils.PortFromDWORD(row.LocalPort),
			RemoteAddress: utils.IpFromUint32(row.RemoteAddr),
			RemotePort:    utils.PortFromDWORD(row.RemotePort),
			State:         utils.TcpStateToStr(row.State),
			PID:           row.OwningPid,
			Process:       utils.ResolvePID(row.OwningPid),
			Protocol:      constants.ProtocolTCP,
		}

		connections = append(connections, connection)
	}

	return connections, nil
}

func GetProcessInfo(pid uint32) (ProcessInfo, error) {
	var info ProcessInfo
	info.PID = pid

	// Open process handle
	h, err := windows.OpenProcess(
		windows.PROCESS_QUERY_LIMITED_INFORMATION,
		false,
		pid,
	)
	if err != nil {
		return info, fmt.Errorf("OpenProcess failed: %v", err)
	}
	defer windows.CloseHandle(h)

	// Get executable path
	exePath := make([]uint16, windows.MAX_PATH)
	size := uint32(windows.MAX_PATH)
	err = windows.QueryFullProcessImageName(h, 0, &exePath[0], &size)
	if err != nil {
		return info, err
	}

	info.ExePath = syscall.UTF16ToString(exePath)
	info.Name = filepath.Base(info.ExePath)

	return info, nil
}

func printActiveConnections(connections []Connection, logger *log.Logger) {
	utils.ClearScreen(logger)

	fmt.Printf("Active TCP Connections [%d] @ %s\n", len(connections), time.Now().Format("15:04:05"))
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("%-21s %-6s  %-21s %-6s  %-12s  %s\n",
		"Local Address", "Local Port", "Remote Address", "RemotePort", "State", "Process")

	for _, c := range connections {
		fmt.Printf("%-21s %-6d  %-21s %-6d  %-12s  %s\n",
			c.LocalAddress, c.LocalPort,
			c.RemoteAddress, c.RemotePort,
			c.State, c.Process)
	}
}
