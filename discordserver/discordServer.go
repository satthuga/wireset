package discordserver

import (
	"context"
	"fmt"
	"github.com/aiocean/wireset/server"
	"github.com/bwmarrin/discordgo"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// DefaultWireset is a wire provider set that provides a DiscordServer
var DefaultWireset = wire.NewSet(
	wire.Struct(new(DiscordServer), "*"),
	wire.Bind(new(server.Server), new(*DiscordServer)),
)

type Config struct {
	Token string
	AppID string
}

type DiscordServer struct {
	Config Config
	Logger *zap.Logger
}

func NewDiscordServer(config Config, logger *zap.Logger) *DiscordServer {
	return &DiscordServer{
		Config: config,
		Logger: logger,
	}
}

// Start starts the discord server
func (s *DiscordServer) Start(ctx context.Context) chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)

		if s.Config.Token == "" {
			errChan <- errors.New("no token provided")
			return
		}

		if s.Config.AppID == "" {
			errChan <- errors.New("no app ID provided")
			return
		}

		// Create a new Discordgo session
		dg, err := discordgo.New("Bot " + s.Config.Token)
		if err != nil {
			errChan <- errors.WithMessage(err, "error creating Discord session")
			return
		}

		// Register handlers
		dg.AddHandler(s.messageCreate)

		// Open the websocket connection
		err = dg.Open()
		if err != nil {
			errChan <- errors.WithMessage(err, "error opening connection")
			return
		}

		s.Logger.Info("Discord bot is now running. Press CTRL-C to exit.")

		// Wait for context cancellation or system interrupt
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-ctx.Done():
			s.Logger.Info("Context cancelled, shutting down...")
		case sig := <-sigChan:
			s.Logger.Info(fmt.Sprintf("Received signal %v, shutting down...", sig))
		}

		// Graceful shutdown
		err = s.shutdown(dg)
		if err != nil {
			errChan <- err
		}
	}()

	return errChan
}

func (s *DiscordServer) shutdown(session *discordgo.Session) error {
	s.Logger.Info("Shutting down Discord connection...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- session.Close()
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return errors.WithMessage(err, "error closing Discord session")
		}
		s.Logger.Info("Discord connection closed successfully")
	case <-ctx.Done():
		return errors.New("shutdown timed out")
	}

	return nil
}

func (s *DiscordServer) messageCreate(session *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages created by the bot itself
	if m.Author.ID == session.State.User.ID {
		return
	}

	// Handle commands here
	if m.Content == "!ping" {
		_, err := session.ChannelMessageSend(m.ChannelID, "Pong!")
		if err != nil {
			s.Logger.Error("Failed to send message", zap.Error(err))
		}
	}
}
