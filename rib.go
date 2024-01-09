package main

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/81ueman/local-clos/message/update"
)

type RibAdjEntry struct {
	ORIGIN           update.Origin
	AS_PATH          update.AS_PATH
	NEXT_HOP         update.NEXT_HOP
	LOCAL_PREF       update.LOCAL_PREF
	ATOMIC_AGGREGATE update.ATOMIC_AGGREGATE
}

type RibAdj map[netip.Prefix]RibAdjEntry

func (R *RibAdj) Update(msg update.Update) {
	for _, prefix := range msg.WithdrawnRoutes {
		delete(*R, prefix)
	}
	entry := RibAdjEntry{
		ORIGIN:     msg.PathAttrOrigin,
		AS_PATH:    msg.PathAttrASPath,
		NEXT_HOP:   msg.PathAttrNextHop,
		LOCAL_PREF: msg.PathAttrLocalPref,
	}
	for _, prefix := range msg.NetworkLayerReachabilityInformation {
		_, ok := (*R)[prefix]
		if !ok || ok && !reflect.DeepEqual((*R)[prefix], entry) {
			(*R)[prefix] = entry
		}
	}
}

// prefix s.t.
// prefix in R && prefix not in other
// or
// prefix in R && prefix in other && R[prefix] != other[prefix]
func (R *RibAdj) diff(other RibAdj) (RibAdj, []netip.Prefix) {
	log.Printf("compare %v and %v", R, other)
	diff := make(RibAdj)
	deleteroute := make([]netip.Prefix, 0)
	for prefix, entry := range *R {
		otherEntry, ok := other[prefix]
		if !ok {
			diff[prefix] = entry
		}
		if ok && !reflect.DeepEqual(entry, otherEntry) {
			diff[prefix] = entry
		}
	}
	for prefix := range other {
		_, ok := (*R)[prefix]
		if !ok {
			deleteroute = append(deleteroute, prefix)
		}
	}
	return diff, deleteroute
}

func (R *RibAdj) ToUpdateMsg(adjRibOut RibAdj) []update.Update {
	ribdiff, deleteroute := R.diff(adjRibOut)
	log.Printf("ribdiff: %v", ribdiff)
	msgs := make([]update.Update, 0)
	for prefix, entry := range ribdiff {
		msg := update.Update{
			NetworkLayerReachabilityInformation: []netip.Prefix{prefix},
			PathAttrOrigin:                      entry.ORIGIN,
			PathAttrASPath:                      entry.AS_PATH,
			PathAttrNextHop:                     entry.NEXT_HOP,
			PathAttrLocalPref:                   entry.LOCAL_PREF,
		}
		msgs = append(msgs, msg)
	}
	if len(deleteroute) != 0 {
		deletemsg := update.Update{
			WithdrawnRoutes: deleteroute,
		}
		msgs = append(msgs, deletemsg)
	}
	return msgs
}

func AdjFromLocal(AS uint16) (RibAdj, error) {
	ifis, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	adjBest := make(RibAdj)
	for _, ifi := range ifis {
		if is_loopback(ifi) {
			continue
		}
		prefix, err := IfiToPrefix(ifi)
		if err != nil {
			log.Printf("failed to get prefix: %v", err)
			continue
		}
		prefix = prefix.Masked()
		netipIP, err := localNetipIp(ifi)
		if err != nil {
			log.Printf("failed to get local netip ip: %v", err)
			continue
		}

		adjBest[prefix] = RibAdjEntry{
			ORIGIN: update.OriginIGP,
			AS_PATH: update.AS_PATH{
				VALUE_SEGMENT: update.VALUE_SEGMENT_AS_SEQUENCE,
				AS_SEQUENCE:   []uint16{AS},
			},
			NEXT_HOP:   update.NEXT_HOP(netipIP),
			LOCAL_PREF: update.LOCAL_PREF(100),
		}
	}
	return adjBest, nil
}
func (R *RibAdj) String() string {
	s := ""
	for prefix, entry := range *R {
		s += fmt.Sprintf("%s: %v\n", prefix.String(), entry)
	}
	return s
}

type Peer struct {
	RibAdjIn   RibAdj
	RibAdjInCh <-chan RibAdj
	LocRibCh   chan<- RibAdj
}

type LocRib struct {
	adjBest      RibAdj
	adjConnected RibAdj
	peers        []Peer
}

// compare two RibAdjEntry and return best one
func betterEntry(a, b RibAdjEntry) RibAdjEntry {
	if a.LOCAL_PREF > b.LOCAL_PREF {
		return a
	} else if a.LOCAL_PREF < b.LOCAL_PREF {
		return b
	}
	if len(a.AS_PATH.AS_SEQUENCE) < len(b.AS_PATH.AS_SEQUENCE) {
		return a
	} else if len(a.AS_PATH.AS_SEQUENCE) > len(b.AS_PATH.AS_SEQUENCE) {
		return b
	}
	if a.ORIGIN < b.ORIGIN {
		return a
	} else if a.ORIGIN > b.ORIGIN {
		return b
	}
	// TODO: まだいろいろ
	return a
}

func (l *LocRib) updateBestPath() {
	l.adjBest = l.adjConnected
	for _, peer := range l.peers {
		for prefix, entry := range peer.RibAdjIn {
			_, ok := l.adjBest[prefix]
			if !ok {
				l.adjBest[prefix] = entry
			} else {
				l.adjBest[prefix] = betterEntry(l.adjBest[prefix], entry)
			}
		}
	}
}

func (L *LocRib) Handle() {
	cases := make([]reflect.SelectCase, len(L.peers))
	for i, peer := range L.peers {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(peer.RibAdjInCh)}
	}
	chosen, value, ok := reflect.Select(cases)
	log.Printf("chosen: %v, value: %v, ok: %v", chosen, value, ok)
	if !ok {
		log.Printf("reflect.Select failed: %v", ok)
		return
	}
	L.peers[chosen].RibAdjIn = value.Interface().(RibAdj)
	L.updateBestPath()
	log.Printf("updated adjBest: %v", L.adjBest)
	for _, peer := range L.peers {
		peer.LocRibCh <- L.adjBest
	}
}

func (L *LocRib) Sig() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP)
	for {
		<-sig
		log.Println("SIGHUP received")
		log.Println("Print adjBest")
		log.Print(L.adjBest)
	}

}
