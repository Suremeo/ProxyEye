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

package session

import "github.com/go-gl/mathgl/mgl32"

// AntiCheatProfile is a anticheat profile for a player connected to the proxy.
type AntiCheatProfile interface {
	// Player returns the Player associated with the AntiCheatProfile.
	Player() Player

	// Yaw ...
	Yaw() float32

	// Pitch ...
	Pitch() float32

	// Facing returns the direction the Player is facing.
	Facing() Direction

	// Position returns the current position of the Player.
	Position() mgl32.Vec3

	// UseChunks returns whether or not the AntiCheatProfile will use chunk related detection features.
	UseChunks() bool

	// Flags returns the packet.AdventureSettings flags for the AntiCheatProfile.
	Flags() uint32

	// Gamemode ...
	Gamemode() int32

	// Teleport is called when the server moves the Player.
	Teleport(pos mgl32.Vec3, pitch, yaw float32)

	// Reset resets all values and Teleports the Player.
	Reset(pos mgl32.Vec3, pitch, yaw float32)

	// SetGamemode ...
	SetGamemode(int32)

	// SetFlags sets the packet.AdventureSettings flags for the AntiCheatProfile.
	SetFlags(uint32)

	// SetUseChunks sets whether or not the AntiCheatProfile will use chunk related detection features.
	SetUseChunks(bool)

	// Close closes the AntiCheatProfile.
	Close()

	// UpdateFlags updates and verifies the packet.AdventureSettings flags for the client.
	UpdateFlags(uint32)

	// Move is called when the Player moves.
	Move(pos mgl32.Vec3, pitch, yaw float32)

	// Attack is called when the Player attacks another Player.
	Attack(Player)
}
