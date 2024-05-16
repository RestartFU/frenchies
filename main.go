package main

import (
	"frenchies/frenchies/session"
	"github.com/pelletier/go-toml"
	hook "github.com/robotn/gohook"
	"github.com/sandertv/gophertunnel/minecraft"
	"log"
	"os"
)

// The following program implements a proxy that forwards players from one local address to a remote address.
func main() {
	s, err := session.NewSession()
	if err != nil {
		panic(err)
	}
	config := readConfig()

	p, err := minecraft.NewForeignStatusProvider(config.Connection.RemoteAddress)
	if err != nil {
		panic(err)
	}

	serverConn, err := minecraft.Dialer{
		TokenSource: s.Token(),
	}.Dial("raknet", config.Connection.RemoteAddress)
	if err != nil {
		panic(err)
	}

	listener, err := minecraft.ListenConfig{
		ResourcePacks:  serverConn.ResourcePacks(),
		StatusProvider: p,
	}.Listen("raknet", config.Connection.LocalAddress)
	if err != nil {
		panic(err)
	}
	serverConn.Close()
	defer listener.Close()
	for {
		c, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleConn(c.(*minecraft.Conn), listener, config, s)
	}
}

// handleConn handles a new incoming minecraft.Conn from the minecraft.Listener passed.
func handleConn(conn *minecraft.Conn, listener *minecraft.Listener, config config, s *session.Session) {
	serverConn, err := s.Dial(config.Connection.RemoteAddress, conn)
	if err != nil {
		panic(err)
	}
	s.ExecuteCommand("/joinqueue hcf")

	go func() {
		hook.Register(hook.MouseDown, []string{}, func(e hook.Event) {
			if e.Button == 5 {
				s.ExecuteCommand("/tl")
			} else if e.Button == 4 {
				s.ExecuteCommand("/kit")
			} else if e.Button == 3 {
				s.ToggleInvisibility()
			}
		})

		s := hook.Start()
		<-hook.Process(s)
	}()
	_ = serverConn
}

type config struct {
	Connection struct {
		LocalAddress  string
		RemoteAddress string
	}
}

func readConfig() config {
	c := config{}
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		f, err := os.Create("config.toml")
		if err != nil {
			log.Fatalf("error creating config: %v", err)
		}
		data, err := toml.Marshal(c)
		if err != nil {
			log.Fatalf("error encoding default config: %v", err)
		}
		if _, err := f.Write(data); err != nil {
			log.Fatalf("error writing encoded default config: %v", err)
		}
		_ = f.Close()
	}
	data, err := os.ReadFile("config.toml")
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		log.Fatalf("error decoding config: %v", err)
	}
	if c.Connection.LocalAddress == "" {
		c.Connection.LocalAddress = "0.0.0.0:19132"
	}
	data, _ = toml.Marshal(c)
	if err := os.WriteFile("config.toml", data, 0644); err != nil {
		log.Fatalf("error writing config file: %v", err)
	}
	return c
}
