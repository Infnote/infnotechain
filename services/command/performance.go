package command

import (
	"github.com/spf13/cobra"
	"math/rand"
	"strconv"
)

var createBlock = &cobra.Command{
	Use: "t_cb",
	Run: func(cmd *cobra.Command, args []string) {
		if chainContext == nil {
			print("chain context not set\n")
			return
		}

		b, _ := strconv.Atoi(cmd.Flag("bytes").Value.String())
		k, _ := strconv.Atoi(cmd.Flag("kilobytes").Value.String())
		m, _ := strconv.Atoi(cmd.Flag("megabytes").Value.String())
		size := b + k * 1024 + m * 1024 * 1024

		_ = cmd.Flag("bytes").Value.Set("0")
		_ = cmd.Flag("kilobytes").Value.Set("0")
		_ = cmd.Flag("megabytes").Value.Set("0")

		if size == 0 {
			print("size should be greater than 0\n")
			return
		}

		content := make([]byte, size)
		for i := range content {
			content[i] = byte(rand.Intn(90) + 32)
		}

		CreateBlock(chainContext.ID, content)
	},
}

func initPerformanceCommands() {
	createBlock.Flags().IntP("bytes", "b", 0, "bytes")
	createBlock.Flags().IntP("kilobytes", "k", 0, "kilobytes")
	createBlock.Flags().IntP("megabytes", "m", 0, "megabytes")

	cliRootCmd.AddCommand(createBlock)
}
