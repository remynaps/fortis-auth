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
	"crypto/rand"
	"encoding/base64"
	"fmt"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"gitlab.com/gilden/fortis/models"
	"golang.org/x/crypto/bcrypt"
)

var name, redirect string // used for flags

// addclientCmd represents the addclient command
var addclientCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a new oauth client",
	Long: `Use this command to add a new oauth client. 
	The secret will be shown once. Be sure to store the secret securely.`,
	Run: func(cmd *cobra.Command, args []string) {

		name, _ := cmd.Flags().GetString("name")
		redirect, _ := cmd.Flags().GetString("redirect")

		clientID := uuid.NewV4().String()

		length := 55 // 55 chars
		array := make([]byte, length)
		if _, err := rand.Read(array); err != nil {
			panic(err)
		}

		hashedSecret, err := bcrypt.GenerateFromPassword([]byte(array), bcrypt.DefaultCost)

		if err != nil {
			panic(err)
		}

		client := models.AuthClient{
			ID:           clientID,
			DisplayName:  name, // retrieve value from viper
			ClientSecret: fmt.Sprintf("%X", hashedSecret),
			RedirectUris: []string{redirect},
			Scopes:       []string{"All"},
			Private:      true,
		}
		err = store.InsertClient(&client)

		if err != nil {
			fmt.Println("Failed to create client: " + err.Error())
		} else {
			fmt.Println("Created client: " + client.DisplayName)
			fmt.Println("Client ID: " + client.ID)
			fmt.Println("Client secret : " + base64.URLEncoding.EncodeToString(array))
			fmt.Println("Store the secret securerly. You will have to generate a new one you lose the secret!")
		}
	},
}

func init() {
	clientCmd.AddCommand(addclientCmd)

	addclientCmd.Flags().StringP("name", "n", "", "Set the client name")
	addclientCmd.Flags().StringP("redirect", "r", "", "Set the redirect url")
	addclientCmd.Flags().BoolP("private", "p", true, "Set if the client is private")

	addclientCmd.MarkFlagRequired("name")
	addclientCmd.MarkFlagRequired("redirect")
	addclientCmd.MarkFlagRequired("private")
}
