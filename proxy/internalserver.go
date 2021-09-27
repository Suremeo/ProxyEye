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
	"crypto/rand"
	"encoding/hex"
	"github.com/Suremeo/ProxyEye/proxy/session"
	"github.com/Suremeo/ProxyEye/proxy/storage"
	"github.com/Suremeo/ProxyEye/proxy/world"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"sync"
)

type internalServer struct {
	mutex sync.Mutex

	id string

	listener session.Listener

	worlds *world.Store

	gamedata minecraft.GameData
}

func NewInternalServer() *internalServer {
	return &internalServer{
		id: func() string {
			t := make([]byte, 128)
			_, _ = rand.Read(t)
			return hex.EncodeToString(t)
		}(),
		worlds:   world.NewStore(),
		listener: NopListener,
		gamedata: minecraft.GameData{
			WorldName:                       text.Colourf(storage.Config.Listener.Motd),
			Difficulty:                      0,
			EntityUniqueID:                  1,
			EntityRuntimeID:                 1,
			PlayerGameMode:                  packet.GameTypeCreative,
			PlayerPosition:                  mgl32.Vec3{0, 0, 0},
			Pitch:                           0,
			Yaw:                             0,
			Dimension:                       packet.DimensionOverworld,
			WorldSpawn:                      protocol.BlockPos{0, 0, 0},
			WorldGameMode:                   0,
			GameRules:                       map[string]interface{}{},
			Items:                           []protocol.ItemEntry{},
			ServerAuthoritativeMovementMode: packet.AuthoritativeMovementModeClient,
			Experiments:                     []protocol.ExperimentData{},
		},
	}
}

func (i *internalServer) Joinable() bool {
	return true
}

func (i *internalServer) SetJoinable(b bool) {}

func (i *internalServer) Count() int {
	return 0
}

func (i *internalServer) Address() string {
	return i.id
}

func (i *internalServer) Raknet() bool {
	return false
}

func (i *internalServer) Connect(player session.Player) error {
	if player.Session() == NopSession {
		player.Anticheat().Reset(mgl32.Vec3{0, 0, 0}, 0, 0)
		s := NewSession(i, player, nil)
		player.Chunks().SetWorld(i.worlds.GetOrCreate("default"))
		s.SetConnected()
		player.SetSession(s)
		var errs = make(chan error)
		var g sync.WaitGroup
		g.Add(1)
		go func() {
			if err := player.Raknet().StartGame(i.gamedata); err != nil {
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
		player.SetTransferring(true)
		player.Session().Close()
		player.SetSession(NewSession(i, player, nil))
		player.SetTransferring(false)
		player.Anticheat().Reset(mgl32.Vec3{0, 0, 0}, 0, 0)
		_ = player.WritePacket(&packet.SetTime{Time: int32(0)})
		_ = player.WritePacket(&packet.SetPlayerGameType{GameType: 1})
		_ = player.WritePacket(&packet.Respawn{
			Position:        mgl32.Vec3{0, 0, 0},
			State:           packet.RespawnStateReadyToSpawn,
			EntityRuntimeID: 1,
		})
		pos := mgl32.Vec3{0, 0, 0}
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
		player.Chunks().SetWorld(i.worlds.GetOrCreate("default"))
		player.Session().SetConnected()
	}
	return nil
}

func (i *internalServer) Leave(player session.Player, s session.Session) {}

func (i *internalServer) PlayerByRuntimeId(u uint64) (session.Player, bool) {
	return nil, false
}

func (i *internalServer) Packet(source session.Source, player session.Player, packet packet.Packet) bool {
	i.listener.Packet(source, player, packet)
	return true
}

func (i *internalServer) SetListener(listener session.Listener) {
	i.mutex.Lock()
	i.listener = listener
	i.mutex.Unlock()
}

func (i *internalServer) Listener() session.Listener {
	return i.listener
}

func (i *internalServer) Players() map[uint64]session.Player {
	return nil
}
