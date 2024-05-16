package session

import (
	"context"
	"crypto/ecdsa"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/google/uuid"
	"github.com/restartfu/gophig"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/oauth2"
	"sync"
	_ "unsafe"
)

type Session struct {
	tok              oauth2.TokenSource
	handler          map[uint32]handlerFunc
	conn, serverConn *minecraft.Conn
	invisibility     bool
}

func NewSession() (*Session, error) {
	tok, err := resolveToken()
	if err != nil {
		return nil, err
	}
	s := &Session{
		tok: tok,
	}

	s.handler = map[uint32]handlerFunc{
		packet.IDSetActorData: s.SetActorData,
		packet.IDMobEffect:    s.MobEffect,
	}
	return s, nil
}

func (s *Session) Token() oauth2.TokenSource {
	return s.tok
}

func (s *Session) ExecuteCommand(command string) {
	_ = s.serverConn.WritePacket(&packet.CommandRequest{
		CommandLine: command,
	})
}

func (s *Session) Dial(addr string, conn *minecraft.Conn) (*minecraft.Conn, error) {
	clientData := conn.ClientData()
	formulateClientData(&clientData)

	serverConn, err := minecraft.Dialer{
		TokenSource: s.tok,
		ClientData:  clientData,
	}.Dial("raknet", addr)
	if err != nil {
		return nil, err
	}

	s.conn = conn
	s.serverConn = serverConn

	go s.handleConn()
	return serverConn, nil

}

func (s *Session) handleClientPackets() {
	for {
		pk, err := s.conn.ReadPacket()
		if err != nil {
			return
		}
		s.handleClientPacket(pk)
	}
}

func (s *Session) handleClientPacket(pk packet.Packet) {
	_ = s.serverConn.WritePacket(pk)
}

func (s *Session) handleServerPackets() {
	for {
		pk, err := s.serverConn.ReadPacket()
		if err != nil {
			return
		}
		s.handleServerPacket(pk)
	}
}

func (s *Session) handleServerPacket(pk packet.Packet) {
	ctx := event.C()
	if h, ok := s.handler[pk.ID()]; ok {
		h(ctx, pk)
	}
	if !ctx.Cancelled() {
		_ = s.conn.WritePacket(pk)
	}
}

func (s *Session) handleConn() {
	var g sync.WaitGroup
	g.Add(2)
	go func() {
		if err := s.conn.StartGame(s.serverConn.GameData()); err != nil {
			panic(err)
		}
		g.Done()
	}()
	go func() {
		if err := s.serverConn.DoSpawn(); err != nil {
			panic(err)
		}
		g.Done()
	}()
	g.Wait()

	go s.handleServerPackets()
	go s.handleClientPackets()
}

func formulateClientData(dat *login.ClientData) {
	dat.DeviceOS = protocol.DeviceWin10
	dat.DeviceID = uuid.New().String()
}

func resolveToken() (oauth2.TokenSource, error) {
	token := new(oauth2.Token)
	g := gophig.NewGophig("token", "json", 699)
	err := g.GetConf(token)
	if err != nil {
		token, err = auth.RequestLiveToken()
		if err != nil {
			return nil, err
		}
	}

	src := auth.RefreshTokenSource(token)
	_, err = src.Token()
	if err != nil {
		// The cached refresh token expired and can no longer be used to obtain a new token. We require the
		// user to log in again and use that token instead.
		token, err = auth.RequestLiveToken()
		src = auth.RefreshTokenSource(token)
	}

	_ = g.SetConf(token)
	return src, nil
}

// noinspection ALL
//
//go:linkname authChain github.com/sandertv/gophertunnel/minecraft.authChain
func authChain(ctx context.Context, src oauth2.TokenSource, key *ecdsa.PrivateKey) (string, error)

// noinspection ALL
//
//go:linkname readChainIdentityData github.com/sandertv/gophertunnel/minecraft.readChainIdentityData
func readChainIdentityData(chainData []byte) login.IdentityData

// noinspection ALL
//
//go:linkname defaultClientData github.com/sandertv/gophertunnel/minecraft.defaultClientData
func defaultClientData(address, username string, d *login.ClientData)

// noinspection ALL
//
//go:linkname defaultIdentityData github.com/sandertv/gophertunnel/minecraft.defaultIdentityData
func defaultIdentityData(data *login.IdentityData)
