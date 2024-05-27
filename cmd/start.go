package cmd

import (
	"context"
	"fmt"
	"github.com/danesparza/package-assistant/api"
	_ "github.com/danesparza/package-assistant/docs" // swagger docs location
	"github.com/danesparza/package-assistant/internal/debian"
	"github.com/danesparza/package-assistant/internal/repo"
	"github.com/danesparza/package-assistant/internal/telemetry"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long:  `The serve command starts hosting the service`,
	Run:   start,
}

func start(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	//	If we have a config file, report it:
	if viper.ConfigFileUsed() != "" {
		log.Debug().
			Str("config", viper.ConfigFileUsed()).
			Msg("Using config file")
	} else {
		log.Debug().Msg("No config file found")
	}

	loglevel := viper.GetString("logger.level")

	//	Trap program exit appropriately
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go handleSignals(ctx, sigs, cancel)

	//	Emit what we know:
	log.Info().
		Str("loglevel", loglevel).
		Str("upload.path", viper.GetString("upload.path")).
		Str("upload.bytelimit", viper.GetString("upload.bytelimit")).
		Str("github.projecturl", viper.GetString("github.projecturl")).
		Str("github.projectfolder", viper.GetString("github.projectfolder")).
		Str("github.user", viper.GetString("github.user")).
		Str("github.password", "********").
		Str("git.name", viper.GetString("git.name")).
		Str("git.email", viper.GetString("git.email")).
		Msg("Starting up")

	// Service initialization
	gitRepo, err := repo.InitPackageRepo(ctx,
		viper.GetString("github.projecturl"),
		viper.GetString("github.projectfolder"),
		viper.GetString("github.user"),
		viper.GetString("github.password"),
	)
	if err != nil {
		log.Err(err).Msg("problem initializing git repo")
		return
	}

	err = debian.InitGPGKey(ctx,
		viper.GetString("gpg.key"),
		viper.GetString("gpg.password"),
	)
	if err != nil {
		log.Err(err).Msg("problem initializing gpg key")
		return
	}

	//	Create an api service object
	apiService := api.Service{
		StartTime: time.Now(),
		RepoSvc: repo.NewGitRepoService(
			viper.GetString("github.projecturl"),
			viper.GetString("github.projectfolder"),
			gitRepo),
	}

	//	Create a router and set up our REST endpoints...
	r := chi.NewRouter()

	//	Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(telemetry.Middleware(telemetry.NRApp))
	r.Use(httprate.LimitByIP(1, 1*time.Second)) // Rate limit to 1 call per second (per IP)
	r.Use(api.ApiVersionMiddleware)
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	//	Routes
	r.Route("/v1", func(r chi.Router) {
		r.Post("/package", apiService.UploadPackage)
	})

	//	SWAGGER
	r.Mount("/swagger", httpSwagger.WrapHandler)

	formattedServerPort := fmt.Sprintf(":%v", viper.GetString("server.port"))

	//	HTTP server
	go func() {
		log.Info().Str("server", formattedServerPort).Msg("Started REST service")
		log.Err(http.ListenAndServe(formattedServerPort, r)).Msg("HTTP API service error")
	}()

	//	Wait for our signal and shutdown gracefully
	<-ctx.Done()
}

func handleSignals(ctx context.Context, sigs <-chan os.Signal, cancel context.CancelFunc) {
	select {
	case <-ctx.Done():
	case sig := <-sigs:
		switch sig {
		case os.Interrupt:
			log.Info().Str("signal", "SIGINT").Msg("Shutting down")
		case syscall.SIGTERM:
			log.Info().Str("signal", "SIGTERM").Msg("Shutting down")
		}

		cancel()
		os.Exit(0)
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
