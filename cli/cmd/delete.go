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

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Args:  cobra.ExactArgs(1),
	Short: "Delete a Key Value Pair",
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
		err = store.Delete(key)
		if err != nil {
			fmt.Printf("Error: %v", err)
		}
		fmt.Printf("Key %s has been deleted.", key)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
