package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd/builtin"
	"github.com/df-mc/dragonfly/server/console"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/whitelist"
	"github.com/pelletier/go-toml"
)

func main() {
	writer := console.NewWriter(os.Stdout, console.SupportsColour(os.Stdout))
	logger := slog.New(console.NewLogHandler(writer, slog.LevelDebug))
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelDebug)
	chat.Global.Subscribe(console.NewChatSubscriber(writer))
	builtin.RegisterPerformance()

	wl, err := whitelist.New("whitelist.json")
	if err != nil {
		panic(err)
	}

	conf, err := readConfig(logger)
	if err != nil {
		panic(err)
	}

	conf.Allower = wl

	srv := conf.New()
	builtin.RegisterServer(srv)
	builtin.RegisterWorld(srv)
	builtin.RegisterWhitelist(srv, wl)
	srv.CloseOnProgramEnd()

	srv.Listen()
	console.Start(os.Stdin, writer, console.NewSource(writer, srv.World().Spawn().Vec3()), srv.World())
	for p := range srv.Accept() {
		_ = p
	}
}

// readConfig reads the configuration from the config.toml file, or creates the
// file if it does not yet exist.
func readConfig(log *slog.Logger) (server.Config, error) {
	c := server.DefaultConfig()
	var zero server.Config
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		data, err := toml.Marshal(c)
		if err != nil {
			return zero, fmt.Errorf("encode default config: %v", err)
		}
		if err := os.WriteFile("config.toml", data, 0644); err != nil {
			return zero, fmt.Errorf("create default config: %v", err)
		}
		return c.Config(log)
	}
	data, err := os.ReadFile("config.toml")
	if err != nil {
		return zero, fmt.Errorf("read config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		return zero, fmt.Errorf("decode config: %v", err)
	}
	return c.Config(log)
}
