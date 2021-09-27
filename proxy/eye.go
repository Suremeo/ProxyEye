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
	"github.com/Suremeo/ProxyEye/proxy/console"
	"github.com/Suremeo/ProxyEye/proxy/session"
	"github.com/Suremeo/ProxyEye/proxy/session/anticheat"
	"github.com/Suremeo/ProxyEye/proxy/session/events"
	"github.com/Suremeo/ProxyEye/proxy/storage"
	_ "github.com/Suremeo/ProxyEye/proxy/world/blocks"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"sync"
)

var Eye = initProxy()
var lobby = NewRemoteServer("127.0.0.1:1001", 100)

type eyeProxy struct {
	events *sync.Map
	mutex  sync.Mutex

	playerMutex sync.Mutex

	players     map[uuid.UUID]session.Player
	playersXuid map[string]session.Player
}

func initProxy() *eyeProxy {
	return &eyeProxy{
		events:      &sync.Map{},
		players:     make(map[uuid.UUID]session.Player),
		playersXuid: make(map[string]session.Player),
	}
}

func (eye *eyeProxy) Run() error {
	if !storage.Config.Listener.Auth {
		console.Warn("Xbox Authentication is disabled, anyone can be anyone so watch your back.")
	}
	l, err := minecraft.ListenConfig{
		ErrorLog:               discardLogger,
		StatusProvider:         minecraft.NewStatusProvider(text.Colourf(storage.Config.Listener.Motd)),
		AuthenticationDisabled: !storage.Config.Listener.Auth,
	}.Listen("raknet", storage.Config.Listener.Address)
	if err != nil {
		return err
	}
	console.Info("Listening for players on %v", storage.Config.Listener.Address)

	eye.HandleEvent(events.LeaveEventHandler(func(event *events.Context, player session.Player) {
		eye.playerMutex.Lock()
		delete(Eye.playersXuid, player.Raknet().IdentityData().XUID)
		delete(Eye.players, uuid.Must(uuid.Parse(player.Raknet().IdentityData().Identity)))
		eye.playerMutex.Unlock()
	}))

	for {
		c, err := l.Accept()
		if err != nil {
			continue
		}
		go func() {
			p := newPlayer(c.(*minecraft.Conn))
			eye.playerMutex.Lock()
			if storage.Config.Listener.Auth {
				eye.playersXuid[c.(*minecraft.Conn).IdentityData().XUID] = p
			}
			eye.players[uuid.Must(uuid.Parse(c.(*minecraft.Conn).IdentityData().Identity))] = p
			eye.playerMutex.Unlock()
			if eye.ExecuteEvent(events.EventJoin, p) {
				return
			}
			err = lobby.Connect(p)
			if err != nil {
				console.Error("Connecting to lobby", err)
			}
		}()
	}
}

func (eye *eyeProxy) EachPlayer(cb func(player session.Player) bool) {
	for _, p := range eye.players {
		if !cb(p) {
			break
		}
	}
}

func (eye *eyeProxy) GetPlayer(uid uuid.UUID) (session.Player, bool) {
	eye.playerMutex.Lock()
	player, ok := eye.players[uid]
	eye.playerMutex.Unlock()
	return player, ok
}

func (eye *eyeProxy) GetPlayerByXUID(xuid string) (session.Player, bool) {
	if !storage.Config.Listener.Auth {
		return nil, false
	}
	eye.playerMutex.Lock()
	player, ok := eye.playersXuid[xuid]
	eye.playerMutex.Unlock()
	return player, ok
}

func (eye *eyeProxy) Count() int {
	eye.playerMutex.Lock()
	c := len(eye.players)
	eye.playerMutex.Unlock()
	return c
}

