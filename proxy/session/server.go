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

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

type Server interface {
	Joinable() bool
	SetJoinable(bool)
	Count() int
	Address() string
	Raknet() bool
	Connect(Player) error
	Leave(Player, Session)
	PlayerByRuntimeId(uint64) (Player, bool)
	Players() map[uint64]Player
	Packet(Source, Player, packet.Packet) bool
	SetListener(Listener)
	Listener() Listener
}
