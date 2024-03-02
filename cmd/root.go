package cmd

import (
	"client/global"

	"github.com/spf13/cobra"
)

var Config string

var rootCmd = &cobra.Command{
	Use:   "", // 子命令的标识
	Short: "", // 简短帮助说明
	Long:  "", // 详细帮助说明
	Run: func(cmd *cobra.Command, args []string) {
		// 主程序，获取自定义配置文件
		global.ConfigPath = Config
	},
}

// 供主程序调用
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 参数分别表示，绑定的变量，参数长名(--str)，参数短名(-s)，默认内容，帮助信息
	rootCmd.Flags().StringVarP(&Config, "config", "c", "config/ingress.yaml", "Path to config file")
}
