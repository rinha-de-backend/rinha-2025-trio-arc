/*
	Conterá o monitor ativo, que roda em background, chama o endpoint de health real e atua sobre o Circuit Breaker.
*/

package processor

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Estrutura json retornada pelo endpoint /payments/service-health
type HealthResponse struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}

const TICKER_INTERVAL = 5500 * time.Millisecond // Intervalo de 5.5 segundos
const HTTP_TIMEOUT = 4 * time.Second            // Timeout de 4 segundos

// Inicia uma goroutine para monitorar continuamente a saúde do endpoint
// Atua sobre o Circuit Breaker isolado
func StartMonitor(cb *CircuitBreaker, healthCheckURL string) {
	log.Printf("Iniciando monitoramento de saúde do endpoint: %s", healthCheckURL)

	// Criando um ticker, responsável por chamar o endpoint periodicamente, com um intervalo de 5.5 segundos
	// Aqui é importante lembrar que devemos usar um intervalo um pouco maior para respeitar o rate limit
	ticker := time.NewTicker(TICKER_INTERVAL)

	// Cria um cliente HTTP com um timeout curto, como boa prática
	httpClient := &http.Client{
		Timeout: HTTP_TIMEOUT,
	}

	// Loop infinito que executa a cada vez que o ticker dispara
	for range ticker.C {
		resp, err := httpClient.Get(healthCheckURL)

		if err != nil {
			// Se a chamada ao endpoint health falhar, consideramos que o serviço está indisponível
			log.Printf("ERRO no monitor de saúde para %s: %v. Forçando abertura do circuito.", healthCheckURL, err)

			cb.ForceOpen() // Força a abertura do circuito
			continue
		}
		defer resp.Body.Close()

		var healthResp HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
			// Erro caso a resposta não seja um JSON válido
			log.Printf("ERRO ao decodificar resposta de saúde de %s: %v. Forçando abertura do circuito.", healthCheckURL, err)

			cb.ForceOpen() // Força a abertura do circuito
			continue
		}

		// [LÓGICA PRINCIPAL] - Atuar sobre o Circuit Breaker com base na resposta
		if healthResp.Failing {
			log.Printf("Monitor de saúde detectou FALHA em %s. Forçando abertura do circuito.", healthCheckURL)
			cb.ForceOpen() // Força a abertura do circuito
		} else {
			log.Printf("Monitor de saúde detectou SUCESSO em %s. Resetando o circuito.", healthCheckURL)
			cb.Reset() // Reseta o Circuit Breaker se o serviço estiver saudável
		}
	}
}
