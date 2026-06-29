// Package cache provee un cache en memoria genérico con TTL y soporte de datos
// "stale". Los handlers lo usan para no golpear las APIs externas (Proxmox,
// Arcane, wttr.in) en cada request del frontend, que hace polling cada pocos
// segundos.
package cache

import (
	"sync"
	"time"
)

// Entry[T] guarda un valor con su tiempo de expiración.
type Entry[T any] struct {
	mu        sync.RWMutex
	value     T
	hasValue  bool
	expiresAt time.Time
	ttl       time.Duration
}

// New crea una entrada de cache con el TTL indicado.
func New[T any](ttl time.Duration) *Entry[T] {
	return &Entry[T]{ttl: ttl}
}

// Get devuelve el valor cacheado, si está fresco (no expirado).
// El segundo retorno indica si el valor es válido y fresco.
func (e *Entry[T]) Get() (T, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if !e.hasValue || time.Now().After(e.expiresAt) {
		return e.value, false
	}
	return e.value, true
}

// Stale devuelve el último valor conocido aunque esté expirado. Se usa como
// fallback cuando la API externa falla. El bool indica si hay algún valor.
func (e *Entry[T]) Stale() (T, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.value, e.hasValue
}

// Set almacena un nuevo valor y reinicia el TTL.
func (e *Entry[T]) Set(value T) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.value = value
	e.hasValue = true
	e.expiresAt = time.Now().Add(e.ttl)
}

// Invalidate marca la entrada como expirada para forzar un refetch en el
// siguiente Get. Se usa tras una mutación (p. ej. start/stop de un contenedor).
func (e *Entry[T]) Invalidate() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.expiresAt = time.Time{}
}
