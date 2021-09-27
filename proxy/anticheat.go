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
	"github.com/Suremeo/ProxyEye/proxy/session/anticheat"
	"github.com/Suremeo/ProxyEye/proxy/session/events"
	"github.com/Suremeo/ProxyEye/proxy/world/chunk"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"go.uber.org/atomic"
	"math"
	"sync"
	"time"
)

type anticheatProfile struct {
	mutex sync.Mutex

	player session.Player

	gamemode atomic.Int32
	flags    atomic.Uint32

	movement *antiCheatMovement

	combat *antiCheatCombat

	cooldowns sync.Map

	useChunks bool

	ticker *time.Ticker

	close sync.Once
}

type antiCheatMovement struct {
	mutex      sync.Mutex
	pos        mgl32.Vec3
	pitch, yaw float32

	speedPoints          atomic.Float64
	speedPerSecond       atomic.Float64
	speedPerSecondPoints atomic.Float64
	speedRequests        atomic.Int64

	flyPoints atomic.Float64
}

type antiCheatCombat struct {
	attackPoints          atomic.Float64
	attackCount           atomic.Int64
	attackPerSecondPoints atomic.Float64
}

func newAntiCheatProfile(p session.Player) session.AntiCheatProfile {
	a := &anticheatProfile{
		player: p,

		movement: &antiCheatMovement{},

		combat: &antiCheatCombat{},

		useChunks: true,

		ticker: time.NewTicker(1 * time.Second),
	}
	a.tick()
	return a
}

func (p *anticheatProfile) Position() mgl32.Vec3 {
	p.movement.mutex.Lock()
	defer p.movement.mutex.Unlock()
	return p.movement.pos
}

func (p *anticheatProfile) Move(vec3 mgl32.Vec3, pitch, yaw float32) {
	p.movement.mutex.Lock()
	defer p.movement.mutex.Unlock()
	if p.movement.pos == vec3 {
		return
	}
	defer func() {
		p.movement.pitch = pitch
		p.movement.yaw = yaw
		p.movement.pos = vec3
	}()
	if p.canFly() {
		return
	}
	if p.useChunks {
		p.player.Chunks().Move(vec3)
		if !p.onGround(vec3) {
			// They are flying, we have to make sure they are actually going down...
			if p.movement.pos.Y() < vec3.Y() {
				// They are going up
				i := p.movement.flyPoints.Add(1)
				if i > 20 {
					p.detected(&anticheat.Detection{
						Type:      anticheat.DetectionTypeFly,
						Arguments: make(map[string]interface{}),
					})
					p.movement.flyPoints.Store(0)
				}
			} else {
				// They are going down
				i := p.movement.flyPoints.Sub(.25)
				if i < 0 {
					p.movement.flyPoints.Store(0)
				}
			}
		} else {
			i := p.movement.flyPoints.Sub(.5)
			if i < 0 {
				p.movement.flyPoints.Store(0)
			}
		}
	}
	p.movement.speedRequests.Inc()
	x1 := float64(p.movement.pos.X())
	z1 := float64(p.movement.pos.Z())

	x2 := float64(vec3.X())
	z2 := float64(vec3.Z())
	distance := math.Sqrt(((x2 - x1) * (x2 - x1)) + ((z2 - z1) * (z2 - z1)))
	if distance > .8 {
		p.movement.speedPoints.Add(2)
	} else {
		i := p.movement.speedPoints.Sub(.5)
		if i < 0 {
			p.movement.speedPoints.Store(0)
		}
	}
	if p.movement.speedPoints.Load() >= 8 {
		p.detected(&anticheat.Detection{
			Type:      anticheat.DetectionTypeSpeed,
			Arguments: make(map[string]interface{}),
		})
		p.movement.speedPoints.Store(0)
	}
	if p.movement.speedPerSecond.Load() > 40 {
		i := p.movement.speedPerSecondPoints.Add(2)
		if i > 24 {
			p.detected(&anticheat.Detection{
				Type: anticheat.DetectionTypeTimer,
				Arguments: map[string]interface{}{
					"TPS": int(p.movement.speedPerSecond.Load()),
				},
			})
			p.movement.speedPerSecondPoints.Store(0)
		}
	} else {
		i := p.movement.speedPerSecondPoints.Sub(2)
		if i < 0 {
			p.movement.speedPerSecondPoints.Store(0)
		}
	}
}

