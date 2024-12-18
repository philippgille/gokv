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

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set value in store",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]
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
			fmt.Println("Error", err)
			os.Exit(1)
		}
		err = store.Set(key, value)
		if err != nil {
			fmt.Println("Error", err)
			os.Exit(1)
		}
		fmt.Println("Value added to store")
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
