package main

// IPMI implementation for Go
//
// Based on https://www-ssl.intel.com/content/www/us/en/servers/ipmi/ipmi-intelligent-platform-mgt-interface-spec-2nd-gen-v2-0-spec-update.html

import (
	"context"
	"net"
	"time"
)

type lanConnection struct {
	conn net.Conn
}

func newLanConnection(host string) (lanConnection, error) {
	l := lanConnection{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	dialer := &net.Dialer{}
	if conn, err := dialer.DialContext(ctx, "udp4", host); err != nil {
		return l, err
	} else {
		l.conn = conn
	}

	// FIXME: Move to send / recv functions
	deadline, _ := ctx.Deadline()
	if err := l.conn.SetDeadline(deadline); err != nil {
		panic(err)
	}

	return l, nil
}

func (l *lanConnection) close() {
	l.conn.Close()
}
