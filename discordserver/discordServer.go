package discordserver

import (
	"context"
	"github.com/aiocean/wireset/server"
	"github.com/bwmarrin/discordgo"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"log"
	"os"
)

var DefaultWireset = wire.NewSet(
	wire.Struct(new(DiscordServer), "*"),
	wire.Bind(new(server.Server), new(*DiscordServer)),
)

type DiscordServer struct {
}

// Start starts the discord server
func (s *DiscordServer) Start(ctx context.Context) chan error {
	errChan := make(chan error, 1)

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		errChan <- errors.New("no token provided")
		return errChan
	}

	appID := os.Getenv("DISCORD_APP_ID")
	if appID == "" {
		errChan <- errors.New("no app ID provided")
		return errChan
	}

	// Create a new Discordgo session
	dg, err := discordgo.New(token)
	if err != nil {
		errChan <- errors.WithMessage(err, "error creating Discord session")
		return errChan
	}

	ap, err := dg.Application(appID)
	log.Printf("ApplicationCreate: err: %+v, app: %+v\n", err, ap)

	return errChan
}
