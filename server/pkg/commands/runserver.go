package commands

import (
	"github.com/spf13/cobra"
	"log"
	"server/pkg/app"
)

func init() {
	MainCmd.PersistentFlags().String("host", "localhost", "server host")
	MainCmd.PersistentFlags().Int("port", 8000, "server port")
}

func Runserver(cmd *cobra.Command, _ []string) {
	host, err := cmd.PersistentFlags().GetString("host")
	if err != nil {
		log.Fatalln(err.Error())
	}

	port, err := cmd.PersistentFlags().GetInt("port")
	if err != nil {
		log.Fatalln(err.Error())
	}

	transactionServer := app.NewTransactionServer(host, port)
	transactionServer.Run()
}

var MainCmd = &cobra.Command{
	Use:   "runserver",
	Short: "Run server with MM",
	Long:  "Run server with monitor manager",
	Run:   Runserver,
}
