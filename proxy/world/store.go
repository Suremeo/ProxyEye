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
	"sync"
)

type Store struct {
	mutex  sync.Mutex
	worlds map[string]*World
}

func NewStore() *Store {
	return &Store{
		worlds: make(map[string]*World),
	}
}

func (store *Store) GetOrCreate(name string) *World {
	store.mutex.Lock()
	world, ok := store.worlds[name]
	if ok {
		store.mutex.Unlock()
		return world
	}
	w := newWorld(store, name)
	store.worlds[name] = w
	store.mutex.Unlock()
	return w
}

func (store *Store) Delete(name string) {
	delete(store.worlds, name)
}

func (store *Store) Import(world *World) {
	store.mutex.Lock()
	store.worlds[world.name] = world
	store.mutex.Unlock()
}
