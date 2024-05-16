package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"strings"
)

func (s *Session) ToggleInvisibility() {
	s.invisibility = !s.invisibility
}

func (s *Session) SetActorData(ctx *event.Context, pk packet.Packet) {
	pkt, ok := pk.(*packet.SetActorData)
	if !ok {
		return
	}
	meta := protocol.EntityMetadata(pkt.EntityMetadata)
	_, ok = meta[protocol.EntityDataKeyFlags]
	if meta == nil || !ok {
		return
	}

	if !s.invisibility {
		tag := meta[protocol.EntityDataKeyName]
		var name string
		if tag != nil {
			name = tag.(string)
			if strings.Contains(name, "sheely4") {
				name = strings.ReplaceAll(name, "sheely4", "Restart")
			}
			meta[protocol.EntityDataKeyName] = name
		}
		if meta.Flag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagInvisible) {
			if tag != nil {
				meta[protocol.EntityDataKeyName] = fmt.Sprintf("§7[I] %s", name)
			}
			removeFlag(protocol.EntityDataKeyFlags, protocol.EntityDataFlagInvisible, meta)
		}
		pkt.EntityMetadata = meta
	}
}

// removeFlag removes a flag from the entity data.
func removeFlag(key uint32, index uint8, m protocol.EntityMetadata) {
	v, ok := m[key]
	if !ok {
		return
	}
	switch key {
	case protocol.EntityDataKeyPlayerFlags:
		m[key] = v.(byte) &^ (1 << index)
	default:
		m[key] = v.(int64) &^ (1 << int64(index))
	}
}