func (eye *eyeProxy) HandleEvent(i interface{}) {
	switch handler := i.(type) {
	case events.AnticheatDetectionEventHandler:
		evts, ok := eye.events.Load(events.EventAntiCheatDetection)
		if !ok {
			eye.events.Store(events.EventAntiCheatDetection, []events.AnticheatDetectionEventHandler{handler})
			return
		}
		evtHandlers := evts.([]events.AnticheatDetectionEventHandler)
		evtHandlers = append(evtHandlers, handler)
		eye.events.Store(events.EventAntiCheatDetection, evtHandlers)

	case events.ConnectEventHandler:
		evts, ok := eye.events.Load(events.EventConnect)
		if !ok {
			eye.events.Store(events.EventConnect, []events.ConnectEventHandler{handler})
			return
		}
		evtHandlers := evts.([]events.ConnectEventHandler)
		evtHandlers = append(evtHandlers, handler)
		eye.events.Store(events.EventConnect, evtHandlers)

	case events.JoinEventHandler:
		evts, ok := eye.events.Load(events.EventJoin)
		if !ok {
			eye.events.Store(events.EventJoin, []events.JoinEventHandler{handler})
			return
		}
		evtHandlers := evts.([]events.JoinEventHandler)
		evtHandlers = append(evtHandlers, handler)
		eye.events.Store(events.EventJoin, evtHandlers)

	case events.LeaveEventHandler:
		evts, ok := eye.events.Load(events.EventLeave)
		if !ok {
			eye.events.Store(events.EventLeave, []events.LeaveEventHandler{handler})
			return
		}
		evtHandlers := evts.([]events.LeaveEventHandler)
		evtHandlers = append(evtHandlers, handler)
		eye.events.Store(events.EventLeave, evtHandlers)
	case events.PacketEventHandler:
		evts, ok := eye.events.Load(events.EventPacket)
		if !ok {
			eye.events.Store(events.EventPacket, []events.PacketEventHandler{handler})
			return
		}
		evtHandlers := evts.([]events.PacketEventHandler)
		evtHandlers = append(evtHandlers, handler)
		eye.events.Store(events.EventPacket, evtHandlers)
	default:
	}
}

func (eye *eyeProxy) ExecuteEvent(evt events.Event, args ...interface{}) bool {
	switch evt {
	case events.EventAntiCheatDetection:
		r, ok := eye.events.Load(events.EventAntiCheatDetection)
		if !ok {
			return false
		}
		arr, ok := r.([]events.AnticheatDetectionEventHandler)
		if !ok {
			return false
		}
		ctx := events.NewContext()
		for _, handler := range arr {
			handler(ctx, args[0].(session.Player), args[1].(*anticheat.Detection))
		}
		return ctx.IsCancelled()
	case events.EventConnect:
		r, ok := eye.events.Load(events.EventConnect)
		if !ok {
			return false
		}
		arr, ok := r.([]events.ConnectEventHandler)
		if !ok {
			return false
		}
		ctx := events.NewContext()
		for _, handler := range arr {
			handler(ctx, args[0].(session.Player))
		}
		return ctx.IsCancelled()
	case events.EventJoin:
		r, ok := eye.events.Load(events.EventJoin)
		if !ok {
			return false
		}
		arr, ok := r.([]events.JoinEventHandler)
		if !ok {
			return false
		}
		ctx := events.NewContext()
		for _, handler := range arr {
			handler(ctx, args[0].(session.Player))
		}
		return ctx.IsCancelled()
	case events.EventLeave:
		r, ok := eye.events.Load(events.EventLeave)
		if !ok {
			return false
		}
		arr, ok := r.([]events.LeaveEventHandler)
		if !ok {
			return false
		}
		ctx := events.NewContext()
		for _, handler := range arr {
			handler(ctx, args[0].(session.Player))
		}
		return ctx.IsCancelled()
	case events.EventPacket:
		r, ok := eye.events.Load(events.EventPacket)
		if !ok {
			return false
		}
		arr, ok := r.([]events.PacketEventHandler)
		if !ok {
			return false
		}
		ctx := events.NewContext()
		for _, handler := range arr {
			handler(ctx, args[0].(session.Player), args[1].(session.Source), args[2].(packet.Packet))
		}
		return ctx.IsCancelled()
	}
	return false
}

func (*eyeProxy) ChatPrefix() string {
	return text.Colourf("<bold><grey>[</grey><dark-purple>Proxy</dark-purple><purple>Eye</purple><grey>]</grey></bold>")
}
