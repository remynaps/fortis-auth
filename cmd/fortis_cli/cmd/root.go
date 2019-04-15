package cmd

import (
	"fmt"
	"os"

	"gitlab.com/gilden/fortis/logging"
	"gitlab.com/gilden/fortis/models"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var store *models.DB

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fortis",
	Short: "Command line interface for fortis",
	Long: `A Command line tool to manage fortis. This tool should only be used by admins as it contains
	some pretty useful functions.`,
	PersistentPreRun: dbConnect,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.fortis_cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".fortis_cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".fortis_cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func dbConnect(cmd *cobra.Command, args []string) {
	config, err := readConfig("config.dev")
	if err != nil {
		logging.Panic(err)
	}
	db, err := models.InitDB(config)
	if err != nil {
		// db connection failed. start the retry logic
		logging.Error("Failed to connect to the database." + err.Error())
	}

	store = db
}

func readConfig(filename string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigName(filename)
	v.AddConfigPath("./config")
	err := v.ReadInConfig()

	return v, err
}
