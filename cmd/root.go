package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "package-assistant",
	Short: "A REST data service for package-repo",
	Long:  `package-asistant is a RESTful data service for package-repo`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/package-assistant.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		log.Err(err).
			Msg("Couldn't find home directory")
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(home)                // adding home directory as first search path
		viper.AddConfigPath(".")                 // also look in the working directory
		viper.SetConfigName("package-assistant") // name the config file (without extension)
	}

	//	Environment settings
	viper.SetEnvPrefix("packassist") // will be uppercased automatically
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	//	Set our defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.allowed-origins", "*")
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
	viper.SetDefault("upload.path", path.Join(home, "package-assistant", "uploads"))
	viper.SetDefault("upload.bytelimit", 30*1024*1024) // 30MB
	viper.SetDefault("github.projecturl", "https://github.com/some/package-repo")
	viper.SetDefault("github.projectfolder", "/data/package-repo")
	viper.SetDefault("github.user", "someuser")
	viper.SetDefault("github.password", "sometoken")

	// If a config file is found, read it in
	_ = viper.ReadInConfig()

}
