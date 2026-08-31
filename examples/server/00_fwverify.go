package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

// probeTimeout bounds a single TCP connection attempt. Short, because a
// blocked port shows up as a timeout and a slow one is indistinguishable
// from an open one only if you wait forever.
const probeTimeout = 5 * time.Second

// probeSettle is how long to keep re-probing before concluding a
// firewall change did not take effect.
//
// The API's job completing means the platform has applied the rule; it
// does not promise the packet filter has converged. Concluding "the
// firewall does nothing" from a single probe a second later would be
// wrong, and is exactly the kind of confident-but-false result this
// journey is meant to avoid.
const probeSettle = 90 * time.Second

// tcpReachable reports whether a TCP connection to host:port completes.
//
// This is the only check in the whole journey that does not ask the API
// about itself. Everything else asserts that the control plane agrees
// with the control plane: attach a group, then ask the same API whether
// the group is attached. That proves the record changed, not that a
// packet was filtered — and those are different claims. This one opens
// a socket.
func tcpReachable(host, port string) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), probeTimeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// waitReachability polls until the port reaches the wanted state, or
// gives up.
//
// Returns whether it got there, and how long it took, so the caller can
// report the convergence time rather than implying it was instant.
func waitReachability(host, port string, want bool) (bool, time.Duration) {
	start := time.Now()
	deadline := start.Add(probeSettle)
	for {
		if tcpReachable(host, port) == want {
			return true, time.Since(start)
		}
		if time.Now().After(deadline) {
			return false, time.Since(start)
		}
		time.Sleep(3 * time.Second)
	}
}

// assertReachable checks the port is open, from outside.
func assertReachable(host, port, why string) error {
	ok, took := waitReachability(host, port, true)
	if !ok {
		return fmt.Errorf("%s:%s is not reachable after %s — %s", host, port, took.Round(time.Second), why)
	}
	log.Printf("  ✓ %s:%s accepts connections (%s)", host, port, took.Round(time.Second))
	return nil
}

// assertBlocked checks the port is closed, from outside.
//
// This is the assertion that makes the firewall step mean something. A
// security group that the API reports as attached, on a server whose
// port 22 still answers, is a firewall that is not doing anything — and
// no amount of asking the API would reveal it.
func assertBlocked(host, port, why string) error {
	ok, took := waitReachability(host, port, false)
	if !ok {
		return fmt.Errorf("%s:%s still accepts connections after %s — %s. The control plane reported the rule applied; the packet filter disagrees",
			host, port, took.Round(time.Second), why)
	}
	log.Printf("  ✓ %s:%s refuses connections (%s)", host, port, took.Round(time.Second))
	return nil
}
