package app

import (
	"fmt"
	"log"
	"net/http"
	"server/pkg/app/middleware"
	"server/pkg/app/services"
)

type TransactionServer struct {
	host    string
	port    int
	mux     *http.ServeMux
	service *services.TransactionService
}

func (trServer *TransactionServer) configureRoutes() {
	trServer.mux.HandleFunc("/apiv1/transaction/", trServer.transactionHandler)
}

func (trServer *TransactionServer) Domain() string {
	return fmt.Sprintf("%s:%d", trServer.host, trServer.port)
}

func (trServer *TransactionServer) Run() {
	// Multiplexer
	trServer.mux = http.NewServeMux()

	// Configuring routes
	trServer.configureRoutes()

	// Configuring middleware
	handler := middleware.LoggingMiddleware(trServer.mux)

	log.Fatal(http.ListenAndServe(trServer.Domain(), handler))
}

func NewTransactionServer(host string, port int) *TransactionServer {
	return &TransactionServer{host: host, port: port, service: services.NewTransactionService()}
}

// Type check
var _ Application = (*TransactionServer)(nil)
