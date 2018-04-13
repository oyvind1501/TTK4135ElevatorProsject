

package conn

import (
	"C"
	"errors"
	"fmt"
	"net"
	"time"
	"unsafe"
)

type WindowsBroadcastConn struct {
	Sock C.SOCKET
}

func (f WindowsBroadcastConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	var addrbuf [16]byte
	r := int(C.cRecvFrom(f.Sock, (*C.char)(unsafe.Pointer(&b[0])), C.int(len(b)), (*C.char)(unsafe.Pointer(&addrbuf[0]))))
	if r == C.SOCKET_ERROR {
		return 0, nil, errors.New(fmt.Sprintf("recvfrom() failed with error code %d", C.WSAGetLastError()))
	} else {
		addr := net.UDPAddr{IP: addrbuf[4:8], Port: int(C.ntohs(C.u_short(addrbuf[2]<<8) | C.u_short(addrbuf[3])))}
		return r, &addr, nil
	}
}

func (f WindowsBroadcastConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	r := int(C.cSendTo(f.Sock, C.CString(addr.(*net.UDPAddr).IP.String()), C.u_short(addr.(*net.UDPAddr).Port), (*C.char)(unsafe.Pointer(&b[0])), C.int(len(b))))
	if r == C.SOCKET_ERROR {
		return 0, errors.New(fmt.Sprintf("sendto() failed with error code %d", C.WSAGetLastError()))
	} else {
		return r, nil
	}
}

func (f WindowsBroadcastConn) Close() error {
	r := C.cClose(f.Sock)
	if r == 0 {
		return nil
	} else {
		return errors.New(fmt.Sprintf("closesocket() failed with error code %d", C.WSAGetLastError()))
	}
}

func (f WindowsBroadcastConn) LocalAddr() net.Addr {
	var addrbuf [16]byte
	r := int(C.cLocalAddr(f.Sock, (*C.char)(unsafe.Pointer(&addrbuf[0]))))
	if r == C.SOCKET_ERROR {
		return nil
	} else {
		addr := net.UDPAddr{IP: addrbuf[4:8], Port: int(C.ntohs(C.u_short(addrbuf[2]<<8) | C.u_short(addrbuf[3])))}
		return &addr
	}
}

func (f WindowsBroadcastConn) SetDeadline(t time.Time) error {
	e := f.SetReadDeadline(t)
	if e != nil {
		return e
	}
	e = f.SetWriteDeadline(t)
	return e
}

func (f WindowsBroadcastConn) SetReadDeadline(t time.Time) error {
	timeout_ms := int64(t.Sub(time.Now())) / 1000000
	r := -1
	if timeout_ms > 0 {
		r = int(C.cSetReadDeadline(f.Sock, C.int(timeout_ms)))
	}
	if r == 0 {
		return nil
	} else {
		return errors.New(fmt.Sprintf("setsockopt() failed with error code %d", C.WSAGetLastError()))
	}
}

func (f WindowsBroadcastConn) SetWriteDeadline(t time.Time) error {
	timeout_ms := int64(t.Sub(time.Now())) / 1000000
	r := -1
	if timeout_ms > 0 {
		r = int(C.cSetWriteDeadline(f.Sock, C.int(timeout_ms)))
	}
	if r == 0 {
		return nil
	} else {
		return errors.New(fmt.Sprintf("setsockopt() failed with error code %d", C.WSAGetLastError()))
	}
}

func DialBroadcastUDP(port int) net.PacketConn {
	return WindowsBroadcastConn{C.cBcastSocket(C.u_short(port))}
}
