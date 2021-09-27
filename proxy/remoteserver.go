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
	"errors"
	"github.com/Suremeo/ProxyEye/proxy/session"
	"github.com/Suremeo/ProxyEye/proxy/session/events"
	"github.com/Suremeo/ProxyEye/proxy/storage"
	"github.com/Suremeo/ProxyEye/proxy/world"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"go.uber.org/atomic"
	"io/ioutil"
	"log"
	"sync"
)

var discardLogger = log.New(ioutil.Discard, "", 0)

type remoteserver struct {
	mutex    sync.Mutex
	address  string
	count    atomic.Int64
	max      int
	players  map[uint64]session.Player
	joinable bool

	listener session.Listener

	worlds *world.Store
}

var emptyChunk = make([]byte, 257)

func NewRemoteServer(address string, max int) session.Server {
	return &remoteserver{address: address, max: max, joinable: true, players: make(map[uint64]session.Player), worlds: world.NewStore(), listener: NopListener}
}

func (r *remoteserver) Joinable() bool {
	if r.max == 0 || r.max > int(r.count.Load()) && r.joinable {
		return true
	}
	return false
}

func (r *remoteserver) SetJoinable(b bool) {
	r.joinable = b
}

func (r *remoteserver) Count() int {
	return int(r.count.Load())
}

func (r *remoteserver) Address() string {
	return r.address
}

func (r *remoteserver) Raknet() bool {
	return true
}

func (r *remoteserver) Connect(player session.Player) error {
	r.count.Inc()
	if !player.Online() {
		r.count.Dec()
		return errors.New("player is not online")
	}
	client := player.Raknet().ClientData()
	client.PlatformOnlineID = player.Raknet().IdentityData().XUID
	conn, err := minecraft.Dialer{
		ErrorLog:     discardLogger,
		ClientData:   client,
		IdentityData: player.Raknet().IdentityData(),
	}.Dial("raknet", r.address)
	if err != nil {
		r.count.Dec()
		return err
	}
	if player.Session() == NopSession {
		gd := conn.GameData()
		gd.EntityRuntimeID = 1
		gd.EntityUniqueID = 1
		gd.WorldName = text.Colourf(storage.Config.Listener.Motd)
		player.Anticheat().Reset(conn.GameData().PlayerPosition, conn.GameData().Pitch, conn.GameData().Yaw)
		s := NewSession(r, player, conn)
		player.Chunks().SetWorld(r.worlds.GetOrCreate(conn.GameData().WorldName))
		s.SetConnected()
		player.SetSession(s)
		var errs = make(chan error)
		var g sync.WaitGroup
		g.Add(2)
		go func() {
			if err := player.Raknet().StartGame(gd); err != nil {
				errs <- err
			}
			g.Done()
		}()
		go func() {
			if err := conn.DoSpawn(); err != nil {
				errs <- err
			}
			g.Done()
		}()
		g.Wait()
		if len(errs) != 0 {
			s.Close()
			return <-errs
		}
	} else {
		if err := conn.DoSpawn(); err != nil {
			_ = conn.Close()
			return err
		}
		player.SetTransferring(true)
		player.Session().Close()
		player.SetSession(NewSession(r, player, conn))
		player.SetTransferring(false)
		player.Anticheat().Reset(conn.GameData().PlayerPosition, conn.GameData().Pitch, conn.GameData().Yaw)
		_ = player.WritePacket(&packet.SetTime{Time: int32(conn.GameData().Time)})
		_ = player.WritePacket(&packet.SetPlayerGameType{GameType: conn.GameData().PlayerGameMode})
		_ = player.WritePacket(&packet.MovePlayer{
			EntityRuntimeID: 1,
			Position:        conn.GameData().PlayerPosition,
			Pitch:           conn.GameData().Pitch,
			Yaw:             conn.GameData().Yaw,
			HeadYaw:         conn.GameData().Yaw,
			OnGround:        true,
		})
		_ = player.WritePacket(&packet.Respawn{
			Position:        conn.GameData().PlayerPosition,
			State:           packet.RespawnStateReadyToSpawn,
			EntityRuntimeID: 1,
		})
		pos := conn.GameData().PlayerPosition
		chunkX := int32(pos.X()) >> 4
		chunkZ := int32(pos.Z()) >> 4
		for x := int32(-2); x <= 2; x++ {
			for z := int32(-2); z <= 2; z++ {
				_ = player.WritePacket(&packet.LevelChunk{
					ChunkX:        chunkX + x,
					ChunkZ:        chunkZ + z,
					SubChunkCount: 0,
					RawPayload:    emptyChunk,
				})
			}
		}
		player.Chunks().SetWorld(r.worlds.GetOrCreate(conn.GameData().WorldName))
		player.Session().SetConnected()
	}
	r.mutex.Lock()
	r.players[player.Session().OldRuntimeId()] = player
	r.mutex.Unlock()
	go Eye.ExecuteEvent(events.EventConnect, player)
	if !player.Online() {
		player.Session().Close()
	}
	return nil
}