func (p *anticheatProfile) Player() session.Player {
	return p.player
}

func (p *anticheatProfile) Close() {
	p.close.Do(func() {
		p.ticker.Stop()
		p.clear()
	})
}

func (p *anticheatProfile) Teleport(vec3 mgl32.Vec3, pitch, yaw float32) {
	p.movement.mutex.Lock()
	defer p.movement.mutex.Unlock()
	p.movement.pitch = pitch
	p.movement.yaw = yaw
	p.movement.pos = vec3
}

func (p *anticheatProfile) tick() {
	go func() {
		for range p.ticker.C {
			p.movement.mutex.Lock()
			p.movement.speedPerSecond.Store(float64(p.movement.speedRequests.Swap(0)))
			p.movement.mutex.Unlock()
			aps := p.combat.attackCount.Swap(0)
			if aps >= 20 {
				if p.combat.attackPerSecondPoints.Add(4) >= 12 {
					p.detected(&anticheat.Detection{
						Type: anticheat.DetectionTypeAutoclicker,
						Arguments: map[string]interface{}{
							"CPS": int(aps),
						},
					})
					p.combat.attackPerSecondPoints.Store(0)
				}
			} else {
				if aps != 0 {
					i := p.combat.attackPerSecondPoints.Sub(.1)
					if i < 0 {
						p.combat.attackPerSecondPoints.Store(0)
					}
				}
			}
		}
	}()
}

func (p *anticheatProfile) Attack(player session.Player) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.combat.attackCount.Inc()
	if p.validAttackDirection(player) {
		i := p.combat.attackPoints.Sub(.5)
		if i < 0 {
			p.combat.attackPoints.Store(0)
		}
	} else {
		if p.combat.attackPoints.Add(2) >= 10 {
			p.combat.attackPoints.Store(0)
			p.detected(&anticheat.Detection{
				Type:      anticheat.DetectionTypeKillaura,
				Arguments: make(map[string]interface{}),
			})
		}
	}
}

func (p *anticheatProfile) Yaw() float32 {
	p.movement.mutex.Lock()
	defer p.movement.mutex.Unlock()
	return p.movement.yaw
}

func (p *anticheatProfile) Pitch() float32 {
	p.movement.mutex.Lock()
	defer p.movement.mutex.Unlock()
	return p.movement.pitch
}

func (p *anticheatProfile) Facing() session.Direction {
	yaw := math.Mod(float64(p.Yaw())-90, 360)
	if yaw < 0 {
		yaw += 360
	}
	switch {
	case (yaw > 0 && yaw < 45) || (yaw > 315 && yaw < 360):
		return session.DirectionWest
	case yaw > 45 && yaw < 135:
		return session.DirectionNorth
	case yaw > 135 && yaw < 225:
		return session.DirectionEast
	case yaw > 225 && yaw < 315:
		return session.DirectionSouth
	}
	return 0
}

func (p *anticheatProfile) validAttackDirection(player session.Player) bool {
	facing := p.Facing()
	pd := p.playerDirection(player)
	switch facing {
	case session.DirectionNorth:
		if pd == session.DirectionNorthWest || pd == session.DirectionNorthEast {
			return true
		}
		break
	case session.DirectionSouth:
		if pd == session.DirectionSouthWest || pd == session.DirectionSouthEast {
			return true
		}
		break
	case session.DirectionEast:
		if pd == session.DirectionNorthEast || pd == session.DirectionSouthEast {
			return true
		}
		break
	case session.DirectionWest:
		if pd == session.DirectionNorthWest || pd == session.DirectionSouthWest {
			return true
		}
		break
	}
	return false
}

func (p *anticheatProfile) playerDirection(player session.Player) session.Direction {
	pos := player.Anticheat().Position()
	x := pos.X()
	y := pos.Z()
	p.movement.mutex.Lock()
	x2 := p.movement.pos.X()
	y2 := p.movement.pos.Z()
	p.movement.mutex.Unlock()
	if x2 < x && y2 > y {
		return session.DirectionNorthEast
	} else if x2 > x && y2 > y {
		return session.DirectionNorthWest
	} else if x2 < x && y2 < y {
		return session.DirectionSouthEast
	} else if x2 > x && y2 < y {
		return session.DirectionSouthWest
	}
	return 0
}

