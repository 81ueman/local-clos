package main

import (
	"log"
	"net/netip"
	"reflect"

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

type Peer struct {
	RibAdjIn   RibAdj
	RibAdjInCh <-chan RibAdj
	LocRibCh   chan<- RibAdj
}

type LocRib struct {
	adjBest RibAdj
	peers   []Peer
}

// compare two RibAdjEntry and return best one
func compareRibAdjEntry(a, b RibAdjEntry) RibAdjEntry {
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

func (l *LocRib) selectBestPath() {
	l.adjBest = make(RibAdj)
	for _, peer := range l.peers {
		for prefix, entry := range peer.RibAdjIn {
			_, ok := l.adjBest[prefix]
			if !ok {
				l.adjBest[prefix] = entry
			} else {
				l.adjBest[prefix] = compareRibAdjEntry(l.adjBest[prefix], entry)
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
		return
	}
	L.peers[chosen].RibAdjIn = value.Interface().(RibAdj)
	L.selectBestPath()
	for _, peer := range L.peers {
		peer.LocRibCh <- L.adjBest
	}
}
