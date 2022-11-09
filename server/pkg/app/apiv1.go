package app

import (
	"encoding/json"
	"fmt"
	"golang.org/x/xerrors"
	"log"
	"net/http"
	"server/pkg/app/models"
	"server/pkg/app/responses"
	"strconv"
	"strings"
)

func (trServer *TransactionServer) createTransactionHandler(writer http.ResponseWriter, request *http.Request) error {
	var transactionRequest models.TransactionRequest
	err := json.NewDecoder(request.Body).Decode(&transactionRequest)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return err
	}

	log.Printf("Incoming transaction request: %s\n", transactionRequest)
	transactionRequest, err = trServer.service.CreateTransaction(transactionRequest)

	responses.JsonResponse(
		writer,
		responses.OkResponseStructure{
			Message: "ok",
			Data:    transactionRequest,
		},
	)
	return err
}

func (trServer *TransactionServer) getAllTransactionsHandler(writer http.ResponseWriter, request *http.Request) error {
	transactions, err := trServer.service.GetAllTransactions()

	if err == nil {
		responses.JsonResponse(writer, transactions)
	} else {
		responses.JsonResponse(writer, transactions, http.StatusNoContent)
	}

	return err
}

func (trServer *TransactionServer) getTransactionHandler(writer http.ResponseWriter, request *http.Request) error {
	uriSections := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
	if len(uriSections) != 3 {
		message := fmt.Sprintf("Incorrect URL: %s!", request.URL.Path)
		http.Error(writer, message, http.StatusNotFound)
		return xerrors.New(message)
	}

	transactionId, err := strconv.Atoi(uriSections[2])
	if err != nil {
		http.Error(writer, err.Error(), http.StatusNotFound)
		return err
	}

	transaction, err := trServer.service.GetTransaction(transactionId)

	if err == nil {
		responses.JsonResponse(writer, transaction)
	} else {
		responses.JsonResponse(writer, transaction, http.StatusNoContent)
	}

	return err
}

func (trServer *TransactionServer) transactionHandler(writer http.ResponseWriter, request *http.Request) {
	var err error
	if request.URL.Path == "/apiv1/transaction/" {
		switch request.Method {
		case http.MethodPost:
			err = trServer.createTransactionHandler(writer, request)
		case http.MethodGet:
			err = trServer.getAllTransactionsHandler(writer, request)
		default:
			http.Error(writer, fmt.Sprintf("Wrong method: %s!", request.Method), http.StatusMethodNotAllowed)
		}
	} else {
		// Seems we have path like /transaction/<id>
		switch request.Method {
		case http.MethodGet:
			err = trServer.getTransactionHandler(writer, request)
		default:
			http.Error(writer, fmt.Sprintf("Wrong method: %s!", request.Method), http.StatusMethodNotAllowed)
		}
	}

	// Logging error if not NULL
	if err != nil {
		log.Println(err.Error())
	}
}
