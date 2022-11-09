package services

import (
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"golang.org/x/xerrors"
	"log"
	"os"
	"server/pkg/app/models"
	"strconv"
	"time"
)

type TransactionService struct {
	dbHost     string
	dbPort     int
	dbUser     string
	dbPassword string
	dbName     string

	queue chan models.Transaction

	db *sqlx.DB
}

func (ts *TransactionService) CreateTransaction(transactionRequest models.TransactionRequest) (models.TransactionRequest, error) {
	err := ts.connect()
	if err != nil {
		return transactionRequest, err
	}

	transaction := models.Transaction.FromTransactionRequest(models.Transaction{}, transactionRequest)
	log.Printf("%s, converted to : %s\n", transactionRequest, transaction)

	t := `
		INSERT INTO "transaction" ("timestamp", "from_balance", "to_balance", "value", "currency") 
		VALUES ($1, $2, $3, $4, $5);
	`
	_, err = ts.db.Exec(
		t,
		transaction.Timestamp.Format("2006-01-02 15:04:05"),
		transaction.FromBalance,
		transaction.ToBalance,
		transaction.Value,
		transaction.Currency,
	)

	if err != nil {
		log.Printf("Added %s to inner pool due to DB error!", transaction)
	}

	return transactionRequest, err
}

func (ts *TransactionService) GetTransaction(id int) (models.Transaction, error) {
	err := ts.connect()
	if err != nil {
		return models.Transaction{}, err
	}

	var transactions []models.Transaction
	t := `SELECT * FROM "transaction" WHERE id=%d;`
	err = ts.db.Select(&transactions, fmt.Sprintf(t, id))

	if len(transactions) == 1 {
		return transactions[0], nil
	}

	if err == nil {
		err = xerrors.New(fmt.Sprintf("No transaction with ID: %d", id))
	}

	return models.Transaction{}, err
}

func (ts *TransactionService) GetAllTransactions() ([]models.Transaction, error) {
	err := ts.connect()
	if err != nil {
		return []models.Transaction{}, err
	}

	var transactions []models.Transaction
	err = ts.db.Select(&transactions, `SELECT * FROM transaction`)

	return transactions, err
}

func (ts *TransactionService) process(statusList []models.TransactionStatus) {
	collectTransactions := func() {
		_ = ts.connect()
		t := `SELECT * FROM "transaction" WHERE status='%s';`

		for _, status := range statusList {
			var transactions []models.Transaction
			_ = ts.db.Select(&transactions, fmt.Sprintf(t, status))
			for _, transaction := range transactions {
				log.Printf("To queue: %s\n", transaction)
				ts.queue <- transaction
			}
		}
	}

	for {
		go collectTransactions()
		go func() {
			for transaction := range ts.queue {
				err := ts.processTransaction(transaction)
				if err != nil {
					log.Println(err.Error())
				}
			}
		}()

		time.Sleep(time.Millisecond * 100)
	}
}

func (ts *TransactionService) updateTransaction(id int, status models.TransactionStatus) error {
	err := ts.connect()
	if err != nil {
		return err
	}

	t := `
		UPDATE "transaction" SET status=$1
  		WHERE id=$2;
	`
	_, err = ts.db.Exec(t, status, id)

	return err
}

func (ts *TransactionService) connect() error {
	if ts.db != nil {
		err := ts.db.Ping()
		if err == nil {
			return nil
		}
	}

	t := `host=%s port=%d user=%s password=%s dbname=%s sslmode=disable`
	connectionString := fmt.Sprintf(
		t,
		ts.dbHost,
		ts.dbPort,
		ts.dbUser,
		ts.dbPassword,
		ts.dbName,
	)

	db, err := sqlx.Open("pgx", connectionString)
	if err == nil {
		ts.db = db
	}

	return err
}

