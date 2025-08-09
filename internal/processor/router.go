/*
	O router não é um "roteador HTTP", e sim um roteador de estratégia de processamento.
	Sua responsabilidade é receber um pagamento e orquestrar as tentativas de processa-lo
	através dos canais "default" e "fallback", usando a lógica do circuit breaker.
*/

package processor

import (
	"context"
	"fmt"
	"internal/config"
	"models/payment"
)

type Processor struct {
	Type   payment.ProcessorType
	Client *PaymentProcessorClient
	CB     *CircuitBreaker
}

type Router struct {
	DefaultProcessor  *Processor
	FallbackProcessor *Processor
	logger            *config.Logger
}

func NewRouter(dp, fp *Processor, logger *config.Logger) *Router {
	return &Router{
		DefaultProcessor:  dp,
		FallbackProcessor: fp,
		logger:           logger,
	}
}

// Implementa a estratégia de roteamento principal.
// Tenta o default, se falhar, o 'fallback'
func (r *Router) RoutePayment(ctx context.Context, payment *payment.Payment){
	r.logger.Infof("Roteando pagamento %s", payment.CorrelationID)

	// Tentativa 1: DEFAULT
	workDefault := func() error {
		return r.DefaultProcessor.Client.ProcessPayment(ctx, payment)
	}

	// Executa o trabalho através do Circuit Breaker do processador default
	errDefault := r.DefaultProcessor.CB.Execute(workDefault)

	// Sucesso
	if errDefault == nil {
		payment.MarkAsProcessed(r.DefaultProcessor.Type)
		r.logger.Infof("Pagamento %s processado com sucesso via %s", payment.CorrelationID, r.DefaultProcessor.Type)
		return 
	}

	// Se chegar até aqui, a tentativa com o default falhou
	r.logger.Warningf(
		"Falha ao processar pagamento %s no canal %s. Erro %v. Tentando Fallback...",
		payment.CorrelationID, r.DefaultProcessor.Type, errDefault,
	)

	// Tentativa 2: FALLBACK
	workFallback := func() error {
		return r.FallbackProcessor.Client.ProcessPayment(ctx, payment)
	}

	// Executa o trabalho através do Circuit Breaker do processador fallback
	errFallback := r.FallbackProcessor.CB.Execute(workFallback)

	// Sucesso
	if errFallback == nil {
		payment.MarkAsProcessed(r.FallbackProcessor.Type)
		r.logger.Infof("Pagamento %s processado com sucesso via %s", payment.CorrelationID, r.FallbackProcessor.Type)
		return 
	}

	// Se chegar até aqui, AMBOS falharam. Falha definitiva
	finalErrorMsg := fmt.Sprintf(
		"Default falhou [%v]; Fallback também falhou: [%v]",
		errDefault, errFallback,
	)
	payment.MarkAsProcessed(finalErrorMsg)
	r.logger.Errorf(
		"FALHA DEFINITIVA para o pagamento %s: %s",
		payment.CorrelationID, finalErrorMsg,
	)
}