func (p *anticheatProfile) detected(detection *anticheat.Detection) {
	c, ok := p.cooldowns.Load(detection.Type)
	if ok {
		if c.(time.Time).After(time.Now()) {
			return
		}
	}
	p.cooldowns.Store(detection.Type, time.Now().Add(30*time.Second))
	Eye.ExecuteEvent(events.EventAntiCheatDetection, p.player, detection)
}

func (p *anticheatProfile) SetUseChunks(b bool) {
	p.mutex.Lock()
	p.useChunks = b
	p.mutex.Unlock()
}

func (p *anticheatProfile) UseChunks() bool {
	return p.useChunks
}

func (p *anticheatProfile) Flags() uint32 {
	return p.flags.Load()
}

func (p *anticheatProfile) Gamemode() int32 {
	return p.gamemode.Load()
}

func (p *anticheatProfile) SetGamemode(i int32) {
	p.gamemode.Store(i)
}

func (p *anticheatProfile) SetFlags(i uint32) {
	p.flags.Store(i)
}

func (p *anticheatProfile) UpdateFlags(i uint32) {
	switch {
	case i&packet.AdventureFlagFlying != 0:
		if !p.canFly() {
			p.detected(&anticheat.Detection{
				Type:      anticheat.DetectionTypeFly,
				Arguments: make(map[string]interface{}),
			})
			return // Client cannot fly because flight is disabled.
		}
	case i&packet.AdventureFlagAllowFlight != 0:
		if p.flags.Load()&packet.AdventureFlagAllowFlight == 0 {
			return // Client is trying to allow flight for itself.
		}
	case i&packet.AdventureFlagNoClip != 0:
		if p.flags.Load()&packet.AdventureFlagNoClip == 0 {
			return // Client is trying to noclip.
		}
	case i&packet.AdventureFlagWorldBuilder != 0:
		if p.flags.Load()&packet.AdventureFlagWorldBuilder == 0 {
			return // Client is trying to enable WorldBuilder flag
		}
	case i&packet.AdventureFlagNoPVP != 0:
		if p.flags.Load()&packet.AdventureFlagNoPVP == 0 {
			return // Client is trying to enable NoPVP flag
		}
	case i&packet.AdventureFlagWorldImmutable != 0:
		if p.flags.Load()&packet.AdventureFlagWorldImmutable == 0 {
			return // Client is trying to enable WorldImmutable flag
		}
	}
	p.flags.Store(i)
}

func (p *anticheatProfile) canFly() bool {
	if packet.AdventureFlagAllowFlight&p.flags.Load() != 0 {
		return true
	}
	switch p.gamemode.Load() {
	case packet.GameTypeCreative: // Spectator adventure flags should be sent by the server
		return true
	}
	return false
}

func (p *anticheatProfile) isFlying() bool {
	return packet.AdventureFlagFlying&p.flags.Load() != 0
}

func (p *anticheatProfile) canNoClip() bool {
	return packet.AdventureFlagNoClip&p.flags.Load() != 0
}

func (p *anticheatProfile) onGround(pos mgl32.Vec3) bool {
	if !p.useChunks {
		return true
	}
	id := p.player.Chunks().Floor(pos)
	if id != chunk.AirRuntimeId {
		return true
	} else {
		pos[1]--
		id := p.player.Chunks().Floor(pos)
		if id != chunk.AirRuntimeId {
			return true
		}
	}
	return false
}

func (p *anticheatProfile) Reset(pos mgl32.Vec3, pitch, yaw float32) {
	p.mutex.Lock()
	p.clear()
	p.movement.pos = pos
	p.movement.pitch = pitch
	p.movement.yaw = yaw
	p.mutex.Unlock()
}

func (p *anticheatProfile) clear() {
	p.useChunks = true
	p.gamemode.Store(0)
	p.flags.Store(0)

	p.movement.pitch = 0
	p.movement.yaw = 0
	p.movement.pos = mgl32.Vec3{}

	p.movement.speedPerSecondPoints.Store(0)
	p.movement.speedPoints.Store(0)
	p.movement.speedPerSecondPoints.Store(0)
	p.movement.speedPerSecondPoints.Store(0)
	p.movement.speedRequests.Store(0)
	p.movement.flyPoints.Store(0)

	p.combat.attackPoints.Store(0)
	p.combat.attackCount.Store(0)
	p.combat.attackPerSecondPoints.Store(0)
}
