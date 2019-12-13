package main

import (
	"strings"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// main is the main entry point of the app.
func main() {
	// Print banner
	color.NoColor = false

	color.HiCyan("                                          __ __      __    ")
	color.HiCyan(".-.--..---.-.--.-.--.-----.-----._____.--|  |__|----|  |_  ")
	color.HiCyan("|  .  |  -  |  . .  |  -__|__ --|_____|  -  |  |  --|   _| ")
	color.HiCyan("|__|__|___._|__|-|__|_____|_____|     |_____|__|____|_____|")
	color.HiCyan("                                                           ")

	// Cobra command
	cmd := &cobra.Command{
		Use:     "names-dict",
		Long:    "Create a password dictionary based on names.",
		Args:    cobra.NoArgs,
		Version: "0.0.1",
		Run:     namesDict,
	}

	cmd.Flags().BoolP("verbose", "v", false, "write more")

	//	cmd.Flags().StringP("listen", "l", "0.0.0.0:80", "IP and port on which the server will listen")
	//	cmd.Flags().StringP("assets", "a", "", "Path to static web assets")
	//
	//	cmd.Flags().StringP("db-host", "H", "localhost", "MySQL host")
	//	cmd.Flags().StringP("db-database", "d", "postfix", "MySQL database")
	//	cmd.Flags().StringP("db-username", "u", "postfix", "MySQL username")
	//	cmd.Flags().StringP("db-password", "p", "", "MySQL password")

	// Viper config
	viper.SetEnvPrefix("NAMES_DICT")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.BindPFlags(cmd.Flags())

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/names-dict")
	viper.AddConfigPath("$HOME/.config/names-dict")
	viper.AddConfigPath(".")

	viper.ReadInConfig()

	// Run command
	cmd.Execute()
}

// aykroyd is called if the CLI interfaces has been satisfied.
func namesDict(cmd *cobra.Command, args []string) {
	// Set logging level
	if viper.GetBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}
