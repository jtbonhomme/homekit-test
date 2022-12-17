package main

import (
	"time"

	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	"github.com/rs/zerolog"

	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	logger := zerolog.New(output).With().Timestamp().Logger()
	logger.Info().Msg("Homekit demo starting...")

	// Create the testing accessory.
	a := accessory.NewSwitch(accessory.Info{
		Name: "MBP-DEMO",
	})
	a.Switch.On.OnValueRemoteUpdate(func(on bool) {
		if on == true {
			logger.Info().Msg("Switch is on")
		} else {
			logger.Info().Msg("Switch is off")
		}
	})
	logger.Info().Msgf("Homekit demo new accessory created: %s", a.Name())

	// Store the data in the "./db" directory.
	fs := hap.NewFsStore("./db")
	logger.Info().Msg("Homekit demo db created")

	// Create the hap server.
	server, err := hap.NewServer(fs, a.A)
	if err != nil {
		// stop if an error happens
		logger.Panic().Msgf("failed while creating new HAP server: %s", err.Error())
	}
	server.Pin = "00102003" // default pincode

	logger.Info().Msgf("Homekit demo server started (%s)", server.Pin)

	// Setup a listener for interrupts and SIGTERM signals
	// to stop the server.
	logger.Info().Msg("Waiting for SIGTERM signal ...")
	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-interrupt
		logger.Info().Msg("\nReceived SIGTERM signal !\n")
		// Stop delivering signals.
		signal.Stop(interrupt)
		// Cancel the context to stop the server.
		cancel()
	}()

	// Run the server.
	logger.Info().Msg("Listening ...")
	server.ListenAndServe(ctx)
	logger.Info().Msg("Exit")
}
