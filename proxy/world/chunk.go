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

package world

import (
	"bytes"
	"github.com/Suremeo/ProxyEye/proxy/world/chunk"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"sync"
)

type Chunk struct {
	*chunk.Chunk
	mutex   sync.Mutex
	world   *World
	Pos     chunk.Pos
	viewers map[uint64]struct{}
}

func (c *Chunk) AddViewer(id uint64) {
	c.mutex.Lock()
	c.viewers[id] = struct{}{}
	c.mutex.Unlock()
}

func (c *Chunk) RemoveViewer(id uint64) {
	c.mutex.Lock()
	delete(c.viewers, id)
	if len(c.viewers) == 0 {
		c.world.removeChunk(c.Pos)
	}
	c.mutex.Unlock()
}

func (c *Chunk) SendTo(conn *minecraft.Conn) error {
	var chunkBuf bytes.Buffer
	data := chunk.NetworkEncode(c.Chunk)

	count := byte(0)
	for y := byte(0); y < 16; y++ {
		if data.SubChunks[y] != nil {
			count = y + 1
		}
	}
	for y := byte(0); y < count; y++ {
		if data.SubChunks[y] == nil {
			_ = chunkBuf.WriteByte(chunk.SubChunkVersion)
			// We write zero here, meaning the sub chunk has no block storages: The sub chunk is completely
			// empty.
			_ = chunkBuf.WriteByte(0)
			continue
		}
		_, _ = chunkBuf.Write(data.SubChunks[y])
	}
	_, _ = chunkBuf.Write(data.Data2D)
	_, _ = chunkBuf.Write(data.BlockNBT)

	return conn.WritePacket(&packet.LevelChunk{
		ChunkX:        c.Pos[0],
		ChunkZ:        c.Pos[1],
		SubChunkCount: uint32(count),
		RawPayload:    append([]byte(nil), chunkBuf.Bytes()...),
	})
}
