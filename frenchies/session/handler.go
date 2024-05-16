package session

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type handlerFunc func(ctx *event.Context, pk packet.Packet)