// This method makes me feel upset - I have to deal smh with these inner variances.
// But... :)
func (ts *TransactionService) processTransaction(transaction models.Transaction) error {
	getBalanceById := func(id int) (models.Balance, error) {
		err := ts.connect()
		if err != nil {
			return models.Balance{}, err
		}

		var balances []models.Balance
		t := `SELECT * FROM "balance" WHERE id=%d;`
		err = ts.db.Select(&balances, fmt.Sprintf(t, id))

		if len(balances) == 1 {
			return balances[0], nil
		}
		return models.Balance{}, err
	}

	setBalance := func(id int, balance float64) error {
		err := ts.connect()
		if err != nil {
			return err
		}

		t := `
		UPDATE "balance" SET balance=$1
  		WHERE id=$2;
		`
		_, err = ts.db.Exec(t, balance, id)

		return err
	}

	fromBalance, err := getBalanceById(transaction.FromBalance)
	if err != nil {
		_ = ts.updateTransaction(transaction.Id, models.Cancelled)
		return xerrors.New("Sender balance does not exist!")
	}

	toBalance, err := getBalanceById(transaction.ToBalance)
	if err != nil {
		_ = ts.updateTransaction(transaction.Id, models.Cancelled)
		return xerrors.New("Receiver balance does not exist!")
	}

	// Так читаемее.
	if !(fromBalance.Currency == toBalance.Currency && fromBalance.Currency == transaction.Currency) {
		_ = ts.updateTransaction(transaction.Id, models.Cancelled)
		return xerrors.New("Wrong currency!")
	}

	if fromBalance.Balance < transaction.Value {
		_ = ts.updateTransaction(transaction.Id, models.Cancelled)
		return xerrors.New("Not enough money!")
	}

	err = setBalance(fromBalance.Id, fromBalance.Balance-transaction.Value)
	if err != nil {
		_ = ts.updateTransaction(transaction.Id, models.Failed)
		return err
	}

	err = setBalance(toBalance.Id, toBalance.Balance+transaction.Value)
	if err != nil {
		_ = ts.updateTransaction(transaction.Id, models.Failed)
		return err
	}

	_ = ts.updateTransaction(transaction.Id, models.Success)
	return nil
}

func NewTransactionService() *TransactionService {
	const (
		envHost     string = "PGS_HOST"
		envPort     string = "PGS_PORT"
		envUser     string = "PGS_USER"
		envPassword string = "PGS_PASSWORD"
		envDbName   string = "PGS_DB"
	)

	pgsHost := os.Getenv(envHost)
	if pgsHost == "" {
		msg := fmt.Sprintf("%s was empty.", envHost)
		panic(xerrors.New(msg).Error())
	}
	log.Printf("%s configured with: %s", envHost, pgsHost)

	pgsPort, err := strconv.Atoi(os.Getenv(envPort))
	if err != nil {
		msg := fmt.Sprintf("%s was empty or incorrect", envPort)
		panic(xerrors.New(msg).Error())
	}
	log.Printf("%s configured with: %d", envPort, pgsPort)

	pgsUser := os.Getenv(envUser)
	if pgsUser == "" {
		msg := fmt.Sprintf("%s was empty.", envUser)
		panic(xerrors.New(msg).Error())
	}
	log.Printf("%s configured with: %s", envUser, pgsUser)

	pgsPassword := os.Getenv(envPassword)
	if pgsPassword == "" {
		msg := fmt.Sprintf("%s was empty.", envPassword)
		panic(xerrors.New(msg).Error())
	}
	log.Printf("%s configured with: %s", envPassword, pgsPassword)

	pgsDbName := os.Getenv(envDbName)
	if pgsDbName == "" {
		msg := fmt.Sprintf("%s was empty.", envDbName)
		panic(xerrors.New(msg).Error())
	}
	log.Printf("%s configured with: %s", envDbName, pgsDbName)

	transactionService := new(TransactionService)
	transactionService.dbHost = pgsHost
	transactionService.dbPort = pgsPort
	transactionService.dbUser = pgsUser
	transactionService.dbPassword = pgsPassword
	transactionService.dbName = pgsDbName
	transactionService.queue = make(chan models.Transaction)

	// Process requests
	go transactionService.process([]models.TransactionStatus{models.Failed, models.Created})

	return transactionService
}
