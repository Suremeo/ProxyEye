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
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"go.uber.org/atomic"
	"sync"
)

type eyeSession struct {
	mutex sync.Mutex

	runtimeId atomic.Uint64
	uniqueId  atomic.Int64

	server session.Server

	conn *minecraft.Conn

	connected chan struct{}

	close sync.Once

	player session.Player

	scoreboards sync.Map
	playerList  sync.Map
	entities    sync.Map
}

func NewSession(server session.Server, player session.Player, conn *minecraft.Conn) session.Session {
	s := &eyeSession{
		server:    server,
		conn:      conn,
		player:    player,
		connected: make(chan struct{}),
	}
	if server.Raknet() {
		s.runtimeId.Store(conn.GameData().EntityRuntimeID)
		s.uniqueId.Store(conn.GameData().EntityUniqueID)
		go func() {
			<-s.connected
			for {
				pk, err := s.conn.ReadPacket()
				if err != nil {
					s.Close()
					return
				}
				if s.Server().Packet(s, s.player, pk) {
					_ = s.Packet(s, player, pk)
				}
			}
		}()
	} else {
		s.runtimeId.Store(1)
		s.uniqueId.Store(1)
	}
	return s
}

func (s *eyeSession) WritePacket(pk packet.Packet) error {
	if s.server.Raknet() {
		return s.conn.WritePacket(pk)
	}
	s.server.Packet(s, s.player, pk)
	return nil
}

func (s *eyeSession) Packet(source session.Source, destination session.Destination, pk packet.Packet) error {
	switch pk := pk.(type) {
	case *packet.SetDisplayObjective:
		if source.Type() != session.SourceTypePlayer {
			s.scoreboards.Store(pk.ObjectiveName, struct{}{})
		}
	case *packet.RemoveObjective:
		if source.Type() != session.SourceTypePlayer {
			s.scoreboards.Delete(pk.ObjectiveName)
		}
	case *packet.ActorEvent:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.ActorPickRequest:
		pk.EntityUniqueID = source.TranslateUid(pk.EntityUniqueID)
	case *packet.AddActor:
		pk.EntityUniqueID = source.TranslateUid(pk.EntityUniqueID)
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
		if source.Type() != session.SourceTypePlayer {
			s.entities.Store(pk.EntityUniqueID, struct{}{})
		}
	case *packet.AddItemActor:
		pk.EntityUniqueID = source.TranslateUid(pk.EntityUniqueID)
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
		if source.Type() != session.SourceTypePlayer {
			s.entities.Store(pk.EntityUniqueID, struct{}{})
		}
	case *packet.AddPainting:
		pk.EntityUniqueID = source.TranslateUid(pk.EntityUniqueID)
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
		if source.Type() != session.SourceTypePlayer {
			s.entities.Store(pk.EntityUniqueID, struct{}{})
		}
	case *packet.AddPlayer:
		pk.EntityUniqueID = source.TranslateUid(pk.EntityUniqueID)
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
		if source.Type() != session.SourceTypePlayer {
			s.entities.Store(pk.EntityUniqueID, struct{}{})
		}
	case *packet.AdventureSettings:
		pk.PlayerUniqueID = source.TranslateUid(pk.PlayerUniqueID)
	case *packet.Animate:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.AnimateEntity:
		for i := range pk.EntityRuntimeIDs {
			pk.EntityRuntimeIDs[i] = source.TranslateEid(pk.EntityRuntimeIDs[i])
		}
	case *packet.BossEvent:
		pk.BossEntityUniqueID = source.TranslateUid(pk.BossEntityUniqueID)
		pk.PlayerUniqueID = source.TranslateUid(pk.PlayerUniqueID)
	case *packet.Camera:
		pk.CameraEntityUniqueID = source.TranslateUid(pk.CameraEntityUniqueID)
		pk.TargetPlayerUniqueID = source.TranslateUid(pk.TargetPlayerUniqueID)
	case *packet.CommandOutput:
		pk.CommandOrigin.PlayerUniqueID = source.TranslateUid(pk.CommandOrigin.PlayerUniqueID)
	case *packet.CommandRequest:
		pk.CommandOrigin.PlayerUniqueID = source.TranslateUid(pk.CommandOrigin.PlayerUniqueID)
	case *packet.ContainerOpen:
		pk.ContainerEntityUniqueID = source.TranslateUid(pk.ContainerEntityUniqueID)
	case *packet.DebugInfo:
		pk.PlayerUniqueID = source.TranslateUid(pk.PlayerUniqueID)
	case *packet.Emote:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.EmoteList:
		pk.PlayerRuntimeID = source.TranslateEid(pk.PlayerRuntimeID)
	case *packet.Event:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.Interact:
		pk.TargetEntityRuntimeID = source.TranslateEid(pk.TargetEntityRuntimeID)
	case *packet.MobArmourEquipment:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.MobEffect:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.MobEquipment:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.MotionPredictionHints:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.MoveActorAbsolute:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.MoveActorDelta:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.MovePlayer:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
		pk.RiddenEntityRuntimeID = source.TranslateEid(pk.RiddenEntityRuntimeID)
	case *packet.NPCRequest:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.PlayerAction:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.PlayerList:
		for i := range pk.Entries {
			pk.Entries[i].EntityUniqueID = source.TranslateUid(pk.Entries[i].EntityUniqueID)
		}
		if source.Type() != session.SourceTypePlayer {
			if pk.ActionType == packet.PlayerListActionRemove {
				for i := range pk.Entries {
					s.playerList.Delete(pk.Entries[i].UUID)
				}
			} else if pk.ActionType == packet.PlayerListActionAdd {
				for i := range pk.Entries {
					s.playerList.Store(pk.Entries[i].UUID, struct{}{})
				}
			}
		}
	case *packet.RemoveActor:
		pk.EntityUniqueID = source.TranslateUid(pk.EntityUniqueID)
		if source.Type() != session.SourceTypePlayer {
			s.entities.Delete(pk.EntityUniqueID)
		}
	case *packet.Respawn:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.SetActorData:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.SetActorLink:
		pk.EntityLink.RiddenEntityUniqueID = source.TranslateUid(pk.EntityLink.RiddenEntityUniqueID)
		pk.EntityLink.RiderEntityUniqueID = source.TranslateUid(pk.EntityLink.RiderEntityUniqueID)
	case *packet.SetActorMotion:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.SetLocalPlayerAsInitialised:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.SetScore:
		for i := range pk.Entries {
			pk.Entries[i].EntityUniqueID = source.TranslateUid(pk.Entries[i].EntityUniqueID)
		}
	case *packet.SetScoreboardIdentity:
		for i := range pk.Entries {
			pk.Entries[i].EntityUniqueID = source.TranslateUid(pk.Entries[i].EntityUniqueID)
		}
	case *packet.ShowCredits:
		pk.PlayerRuntimeID = source.TranslateEid(pk.PlayerRuntimeID)
	case *packet.SpawnParticleEffect:
		pk.EntityUniqueID = source.TranslateUid(pk.EntityUniqueID)
	case *packet.StartGame:
		pk.EntityUniqueID = source.TranslateUid(pk.EntityUniqueID)
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.StructureBlockUpdate:
		pk.Settings.LastEditingPlayerUniqueID = source.TranslateUid(pk.Settings.LastEditingPlayerUniqueID)
	case *packet.StructureTemplateDataRequest:
		pk.Settings.LastEditingPlayerUniqueID = source.TranslateUid(pk.Settings.LastEditingPlayerUniqueID)
	case *packet.TakeItemActor:
		pk.ItemEntityRuntimeID = source.TranslateEid(pk.ItemEntityRuntimeID)
		pk.TakerEntityRuntimeID = source.TranslateEid(pk.TakerEntityRuntimeID)
	case *packet.UpdateAttributes:
		pk.EntityRuntimeID = source.TranslateEid(pk.EntityRuntimeID)
	case *packet.UpdateEquip:
		pk.EntityUniqueID = source.TranslateUid(pk.EntityUniqueID)
	case *packet.UpdatePlayerGameType:
		pk.PlayerUniqueID = source.TranslateUid(pk.PlayerUniqueID)
	case *packet.UpdateTrade:
		pk.VillagerUniqueID = source.TranslateUid(pk.VillagerUniqueID)
		pk.EntityUniqueID = source.TranslateUid(pk.EntityUniqueID)
	}
	return destination.WritePacket(pk)
}

