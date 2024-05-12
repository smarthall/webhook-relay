package cmd

import (
	"log"
	"net/http"
	"time"

	"github.com/smarthall/webhook-relay/internal/messaging"
	"github.com/spf13/cobra"
)

var listenAddr string

func init() {
	receiverCmd.Flags().StringVarP(&listenAddr, "listen", "", ":8080", "Address to listen on")
	rootCmd.AddCommand(receiverCmd)
}

var receiverCmd = &cobra.Command{
	Use:   "receiver",
	Short: "Receives webhooks and forwards them to RabbitMQ",
	Long:  `Receiver listens for incoming webhooks and forwards them to a RabbitMQ exchange.`,
	Run: func(cmd *cobra.Command, args []string) {
		var pub = messaging.NewPublisher(amqpURI)

		s := &http.Server{
			Addr: listenAddr,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Publish the request to the message broker
				err := pub.Publish(*r)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// If the message was published successfully, return a 204 No Content response
				w.WriteHeader(http.StatusNoContent)
			}),
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		log.Fatal(s.ListenAndServe())
	},
}
