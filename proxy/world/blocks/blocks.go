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

package blocks

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/Suremeo/ProxyEye/proxy/world/chunk"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"math"
	"sort"
	"sync"
	"unsafe"
)

func BlockPosFromVec3(vec3 mgl32.Vec3) protocol.BlockPos {
	return protocol.BlockPos{int32(math.Floor(float64(vec3[0]))), int32(math.Floor(float64(vec3[1]))), int32(math.Floor(float64(vec3[2])))}
}

type BlockState struct {
	Name       string                 `nbt:"name"`
	Properties map[string]interface{} `nbt:"states"`
	Version    int32                  `nbt:"version"`
}

var buffers = sync.Pool{New: func() interface{} {
	return bytes.NewBuffer(make([]byte, 0, 128))
}}

func (s BlockState) HashProperties() string {
	if s.Properties == nil {
		return ""
	}
	keys := make([]string, 0, len(s.Properties))
	for k := range s.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b := buffers.Get().(*bytes.Buffer)
	for _, k := range keys {
		switch v := s.Properties[k].(type) {
		case bool:
			if v {
				b.WriteByte(1)
			} else {
				b.WriteByte(0)
			}
		case uint8:
			b.WriteByte(v)
		case int32:
			a := uint32(v)
			b.WriteByte(byte(a))
			b.WriteByte(byte(a >> 8))
			b.WriteByte(byte(a >> 16))
			b.WriteByte(byte(a >> 24))
		case string:
			b.WriteString(v)
		}
	}

	data := append([]byte(nil), b.Bytes()...)
	b.Reset()
	buffers.Put(b)
	return *(*string)(unsafe.Pointer(&data))
}

type stateHash struct {
	name, properties string
}

var states []BlockState
var stateRuntimeIDs = map[stateHash]uint32{}

func RegisterBlockState(s BlockState) error {
	h := stateHash{name: s.Name, properties: s.HashProperties()}
	if _, ok := stateRuntimeIDs[h]; ok {
		return fmt.Errorf("cannot register the same state twice (%+v)", s)
	}
	rid := uint32(len(states))
	stateRuntimeIDs[h] = rid
	states = append(states, s)
	return nil
}

func init() {
	b, _ := base64.StdEncoding.DecodeString(BlockStates)
	dec := nbt.NewDecoder(bytes.NewBuffer(b))

	var s BlockState
	for {
		if err := dec.Decode(&s); err != nil {
			break
		}
		if err := RegisterBlockState(s); err != nil {
			// Should never happen.
			panic("duplicate block state registered")
		}
	}

	chunk.RuntimeIDToState = func(runtimeID uint32) (name string, properties map[string]interface{}, found bool) {
		if runtimeID >= uint32(len(states)) {
			return "", nil, false
		}
		s := states[runtimeID]
		return s.Name, s.Properties, true
	}
	chunk.StateToRuntimeID = func(name string, properties map[string]interface{}) (runtimeID uint32, found bool) {
		s := BlockState{Name: name, Properties: properties}
		h := stateHash{name: name, properties: s.HashProperties()}

		rid, ok := stateRuntimeIDs[h]
		return rid, ok
	}
}
