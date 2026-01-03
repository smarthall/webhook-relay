package cmd

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/smarthall/webhook-relay/internal/messaging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	receiverCmd.Flags().String("listen", ":8080", "Address to listen on")
	viper.BindPFlag("listen", receiverCmd.Flags().Lookup("listen"))

	rootCmd.AddCommand(receiverCmd)
}

var receiverCmd = &cobra.Command{
	Use:   "receiver",
	Short: "Receives webhooks and forwards them to RabbitMQ",
	Long:  `Receiver listens for incoming webhooks and forwards them to a RabbitMQ exchange.`,
	Run: func(cmd *cobra.Command, args []string) {
		pub := messaging.NewPublisher(viper.GetString("amqp"))

		// Create and start a health checker. If the health checker signals
		// failure the process will gracefully shut down.
		hc := messaging.NewHealthChecker(viper.GetString("amqp"), 1*time.Second, 2*time.Second)
		defer hc.Stop()

		// Create a context that is cancelled on SIGINT or SIGTERM
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		s := &http.Server{
			Addr:           viper.GetString("listen"),
			Handler:        requestHandler(pub),
			ReadTimeout:    2 * time.Second,
			WriteTimeout:   2 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		// Start server in a goroutine so we can listen for shutdown signals.
		go func() {
			log.Printf("starting server on %s", s.Addr)
			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("server error: %v", err)
			}
		}()

		// Monitor healthcheck failures and trigger shutdown when they occur.
		go func() {
			select {
			case <-ctx.Done():
				return
			case <-hc.Failure:
				log.Printf("healthcheck failure detected, initiating shutdown")
				stop()
			}
		}()

		// Wait for signal
		<-ctx.Done()
		log.Printf("shutdown signal received, shutting down server")

		// Allow up to 10 seconds for graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown error: %v", err)
		} else {
			log.Printf("server stopped")
		}
	},
}

// requestHandler returns an http.HandlerFunc that publishes incoming requests
// using the provided publisher. The publisher is expected to implement
// Publish(http.Request) error.
func requestHandler(pub interface{ Publish(http.Request) error }) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request at: %s", r.URL.Path)

		if err := pub.Publish(*r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
