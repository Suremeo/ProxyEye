/*
 *
 *           ____                        ______
 *          / __ \_________  _  ____  __/ ____/_  _____
 *         / /_/ / ___/ __ \| |/_/ / / / __/ / / / / _ \
 *        / ____/ /  / /_/ />  </ /_/ / /___/ /_/ /  __/
 *       /_/   /_/   \____/_/|_|\__, /_____/\__, /\___/
 *                                /_/         /_/
 *       ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀
 *
 *                     Author: Suremeo (github.com/Suremeo)
 *
 *
 */

package proxy

import (
	"github.com/Suremeo/ProxyEye/proxy/session"
	"github.com/Suremeo/ProxyEye/proxy/session/events"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"go.uber.org/atomic"
	"strings"
	"sync"
	"time"
)

type eyePlayer struct {
	mutex        sync.Mutex
	sessionMutex sync.Mutex
	close        sync.Once
	conn         *minecraft.Conn
	online       atomic.Bool
	transferring atomic.Bool
	session      session.Session
	anticheat    session.AntiCheatProfile
	chunks       session.Chunks
}

func newPlayer(c *minecraft.Conn) session.Player {
	p := &eyePlayer{
		conn: c,
	}
	p.online.Store(true)
	p.chunks = newChunks(p, c.ChunkRadius(), nil)
	p.anticheat = newAntiCheatProfile(p)
	var to session.Session
	go func() {
		for {
			pk, err := c.ReadPacket()
			if err != nil {
				p.Close()
				return
			}
			if p.Session().Server().Packet(p, p, pk) {
				if p.Session().IsConnected() {
					to = p.Session()
				}
				err = to.Packet(p, p.Session(), pk)
				if err != nil {
					to.Close()
				}
			}
		}
	}()
	return p
}

func (p *eyePlayer) Chunks() session.Chunks {
	return p.chunks
}

func (p *eyePlayer) Close() {
	p.online.Store(false)
	p.close.Do(func() {
		Eye.ExecuteEvent(events.EventLeave, p)
		s := p.Session()
		if s != nil {
			s.Close()
		}
		if p.anticheat != nil {
			p.anticheat.Close()
		}
		_ = p.conn.Close()
		p.chunks.Close()
	})
}

func (p *eyePlayer) OldRuntimeId() uint64 {
	return 1
}

func (p *eyePlayer) NewRuntimeId() uint64 {
	s := p.Session()
	if s == nil {
		return 1
	}
	return s.OldRuntimeId()
}

func (p *eyePlayer) OldUniqueId() int64 {
	return 1
}

func (p *eyePlayer) NewUniqueId() int64 {
	s := p.Session()
	if s == nil {
		return 1
	}
	return s.OldUniqueId()
}

func (p *eyePlayer) TranslateEid(now uint64) uint64 {
	if p.OldRuntimeId() == now {
		return p.NewRuntimeId()
	} else if p.NewRuntimeId() == now {
		return p.OldRuntimeId()
	}
	return now
}

func (p *eyePlayer) TranslateUid(now int64) int64 {
	if p.OldUniqueId() == now {
		return p.NewUniqueId()
	} else if p.NewUniqueId() == now {
		return p.OldUniqueId()
	}
	return now
}

func (p *eyePlayer) WritePacket(packet packet.Packet) error {
	return p.conn.WritePacket(packet)
}

func (p *eyePlayer) Online() bool {
	return p.online.Load()
}

func (p *eyePlayer) Session() session.Session {
	p.sessionMutex.Lock()
	s := p.session
	p.sessionMutex.Unlock()
	if s == nil {
		s = NopSession
	}
	return s
}

func (p *eyePlayer) SetSession(session session.Session) {
	p.sessionMutex.Lock()
	p.session = session
	p.sessionMutex.Unlock()
}

func (p *eyePlayer) Raknet() *minecraft.Conn {
	return p.conn
}

func (p *eyePlayer) Type() session.SourceType {
	return session.SourceTypePlayer
}

func (p *eyePlayer) Anticheat() session.AntiCheatProfile {
	return p.anticheat
}

func (p *eyePlayer) Chat(msg string, format ...interface{}) {
	_ = p.WritePacket(&packet.Text{Message: text.Colourf(msg, format...)})
}

func (p *eyePlayer) Kick(reason ...string) {
	if len(reason) == 0 {
		reason = []string{text.Colourf("<red>No reason provided</red>")}
	}
	_ = p.WritePacket(&packet.Disconnect{Message: text.Colourf(strings.Join(reason, "\n"))})
	time.Sleep(1 * time.Second)
	p.Close()
}

func (p *eyePlayer) IsTransferring() bool {
	return p.transferring.Load()
}

func (p *eyePlayer) SetTransferring(b bool) {
	p.transferring.Store(b)
}

func (p *eyePlayer) Sound(name string, volume, pitch float32) {
	_ = p.WritePacket(&packet.PlaySound{
		SoundName: name,
		Position:  p.Anticheat().Position().Sub(mgl32.Vec3{0, 25, 0}),
		Volume:    volume,
		Pitch:     pitch,
	})
}
