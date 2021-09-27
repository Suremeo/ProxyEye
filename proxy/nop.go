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
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

var NopSession session.Session = &nopSessionStructure{}
var NopServer session.Server = &nopServerStructure{}
var NopListener session.Listener = &nopListenerStructure{}

type nopSessionStructure struct{}

func (n *nopSessionStructure) TranslateEid(now uint64) uint64 {
	return now
}

func (n *nopSessionStructure) TranslateUid(now int64) int64 {
	return now
}

func (n *nopSessionStructure) WritePacket(packet packet.Packet) error {
	return nil
}

func (n *nopSessionStructure) Type() session.SourceType {
	return session.SourceTypeSession
}

func (n *nopSessionStructure) OldRuntimeId() uint64 {
	return 0
}

func (n *nopSessionStructure) NewRuntimeId() uint64 {
	return 0
}

func (n *nopSessionStructure) OldUniqueId() int64 {
	return 0
}

func (n *nopSessionStructure) NewUniqueId() int64 {
	return 0
}

func (n *nopSessionStructure) Server() session.Server {
	return NopServer
}

func (n *nopSessionStructure) Packet(source session.Source, destination session.Destination, pk packet.Packet) error {
	return nil
}

func (n *nopSessionStructure) Close() {}

func (n *nopSessionStructure) SetConnected() {}

func (n *nopSessionStructure) IsConnected() bool {
	return false
}

type nopServerStructure struct{}

func (n *nopServerStructure) Players() map[uint64]session.Player {
	panic("implement me")
}

func (n *nopServerStructure) Joinable() bool {
	return true
}

func (n *nopServerStructure) SetJoinable(b bool) {
}

func (n *nopServerStructure) Count() int {
	return 0
}

func (n *nopServerStructure) Address() string {
	return ""
}

func (n *nopServerStructure) Raknet() bool {
	return false
}

func (n *nopServerStructure) Connect(player session.Player) error {
	return nil
}

func (n *nopServerStructure) Leave(player session.Player, s session.Session) {
}

func (n *nopServerStructure) PlayerByRuntimeId(u uint64) (session.Player, bool) {
	return nil, false
}

func (n *nopServerStructure) Packet(source session.Source, player session.Player, p packet.Packet) bool {
	return false
}

func (n *nopServerStructure) Listener() session.Listener {
	return NopListener
}

func (n *nopServerStructure) SetListener(listener session.Listener) {}

type nopListenerStructure struct{}

func (n *nopListenerStructure) Leave(player session.Player) {
}

func (n *nopListenerStructure) Packet(source session.Source, player session.Player, p packet.Packet) {
}
