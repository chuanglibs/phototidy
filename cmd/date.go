/*
Package cmd
Copyright © 2025 Chuang Libs <chan.toddd@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// dateCmd represents the date command
var dateCmd = &cobra.Command{
	Use:   "date",
	Short: "把照片、视频按日期来进行目录分类",
	Long: `读取照片、视频文件的日期来进行识别
以其拍摄日期为准来进行目录分类，按年份-》月份来进行目录创建，并把图片、视频文件移动到对应的目录中
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("date called")
	},
}

func init() {
	rootCmd.AddCommand(dateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