func (r *remoteserver) Packet(source session.Source, player session.Player, pk packet.Packet) bool {
	if Eye.ExecuteEvent(events.EventPacket, player, source, pk) {
		return false
	}
	r.listener.Packet(source, player, pk)
	if source.Type() == session.SourceTypePlayer {
		switch pk := pk.(type) {
		case *packet.MovePlayer:
			player.Anticheat().Move(pk.Position, pk.Pitch, pk.HeadYaw)
		case *packet.InventoryTransaction:
			data, ok := pk.TransactionData.(*protocol.UseItemOnEntityTransactionData)
			if ok {
				if data.ActionType == protocol.UseItemOnEntityActionAttack {
					target, ok := r.PlayerByRuntimeId(source.TranslateEid(data.TargetEntityRuntimeID))
					if ok {
						player.Anticheat().Attack(target)
					}
				}
			}
		case *packet.PlayerAuthInput:
			player.Anticheat().Move(pk.Position, pk.Pitch, pk.HeadYaw)
		case *packet.UpdateBlock:
			player.Chunks().UpdateBlock(pk.Position, pk.NewBlockRuntimeID)
		case *packet.AdventureSettings:
			player.Anticheat().UpdateFlags(pk.Flags)
		}
	} else {
		switch pk := pk.(type) {
		case *packet.MovePlayer:
			if pk.EntityRuntimeID == player.NewRuntimeId() {
				player.Anticheat().Teleport(pk.Position, pk.Pitch, pk.HeadYaw)
			}
		case *packet.CorrectPlayerMovePrediction:
			player.Anticheat().Teleport(pk.Position, player.Anticheat().Pitch(), player.Anticheat().Yaw())
		case *packet.UpdateBlock:
			player.Chunks().UpdateBlock(pk.Position, pk.NewBlockRuntimeID)
		case *packet.UpdateBlockSynced:
			player.Chunks().UpdateBlock(pk.Position, pk.NewBlockRuntimeID)
		case *packet.ChunkRadiusUpdated:
			player.Chunks().UpdateRadius(int(pk.ChunkRadius))
		case *packet.LevelChunk:
			if player.Anticheat().UseChunks() {
				player.Chunks().AddChunk(pk)
			}
		case *packet.SetActorMotion:
		case *packet.AdventureSettings:
			player.Anticheat().SetFlags(pk.Flags)
		}
	}
	return true
}

func (r *remoteserver) Leave(player session.Player, session session.Session) {
	r.count.Dec()
	r.mutex.Lock()
	delete(r.players, session.OldRuntimeId())
	r.mutex.Unlock()
	if !player.IsTransferring() && player.Session().Server().Address() == r.Address() {
		_ = player.WritePacket(&packet.Disconnect{
			Message: text.Colourf("<red>The server you were on has went down...</red>"),
		})
		player.Close()
	}
	r.listener.Leave(player)
}

func (r *remoteserver) PlayerByRuntimeId(u uint64) (session.Player, bool) {
	p, ok := r.players[u]
	return p, ok
}

func (r *remoteserver) SetListener(listener session.Listener) {
	r.listener = listener
}

func (r *remoteserver) Listener() session.Listener {
	return r.listener
}

func (r *remoteserver) Players() map[uint64]session.Player {
	return r.players
}
