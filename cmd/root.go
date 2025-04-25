package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/mathismqn/godeez/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   "godeez",
	Short: "GoDeez is a tool to download music from Deezer",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.godeez)")
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		homedir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not get home directory: %v\n", err)
			os.Exit(1)
		}

		path := path.Join(homedir, ".godeez")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("Config file not found, creating one at %s\n", path)

			content := []byte("arl_cookie = ''\nsecret_key = ''\niv = '0001020304050607'\n")
			if err := os.WriteFile(path, content, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error: could not create config file: %v\n", err)
				os.Exit(1)
			}
		}

		viper.AddConfigPath(homedir)
		viper.SetConfigName(".godeez")
	}

	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not read config file: %v\n", err)
		os.Exit(1)
	}

	cfg := &config.Cfg
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not unmarshal config file: %v\n", err)
		os.Exit(1)
	}

	if cfg.SecretKey == "" {
		fmt.Fprintln(os.Stderr, "Error: secret_key is not set in config file")
		os.Exit(1)
	}
	if cfg.IV == "" {
		fmt.Fprintln(os.Stderr, "Error: iv is not set in config file")
		os.Exit(1)
	}
}
