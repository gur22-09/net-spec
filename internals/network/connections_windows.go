package network

import (
	"log"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type tcpTableClass int32

type ProcessInfo struct {
	Name    string
	PID     uint32
	ExePath string
}

type Connection struct {
	LocalAddress  string
	LocalPort     uint16
	RemoteAddress string
	RemotePort    uint16
	State         string
	PID           uint32
	Process       string
	Protocol      string
}

type NetworkHandler struct {
	logger *log.Logger
}

const TCP_TABLE_OWNER_PID_ALL = 5

var (
	iphlpapi                = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetExtendedTCPTable = iphlpapi.NewProc("GetExtendedTcpTable")
)

func NewNetworkHandler(logger *log.Logger) *NetworkHandler {
	return &NetworkHandler{
		logger: logger,
	}
}

func getUintptrFromBool(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}

// https://learn.microsoft.com/en-us/windows/win32/api/iphlpapi/nf-iphlpapi-getextendedtcptable
func getExtendedTCPTable(pTCPTablePtr uintptr, pdwSize *uint32, bOrder bool, ulAf uint32, tableClass tcpTableClass, reserved uint32) (errcode error) {
	r1, _, _ := syscall.Syscall6(procGetExtendedTCPTable.Addr(), 6, pTCPTablePtr, uintptr(unsafe.Pointer(pdwSize)), getUintptrFromBool(bOrder), uintptr(ulAf), uintptr(tableClass), uintptr(reserved))
	if r1 != 0 {
		errcode = syscall.Errno(r1)
	}
	return
}

func (n *NetworkHandler) fetchTCPTables(af uint32, tableClass tcpTableClass, bOrder bool) (*[]byte, error) {
	var size uint32

	// get the buffer size as
	// number of connections varies to allocate
	n.logger.Printf("Calling fetchTCPTables with: af=%d, class=%d", af, tableClass)

	err := getExtendedTCPTable(
		0,
		&size,
		false,
		af,
		tableClass,
		0,
	)

	if err != syscall.ERROR_INSUFFICIENT_BUFFER {
		n.logger.Println("Error: getting buffer size for getExtendedTCPTable in GetTCPConnections")
		return nil, err
	}

	// allocate buffer
	buffer := make([]byte, size)

	// call to get the data
	err = getExtendedTCPTable(uintptr(unsafe.Pointer(&buffer[0])), &size, bOrder, af, tableClass, 0)

	if err != nil {
		return nil, err
	}

	return &buffer, nil
}

func (n *NetworkHandler) getTCPConnections() ([]Connection, error) {
	var connections []Connection

	ipv4Buf, err := n.fetchTCPTables(syscall.AF_INET, TCP_TABLE_OWNER_PID_ALL, false)

	if err != nil {
		return nil, err
	}

	v4Conns, err := ParseTCPv4Connections(ipv4Buf)

	if err != nil {
		n.logger.Printf("failed to get ipv4 connections: %v\n", err)
		return nil, err
	}

	connections = append(connections, v4Conns...)

	ipv6Buf, err := n.fetchTCPTables(syscall.AF_INET6, TCP_TABLE_OWNER_PID_ALL, true)

	if err != nil {
		return nil, err
	}

	v6Conns, err := ParseTCPv6Connections(ipv6Buf)

	if err != nil {
		n.logger.Printf("failed to get ipv6 connections: %v\n", err)
		return nil, err
	}

	connections = append(connections, v6Conns...)

	return connections, nil
}

func (n *NetworkHandler) GetAllConnections() ([]Connection, error) {
	var connections []Connection

	tcp, err := n.getTCPConnections()

	if err != nil {
		return connections, err
	}

	connections = append(connections, tcp...)

	return connections, nil
}
