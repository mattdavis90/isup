package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gopkg.in/natefinch/lumberjack.v2"

	"isup/config"
	"isup/scheduler"
)

func configureLogging(c *cli.Context) error {
	var writers []io.Writer

	if c.Bool("logging.console") {
		console := os.Stderr
		if c.Bool("logging.console.stdout") {
			console = os.Stdout
		}
		if c.Bool("logging.console.json") {
			writers = append(writers, console)
		} else {
			writers = append(writers, zerolog.ConsoleWriter{
				Out:        console,
				NoColor:    !c.Bool("logging.console.color"),
				TimeFormat: time.RFC3339,
			})
		}
	}

	logFilename := c.String("logging.filename")
	if logFilename != "" {
		logDir := filepath.Dir(logFilename)
		err := os.MkdirAll(logDir, 0744)
		if err != nil {
			return err
		} else {
			writers = append(writers, &lumberjack.Logger{
				Filename:   logFilename,
				MaxBackups: c.Int("logging.backups"),
				MaxSize:    c.Int("logging.size"),
				MaxAge:     c.Int("logging.age"),
			})
		}
	}

	level, err := zerolog.ParseLevel(c.String("logging.level"))
	if err != nil {
		return err
	}

	multiwriter := io.MultiWriter(writers...)
	log.Logger = zerolog.New(multiwriter).Level(level).With().Timestamp().Logger()

	return nil
}

func run(c *cli.Context) error {
	err := configureLogging(c)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not configure logging")
	}

	configfile := c.String("config")
	log.Info().Str("filename", configfile).Msg("Loading config")
	cfg, err := config.LoadConfig(configfile)
	if err != nil {
		log.Fatal().Err(err).Msg("Config error")
	}

	log.Info().Msg("Initialisation complete")

	sigExit := make(chan os.Signal, 1)
	sigReload := make(chan os.Signal, 1)
	signal.Notify(sigExit, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(sigReload, syscall.SIGUSR1)

	var cancel context.CancelFunc
	cancel = startScheduler(cfg)

	for {
		select {
		case <-sigExit:
			cancel()
			log.Info().Msg("Exiting")
			return nil
		case <-sigReload:
			log.Info().Msg("Reloading config")
			newCfg, err := cfg.Reload()
			if err != nil {
				log.Error().Err(err).Msg("Could not reload config")
			} else {
				cfg = newCfg
				log.Info().Msg("Config reloaded successfully")
				cancel()
				cancel = startScheduler(cfg)
			}
		}
	}
}

func startScheduler(cfg *config.Config) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "http.client", cfg.Client.HttpClient())

	r := scheduler.NewRouter()
	r.Alerters = cfg.Alerters

	go cfg.Schedule.Run(ctx, r.Alerts)
	go r.Run(ctx)

	return cancel
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			// Config File Options
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				EnvVars:  []string{"ISUP_CONFIG"},
				Required: true,
				Usage:    "Load config from `FILE` (required)",
			},
			//Logging Options
			&cli.StringFlag{
				Name:    "logging.level",
				EnvVars: []string{"ISUP_LOGGING_LEVEL"},
				Value:   "info",
				Usage:   "Set the log level",
			},
			&cli.BoolFlag{
				Name:    "logging.console",
				EnvVars: []string{"ISUP_LOGGING_CONSOLE"},
				Value:   true,
				Usage:   "Enable console logging",
			},
			&cli.BoolFlag{
				Name:    "logging.console.stdout",
				EnvVars: []string{"ISUP_LOGGING_CONSOLE_STDOUT"},
				Value:   false,
				Usage:   "Log to stdout instead of stderr",
			},
			&cli.BoolFlag{
				Name:    "logging.console.json",
				EnvVars: []string{"ISUP_LOGGING_CONSOLE_JSON"},
				Value:   false,
				Usage:   "Enable json logging to console",
			},
			&cli.BoolFlag{
				Name:    "logging.console.color",
				EnvVars: []string{"ISUP_LOGGING_CONSOLE_COLOR"},
				Value:   true,
				Usage:   "Enable color logging to console",
			},
			&cli.StringFlag{
				Name:    "logging.filename",
				EnvVars: []string{"ISUP_LOGGING_FILENAME"},
				Usage:   "Log in JSON format to `FILENAME`",
			},
			&cli.IntFlag{
				Name:    "logging.backups",
				EnvVars: []string{"ISUP_LOGGING_BACKUPS"},
				Value:   3,
				Usage:   "Store `BACKUPS` hisotric logs",
			},
			&cli.IntFlag{
				Name:    "logging.size",
				EnvVars: []string{"ISUP_LOGGING_SIZE"},
				Value:   5,
				Usage:   "Max `SIZE` of logs in MB",
			},
			&cli.IntFlag{
				Name:    "logging.age",
				EnvVars: []string{"ISUP_LOGGING_AGE"},
				Value:   7,
				Usage:   "Max `AGE` of logs in days",
			},
		},
		Action: run,
	}
	app.Run(os.Args)
}
