/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/philippgille/gokv/Client"
	"github.com/philippgille/gokv/redis"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get Value from key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		cfg := viper.GetString("store")
		if cfg != "redis" {
			fmt.Println("unrecognised store")
			os.Exit(1)
		}
		options := redis.Options{
			Address:  viper.GetString("Address"),
			Password: viper.GetString("Password"),
			DB:       viper.GetInt("DB"),
		}

		client := Client.NewStorage()
		store, err := client.Redis.GetClient(options)
		if err != nil {
			fmt.Println("Error")
		}
		retrievedVal := ""
		found, err := store.Get(key, &retrievedVal)
		if err != nil {
			fmt.Printf("Error: %v", err)
		}
		if !found {
			fmt.Println("Key-Value not found")
		}
		fmt.Printf("Retrieved Value: %v\n", retrievedVal)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
