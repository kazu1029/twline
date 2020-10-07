/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"strings"

	"github.com/kazu1029/twline/get"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	listKey   = "list"
	outputKey = "output"
)

var outputSource string

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get command will get timelines from you specified",
	Long: `get command will get timelimes from you specified urls.
e.g.
twline get "search?q=trip&src=typed_query"`,
	Run: func(cmd *cobra.Command, args []string) {
		isList := viper.GetBool(listKey)

		var urls []string
		if isList {
			urls = strings.Split(args[0], ",")
		} else {
			urls = []string{args[0]}
		}

		get.GetTimeline(urls, outputSource)
	},
	Args: cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().BoolP(listKey, "l", false, "list for multiple target urls.")
	viper.BindPFlag(listKey, getCmd.Flags().Lookup(listKey))
	getCmd.Flags().StringVarP(&outputSource, outputKey, "o", "", "Output Destrination.")
}
