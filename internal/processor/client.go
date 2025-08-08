package processor

import (
	"/internal/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"internal/models"
	"net"
	"net/http"
	"time"
)

// PaymentProcessorClient encapsula a lógica para se comunicar com um processador
type PaymentProcessorClient struct {
	httpClient *http.Client
	baseURL    string
	logger     *config.Logger
}

const TIMEOUT_DURATION = 5 * time.Second
const MAX_IDLE_CONNECTIONS = 100
const MAX_IDLE_CONNECTIONS_PER_HOST = 100
const IDLE_CONNECTION_TIMEOUT = 90 * time.Second

const POST_PROCESS_PAYMENT = "/payments"

// NewPaymentProcessorClient cria uma nova instância do cliente.
func NewPaymentProcessorClient(baseURL string, logger *config.Logger) *PaymentProcessorClient {
	// A configuração do http.Transport para performance e resiliência.
	transport := &http.Transport{
		// Timeout para estabelecer a conexão TCP.
		DialContext: (&net.Dialer{
			Timeout: TIMEOUT_DURATION,
		}).DialContext,

		// Timeout para o handshake TLS (se usarmos HTTPS).
		TLSHandshakeTimeout: TIMEOUT_DURATION,
		// Timeout para esperar os cabeçalhos da resposta após enviar a requisição.
		ResponseHeaderTimeout: TIMEOUT_DURATION,
		// Limite de conexões ociosas (reutilizáveis) no pool. Essencial para performance.
		MaxIdleConns: MAX_IDLE_CONNECTIONS,
		// Limite de conexões ociosas por host.
		MaxIdleConnsPerHost: MAX_IDLE_CONNECTIONS_PER_HOST,
		// Tempo que uma conexão ociosa fica no pool antes de ser fechada.
		IdleConnTimeout: IDLE_CONNECTION_TIMEOUT,
	}

	return &PaymentProcessorClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: transport,

			// Timeout total para a requisição inteira. Uma salvaguarda final.
			Timeout: 15 * time.Second,
		},
		logger: logger,
	}
}

// ProcessPayment para enviar uma requisição de pagamento para o processador.
func (c *PaymentProcessorClient) ProcessPayment(ctx context.Context, payment *models.Payment) error {

	c.logger.Infof("Iniciando o processamento do pagamento: %s", payment.CorrelationID)

	// Montar o corpo da requisição usando o modelo PaymentProcessorRequest.
	payload := payment.ToProcessorRequest()

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		c.logger.Error("ERRO ao serializar o payload de pagamento para JSON:", err)
		return fmt.Errorf("ERRO ao hidratar o payload de pagamento para JSON: %w", err)
	}

	// Cria a requisição HTTP
	url := c.baseURL + POST_PROCESS_PAYMENT
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		c.logger.Error("ERRO ao criar a requisição HTTP:", err)
		return fmt.Errorf("ERRO ao criar a requisição HTTP: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Executa a requisição
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("ERRO ao enviar a requisição HTTP:", err)
		return fmt.Errorf("ERRO ao enviar a requisição HTTP: %w", err)
	}
	defer resp.Body.Close()

	// Valida a resposta
	if resp.StatusCode >= 400 {
		c.logger.Warningf("Processador em %s respondeu com status de erro %d para o pagamento %s", c.baseURL, resp.StatusCode, payment.CorrelationID)
		return fmt.Errorf("processador respondeu com erro: status %d", resp.StatusCode)
	}

	c.logger.Infof("Pagamento %s processado com sucesso em %s", payment.CorrelationID, c.baseURL)
	return nil
}
