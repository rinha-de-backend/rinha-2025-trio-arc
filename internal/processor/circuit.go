/*
	Contém a lógica pura do disjuntor, a máquina de estado que é agnóstica a HTTP.
*/

package processor

import (
	"errors"
	"sync"
	"time"
)

const (
	StateClosed   = iota
	StateOpen     = iota
	StateHalfOpen = iota
)

// Máquina de estado do Circuit Breaker
type CircuitBreaker struct {
	mutex               sync.Mutex
	state               int
	consecutiveFailures int
	maxFailures         int
	openStateTimeout    time.Duration
	lastFailureTime     time.Time
}

// NewCircuitBreaker cria uma nova instância do disjuntor
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		maxFailures:      maxFailures,
		openStateTimeout: timeout,
	}
}

// Envolve a chamada de trabalho com a lógica do disjuntor
func (cb *CircuitBreaker) Execute(work func() error) error {
	if !cb.canExecute() {
		return errors.New("Circuit breaker is open")
	}

	err := work()
	cb.afterExecute(err)

	return err
}

// FUNÇÔES PARA CONTROLE DO ESTADO

func (cb *CircuitBreaker) canExecute() bool {

	cb.mutex.Lock()         // Trava o mutex para proteger o estado
	defer cb.mutex.Unlock() // Garante que o mutex será liberado

	if cb.state == StateOpen { // Se estiver aberto, verifica se o tempo de espera expirou
		if time.Since(cb.lastFailureTime) > cb.openStateTimeout { // Se expirou, muda para meio-aberto
			cb.state = StateHalfOpen
			return true // Permite uma tentativa
		}
		return false // Ainda está aberto, não permite
	}
	return true // Se estiver fechado ou meio-aberto, permite
}

// Atualiza o estado do disjuntor após a execução do trabalho
func (cb *CircuitBreaker) afterExecute(err error) {
	cb.mutex.Lock()         // Trava o mutex para proteger o estado
	defer cb.mutex.Unlock() // Garante que o mutex será liberado

	if err == nil {
		cb.reset() // Sucesso, reseta o disjuntor
		return
	}

	// Falha
	cb.consecutiveFailures++
	if cb.state == StateHalfOpen || cb.consecutiveFailures > cb.maxFailures {
		cb.open() // Se meio-aberto ou excedeu falhas, abre o disjuntor
	}
}

// MÉTODOS PARA MONITORAR SAÚDE

// Permite que um agente externo, como health.go , abra o circuito
func (cb *CircuitBreaker) ForceOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.open()
}

// Permite que um agente externo feche o circuito e zere as falhas
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.reset()
}

// FUNÇÕES PRIVADAS PARA MANIPULAR ESTADO
func (cb *CircuitBreaker) reset() {
	cb.state = StateClosed
	cb.consecutiveFailures = 0
}

func (cb *CircuitBreaker) open() {
	cb.state = StateOpen
	cb.lastFailureTime = time.Now()
}
