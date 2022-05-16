/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/benaheilman/factor/worker"
	"github.com/spf13/cobra"
)

// factorCmd represents the factor command
var factorCmd = &cobra.Command{
	Use:   "factor",
	Short: "Factor random 64 bit numbers into their primes",
	Long:  `Testing performance of different factoring methods`,
	Run: func(cmd *cobra.Command, args []string) {
		method, err := cmd.Flags().GetString("method")
		if err != nil {
			log.Fatal(err)
		}
		shift, err := cmd.Flags().GetInt("shift")
		if err != nil {
			log.Fatal(err)
		}
		timeout, err := cmd.Flags().GetInt("timeout")
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
		worker.Manage(time.Second*time.Duration(timeout), shift, method)
	},
}

func init() {
	rootCmd.AddCommand(factorCmd)

	factorCmd.Flags().StringP("method", "m", "naive", "Prime factoring method")
	factorCmd.Flags().IntP("shift", "s", 16, "Shift inputs right by shift bits")
	factorCmd.Flags().IntP("timeout", "t", 1, "Minumum number of seconds to run")
}
