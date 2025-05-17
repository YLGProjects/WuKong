/**
 *MIT License
 *
 *Copyright (c) 2025 ylgeeker
 *
 *Permission is hereby granted, free of charge, to any person obtaining a copy
 *of this software and associated documentation files (the "Software"), to deal
 *in the Software without restriction, including without limitation the rights
 *to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *copies of the Software, and to permit persons to whom the Software is
 *furnished to do so, subject to the following conditions:
 *
 *copies or substantial portions of the Software.
 *
 *THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *SOFTWARE.
**/

package main

import (
	"YLGProjects/WuKong/internal/controller"
	"YLGProjects/WuKong/pkg/logger"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:          "Controller",
		Short:        "WuKong Controller Server",
		SilenceUsage: true,
		RunE:         controller.Run,
	}

	rootCmd.PersistentFlags().StringVarP(&controller.ConfigFilePath, "config", "c", "/var/ylg/wukong/etc/controller.yaml", "")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(controller.VersionCmd)
	if err := rootCmd.Execute(); err != nil {
		logger.Error("failed to start controller server. errmsg:%s", err.Error())
		return
	}

}
