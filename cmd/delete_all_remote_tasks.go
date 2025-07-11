/*
Copyright © 2025 Linus Karsai <linus@linusk.com.au>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"log/slog"

	"github.com/karsai5/tw-caldav/internal/caldav"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deleteAllRemoteTasks represents the test command
var deleteAllRemoteTasks = &cobra.Command{
	Use:   "delete-all-remote-tasks",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := caldav.NewClient(viper.GetString("url"), viper.GetString("user"), viper.GetString("pass"))
		if err != nil {
			panic(err)
		}

		todos, err := client.GetAllTodos()
		if err != nil {
			panic(err)
		}

		for _, t := range todos {
			err := t.Delete()
			if err != nil {
				slog.Error("Error deleting task", "err", err)
			} else {
				slog.Info("Task deleted", "task", t.Description(), "path", *t.RemotePath())
			}
		}

		// err = remote.CreateCalendar("/hello6", "Test")
		// if err != nil {
		// 	panic(err)
		// }
		// // remote.Client.PutCalendarObject()
		// fmt.Println("test called")
	},
}

func init() {
	rootCmd.AddCommand(deleteAllRemoteTasks)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteAllRemoteTasks.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteAllRemoteTasks.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
