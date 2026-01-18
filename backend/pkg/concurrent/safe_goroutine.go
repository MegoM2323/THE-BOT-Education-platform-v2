package concurrent

import (
	"github.com/rs/zerolog/log"
	"runtime/debug"
)

// SafeGo запускает функцию в goroutine с обработкой паник
// При panic логирует ошибку и stack trace вместо краша приложения
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Info().Interface("panic", r).Bytes("stack", debug.Stack()).Msg("Recovered from panic in goroutine")
			}
		}()
		fn()
	}()
}
