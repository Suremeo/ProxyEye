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

package main

import (
	"github.com/Suremeo/ProxyEye/proxy"
	"github.com/Suremeo/ProxyEye/proxy/console"
	"github.com/Suremeo/ProxyEye/proxy/session"
	"github.com/Suremeo/ProxyEye/proxy/session/anticheat"
	"github.com/Suremeo/ProxyEye/proxy/session/events"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"os"
	"runtime/pprof"
	"strconv"
	"time"
)

func main() {
	proxy.Eye.HandleEvent(events.JoinEventHandler(func(event *events.Context, player session.Player) {
		console.Info("Player connected: %v", player.Raknet().IdentityData().DisplayName)
	}))
	proxy.Eye.HandleEvent(events.LeaveEventHandler(func(event *events.Context, player session.Player) {
		console.Warn("Player left: %v", player.Raknet().IdentityData().DisplayName)
	}))
	one := proxy.NewRemoteServer("127.0.0.1:1001", 100)
	two := proxy.NewInternalServer()
	proxy.Eye.HandleEvent(events.ConnectEventHandler(func(event *events.Context, player session.Player) {
		console.Debug("%v connected to %v", player.Raknet().IdentityData().DisplayName, player.Session().Server().Address())
		time.Sleep(30 * time.Second)
		if player.Session().Server().Address() == one.Address() {
			_ = two.Connect(player)
		} else {
			_ = one.Connect(player)
		}
	}))
	proxy.Eye.HandleEvent(events.AnticheatDetectionEventHandler(func(event *events.Context, player session.Player, detection *anticheat.Detection) {
		console.Debug("%v is suspected of using %v", player.Raknet().IdentityData().DisplayName, detection.Type)
	}))
	proxy.Eye.HandleEvent(events.PacketEventHandler(func(event *events.Context, player session.Player, source session.Source, pk packet.Packet) {

	}))
	go func() {
		i := 0
		t := time.NewTicker(1 * time.Minute)
		for range t.C {
			i++
			c, err := os.Create("./heaps/" + strconv.Itoa(i) + "-heap.pprof")
			if err != nil {
				console.Error("Failed to create heap file", err)
			} else {
				err = pprof.WriteHeapProfile(c)
				if err != nil {
					console.Error("Failed to write to heap file", err)
				} else {
					console.Info("Wrote heap %v", i)
				}
			}
		}
	}()
	console.Fatal("Error starting proxy", proxy.Eye.Run())
}
