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
	"github.com/Suremeo/ProxyEye/proxy/world"
	"github.com/Suremeo/ProxyEye/proxy/world/blocks"
	"github.com/Suremeo/ProxyEye/proxy/world/chunk"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"math"
	"strings"
	"sync"
)

type chunkManager struct {
	mutex sync.Mutex

	player session.Player

	pos chunk.Pos

	world *world.World

	chunks map[chunk.Pos]*world.Chunk
	radius int

	close sync.Once

	closed bool
}

func newChunks(player session.Player, radius int, w *world.World) session.Chunks {
	return &chunkManager{
		mutex:  sync.Mutex{},
		player: player,
		chunks: make(map[chunk.Pos]*world.Chunk),
		radius: radius,
		world:  w,
	}
}

func (chunks *chunkManager) World() *world.World {
	chunks.mutex.Lock()
	w := chunks.world
	chunks.mutex.Unlock()
	return w
}

func (chunks *chunkManager) SetWorld(w *world.World) {
	chunks.mutex.Lock()
	chunks.world = w
	chunks.mutex.Unlock()
}

func (chunks *chunkManager) UpdateBlock(pos protocol.BlockPos, block uint32) {
	if chunks.closed {
		return
	}
	if !chunks.player.Anticheat().UseChunks() {
		return
	}
	y := pos[1]
	if y > 255 || y < 0 {
		return
	}
	x, z := pos[0]>>4, pos[2]>>4
	c, ok := chunks.chunks[chunk.Pos{x, z}]
	if !ok {
		return
	}
	c.SetRuntimeID(uint8(pos[0]), uint8(pos[1]), uint8(pos[2]), 0, block)
}

func (chunks *chunkManager) GetBlock(pos protocol.BlockPos) uint32 {
	if chunks.closed {
		return chunk.AirRuntimeId
	}
	if !chunks.player.Anticheat().UseChunks() {
		return chunk.AirRuntimeId
	}
	y := pos[1]
	if y > 255 || y < 0 {
		return chunk.AirRuntimeId
	}
	x, z := pos[0]>>4, pos[2]>>4
	chunks.mutex.Lock()
	c, ok := chunks.chunks[chunk.Pos{x, z}]
	chunks.mutex.Unlock()
	if !ok {
		return chunk.AirRuntimeId
	}
	return c.RuntimeID(uint8(pos[0]), uint8(pos[1]), uint8(pos[2]), 0)
}

func (chunks *chunkManager) Move(pos mgl32.Vec3) {
	if chunks.closed {
		return
	}
	if !chunks.player.Anticheat().UseChunks() {
		return
	}
	chunks.mutex.Lock()

	floorX, floorZ := math.Floor(float64(pos[0])), math.Floor(float64(pos[2]))
	chunkPos := chunk.Pos{int32(floorX) >> 4, int32(floorZ) >> 4}

	if chunkPos == chunks.pos {
		chunks.mutex.Unlock()
		return
	}
	chunks.pos = chunkPos
	chunks.evictUnused()
	chunks.mutex.Unlock()
}

func (chunks *chunkManager) Chunk() (chunk *world.Chunk, ok bool) {
	if chunks.closed {
		return
	}
	if !chunks.player.Anticheat().UseChunks() {
		return
	}
	chunks.mutex.Lock()
	chunk, ok = chunks.chunks[chunks.pos]
	chunks.mutex.Unlock()
	return
}

func (chunks *chunkManager) AddChunk(pk *packet.LevelChunk) {
	if chunks.closed {
		return
	}
	if !chunks.player.Anticheat().UseChunks() {
		return
	}
	c, err := chunks.World().AddChunk(pk)
	if err != nil {
		if strings.HasPrefix(err.Error(), "unknown sub chunk version") {
			chunks.player.Anticheat().SetUseChunks(false)
			console.Debug("%v is on a server that uses an unsupported chunk version (All anticheat features that require chunks will be disabled for this user)", chunks.player.Raknet().IdentityData().DisplayName)
		}
		return
	}
	chunks.mutex.Lock()
	chunks.chunks[chunk.Pos{pk.ChunkX, pk.ChunkZ}] = c
	c.AddViewer(chunks.player.NewRuntimeId())
	chunks.mutex.Unlock()
}

func (chunks *chunkManager) Radius() int {
	return chunks.radius
}

func (chunks *chunkManager) UpdateRadius(i int) {
	chunks.mutex.Lock()
	chunks.radius = i
	chunks.mutex.Unlock()
}

func (chunks *chunkManager) Close() {
	chunks.close.Do(func() {
		chunks.mutex.Lock()
		chunks.closed = true
		for _, c := range chunks.chunks {
			c.RemoveViewer(chunks.player.NewRuntimeId())
		}
		chunks.chunks = nil
		chunks.player = nil
		chunks.mutex.Unlock()
	})
}

func (chunks *chunkManager) Player() session.Player {
	return chunks.player
}

func (chunks *chunkManager) evictUnused() {
	if chunks.closed {
		return
	}
	for pos, c := range chunks.chunks {
		diffX, diffZ := pos[0]-chunks.pos[0], pos[1]-chunks.pos[1]
		dist := math.Sqrt(float64(diffX*diffX) + float64(diffZ*diffZ))
		if int(dist) > chunks.radius {
			delete(chunks.chunks, pos)
			c.RemoveViewer(chunks.player.NewRuntimeId())
		}
	}
}

func (chunks *chunkManager) Floor(vec3 mgl32.Vec3) uint32 {
	if chunks.closed {
		return chunk.AirRuntimeId
	}
	if !chunks.player.Anticheat().UseChunks() {
		return chunk.AirRuntimeId
	}
	vec3[1] -= 2
	bpos := blocks.BlockPosFromVec3(vec3)
	b := chunks.player.Chunks().GetBlock(bpos)
	if b == chunk.AirRuntimeId {
		var toCheck = []protocol.BlockPos{
			{bpos[0] + 1, bpos[1], bpos[2] + 1},
			{bpos[0] + 1, bpos[1], bpos[2] - 1},
			{bpos[0] + 1, bpos[1], bpos[2]},
			{bpos[0] - 1, bpos[1], bpos[2] - 1},
			{bpos[0] - 1, bpos[1], bpos[2] + 1},
			{bpos[0] - 1, bpos[1], bpos[2]},
			{bpos[0], bpos[1], bpos[2] + 1},
			{bpos[0], bpos[1], bpos[2] - 1},
		}
		for _, c := range toCheck {
			id := chunks.player.Chunks().GetBlock(c)
			if id != chunk.AirRuntimeId {
				return id
			}
		}
	}
	return b
}
