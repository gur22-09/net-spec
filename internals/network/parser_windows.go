package network

import (
	"errors"
	"fmt"
	"path/filepath"
	"syscall"
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

type MIB_TP6ROW_OWNER_PID struct {
	LocalAddr     [16]byte
	LocalScopeId  uint32
	LocalPort     uint32
	RemoteAddr    [16]byte
	RemoteScopeId uint32
	RemotePort    uint32
	State         uint32
	OwningPid     uint32
}

func ParseTCPv6Connections(buffer *[]byte) ([]Connection, error) {
	buff := *buffer
	if len(buff) < 4 {
		return nil, errors.New("insufficient buffer size")
	}

	count := *(*uint32)(unsafe.Pointer(&buff[0]))

	fmt.Println("number of ipv6 connections", count)

	connections := make([]Connection, 0, count)

	entrySize := int(unsafe.Sizeof(MIB_TP6ROW_OWNER_PID{}))

	base := unsafe.Pointer(&buff[4])

	for i := 0; i < int(count); i++ {
		entryPtr := uintptr(base) + uintptr(i*entrySize)

		row := (*MIB_TP6ROW_OWNER_PID)(unsafe.Pointer(entryPtr))

		if len(buff) < 4+(int(count)*entrySize) {
			return nil, errors.New("buffer too small for expected entry count")
		}

		connection := Connection{
			LocalAddress:  utils.IpFrom16Bytes(row.LocalAddr, row.LocalScopeId),
			LocalPort:     utils.PortFromDWORD(row.LocalPort),
			RemoteAddress: utils.IpFrom16Bytes(row.RemoteAddr, row.RemoteScopeId),
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

func ParseTCPv4Connections(buffer *[]byte) ([]Connection, error) {
	buff := *buffer
	if len(buff) < 4 {
		return nil, errors.New("insufficient buffer size")
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
