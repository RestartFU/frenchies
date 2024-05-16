package session

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func (s *Session) MobEffect(ctx *event.Context, pk packet.Packet) {
	pkt, ok := pk.(*packet.MobEffect)
	if !ok {
		return
	}

	if pkt.Operation == packet.MobEffectAdd {
		if pkt.EffectType == 14 {
			ctx.Cancel()
			return
		}
	}
}
