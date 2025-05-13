package network

import (
	"log"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type tcpTableClass int32

type Connection struct {
	LocalAddress  string
	LocalPort     uint16
	RemoteAddress string
	RemotePort    uint16
	State         string
	PID           uint32
	Process       string
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
		n.logger.Printf("failed to get tcp connections: %v", err)
		return nil, err
	}

	connections = append(connections, v4Conns...)

	// TODO - add ipv6 connections

	return connections, nil
}

func (n *NetworkHandler) SetupNetworkListener() error {
	connections, err := n.getTCPConnections()

	if err != nil {
		n.logger.Println("failed to get tcp connections")
		return err
	}

	for _, conn := range connections {
		n.logger.Printf("Local Address %s, Local Port %d -> Remote Address %s: Remote Port %d, Connection State %s, PID=%d with Process %s\n",
			conn.LocalAddress, conn.LocalPort, conn.RemoteAddress, conn.RemotePort, conn.State, conn.PID, conn.Process,
		)
	}

	return nil
}
