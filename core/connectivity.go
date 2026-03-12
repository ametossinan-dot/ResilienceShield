package core

import (
	"net"
	"resilienceshield/logs"
	"sync"
	"time"
)

type ConnectivityChecker struct {
	remoteHost     string
	isOnline       bool
	mu             sync.RWMutex
	checkInterval  time.Duration
	OnStatusChange func(online bool)
}

func NewConnectivityChecker(host string, interval int) *ConnectivityChecker {
	return &ConnectivityChecker{
		remoteHost:    host,
		isOnline:      false,
		checkInterval: time.Duration(interval) * time.Second,
	}
}

func (c *ConnectivityChecker) check() bool {
	conn, err := net.DialTimeout("tcp", c.remoteHost, 3*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (c *ConnectivityChecker) IsOnline() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isOnline
}

func (c *ConnectivityChecker) Start() {
	go func() {
		for {
			online := c.check()
			c.mu.Lock()
			previous := c.isOnline
			c.isOnline = online
			c.mu.Unlock()

			if online != previous {
				if online {
					logs.Success("Réseau rétabli — passage en mode ONLINE")
				} else {
					logs.Warning("Réseau perdu — passage en mode OFFLINE")
				}
				if c.OnStatusChange != nil {
					c.OnStatusChange(online)
				}
			}
			time.Sleep(c.checkInterval)
		}
	}()
}
