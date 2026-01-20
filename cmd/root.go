/*
Package cmd
Copyright © 2025 Chuang Libs <chan.toddd@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "phototidy",
	Short: "一个用于整理照片的小工具",
	Long: `一个用于整理照片的小工具
把照片、视频文件按照日期分类到指定目录，例： phototidy date -d <目录>`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("help", "h", false, "查看 phototidy 的帮助")
	setCustomHelpCommand(rootCmd)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.phototidy.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
func setCustomHelpCommand(rootCmd *cobra.Command) {
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:   "help [command]",
		Short: "查看命令的帮助信息",
		Long:  "为所有命令显示帮助信息",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				rootCmd.Help()
				return
			}

			c, _, err := rootCmd.Find(args)
			if err != nil || c == nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Unknown help topic: %s\n", args[0])
				return
			}
			c.Help()
		},
	})
}
