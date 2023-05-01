package main

import (
	"encoding/binary"
	"errors"
	"net"
	"sync/atomic"
)

type Netwalker struct {
	Start   uint32
	End     uint32
	Mask    uint32
	Current uint32
}

func CreateNetwalker(netBlock *net.IPNet) *Netwalker {

	nw := Netwalker{
		Start: binary.BigEndian.Uint32(netBlock.IP),
		Mask:  binary.BigEndian.Uint32(netBlock.Mask),
	}

	nw.End = (nw.Start & nw.Mask) | (nw.Mask ^ 0xffffffff)
	nw.Current = nw.Start

	return &nw
}

func (nw *Netwalker) IncludeIfCan(nb *net.IPNet) bool {

	other := CreateNetwalker(nb)

	if int64(nw.Start)-int64(other.End) > 1 || int64(other.Start)-int64(nw.End) > 1 {
		return false
	}

	if other.Start < nw.Start {
		nw.Start = other.Start
	}

	if other.End > nw.End {
		nw.End = other.End
	}

	return true
}

func (nw *Netwalker) CurrentAddress() net.IP {

	return Uint32ToIp(nw.Current)
}

func Uint32ToIp(uint uint32) net.IP {

	ip := make([]byte, 4)
	binary.BigEndian.PutUint32(ip, uint)

	return ip
}

func (nw *Netwalker) HaveNextAddress() bool {

	return nw.Current < nw.End
}

func (nw *Netwalker) GoNextAddress() error {

	if nw.Current >= nw.End {
		return errors.New("address is out of range")
	}

	atomic.AddUint32(&nw.Current, 1)
	return nil
}

func (nw *Netwalker) AddressRange() uint32 {

	return (nw.End - nw.Start) + 1
}
