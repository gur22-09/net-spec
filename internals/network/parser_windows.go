package network

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"unsafe"

	"github.com/gur22-09/net-spec/internals/utils"
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

func ParseTCPv4ConnectionsV2(buffer *[]byte) ([]Connection, error) {
	if len(*buffer) < 4 {
		return nil, errors.New("insufficient buffer size")
	}

	count := binary.LittleEndian.Uint32((*buffer)[:4])
	connections := make([]Connection, 0, count)

	// Correct entry size - should be 24 bytes for MIB_TCPROW_OWNER_PID
	const entrySize = 24

	// Verify buffer has enough data
	requiredSize := 4 + int(count)*entrySize
	if len(*buffer) < requiredSize {
		return nil, errors.New("buffer too small for claimed entry count")
	}

	// Slice the buffer to just the table data
	tableData := (*buffer)[4:requiredSize]

	for i := 0; i < int(count); i++ {
		offset := i * entrySize
		if offset+entrySize > len(tableData) {
			break // prevent out-of-bounds
		}

		rowData := tableData[offset : offset+entrySize]

		// Parse using binary.Read for stability
		var row MIB_TCPROW_OWNER_PID
		buf := bytes.NewReader(rowData)
		err := binary.Read(buf, binary.LittleEndian, &row)
		if err != nil {
			return nil, fmt.Errorf("failed to parse row %d: %v", i, err)
		}

		connection := Connection{
			LocalAddress:  utils.IpFromUint32(row.LocalAddr),
			LocalPort:     utils.PortFromDWORD(row.LocalPort),
			RemoteAddress: utils.IpFromUint32(row.RemoteAddr),
			RemotePort:    utils.PortFromDWORD(row.RemotePort),
			State:         utils.TcpStateToStr(row.State),
			PID:           row.OwningPid,
			Process:       utils.ResolvePID(row.OwningPid),
		}

		connections = append(connections, connection)
	}

	return connections, nil
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
		}

		connections = append(connections, connection)
	}

	return connections, nil
}
