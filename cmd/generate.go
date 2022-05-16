/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"
	"os"
	"runtime/pprof"

	"github.com/benaheilman/factor/worker"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a binary file of primes",
	Long:  `Output of this command is used by the disk factoring method`,
	Run: func(cmd *cobra.Command, args []string) {
		path, err := cmd.Flags().GetString("output")
		if err != nil {
			log.Fatal(err)
		}
		limit, err := cmd.Flags().GetInt8("limit")
		if err != nil {
			log.Fatal(err)
		}
		cpu, err := cmd.Parent().Flags().GetString("cpu-profiler")
		if err != nil {
			log.Fatal(err)
		}
		f, err := os.Create(cpu)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		err = pprof.StartCPUProfile(f)
		if err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
		worker.Generate(path, int(limit))
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	generateCmd.Flags().StringP("output", "o", "primes.bin", "Output binary file")
	generateCmd.Flags().Int8P("limit", "l", 16, "Limit construction to limit bits")
}
