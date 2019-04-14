// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// addclientCmd represents the addclient command
var addclientCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a new oauth client",
	Long: `Use this command to add a new oauth client. 
	The secret will be shown once. Be sure to store the secret securely.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("addclient called")
	},
}

func init() {
	clientCmd.AddCommand(addclientCmd)

	clientCmd.Flags().StringP("name", "n", "", "Set your name")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addclientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addclientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