func (s *eyeSession) OldRuntimeId() uint64 {
	return s.runtimeId.Load()
}

func (s *eyeSession) NewRuntimeId() uint64 {
	return s.player.OldRuntimeId()
}

func (s *eyeSession) OldUniqueId() int64 {
	return s.uniqueId.Load()
}

func (s *eyeSession) NewUniqueId() int64 {
	return s.player.OldUniqueId()
}

func (s *eyeSession) TranslateEid(now uint64) uint64 {
	if s.OldRuntimeId() == now {
		return s.NewRuntimeId()
	} else if s.NewRuntimeId() == now {
		return s.OldRuntimeId()
	}
	return now
}

func (s *eyeSession) TranslateUid(now int64) int64 {
	if s.OldUniqueId() == now {
		return s.NewUniqueId()
	} else if s.NewUniqueId() == now {
		return s.OldUniqueId()
	}
	return now
}

func (s *eyeSession) Server() session.Server {
	s.mutex.Lock()
	serv := s.server
	s.mutex.Unlock()
	return serv
}

func (s *eyeSession) Close() {
	s.close.Do(func() {
		if s.server.Raknet() {
			if s.conn != nil {
				_ = s.conn.Close()
			}
		}
		if s.player.Online() {
			var wg sync.WaitGroup
			wg.Add(3)
			go func() {
				s.entities.Range(func(key, value interface{}) bool {
					_ = s.player.WritePacket(&packet.RemoveActor{EntityUniqueID: key.(int64)})
					s.entities.Delete(key)
					return true
				})
				wg.Done()
			}()
			go func() {
				s.scoreboards.Range(func(key, value interface{}) bool {
					_ = s.player.WritePacket(&packet.RemoveObjective{
						ObjectiveName: key.(string),
					})
					s.scoreboards.Delete(key)
					return true
				})
				wg.Done()
			}()
			go func() {
				var temp []protocol.PlayerListEntry
				s.playerList.Range(func(key, value interface{}) bool {
					temp = append(temp, protocol.PlayerListEntry{UUID: key.(uuid.UUID)})
					return true
				})
				wg.Done()
			}()
			wg.Wait()
		}
		select {
		case <-s.connected:
		default:
			close(s.connected)
		}
		s.server.Leave(s.player, s)
	})
}

func (s *eyeSession) SetConnected() {
	select {
	case <-s.connected:
	default:
		close(s.connected)
	}
}

func (s *eyeSession) IsConnected() bool {
	select {
	case <-s.connected:
		return true
	default:
		return false
	}
}

func (s *eyeSession) Type() session.SourceType {
	return session.SourceTypeSession
}
