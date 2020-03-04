package cmd

import (
	"context"

	regions "github.com/mdelapenya/cansino/regions"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(chaseCmd)
}

var rootCmd = &cobra.Command{
	Use:   "cansino",
	Short: "Cansino will scrap politicians' public agendas.",
	Long:  `A Fast and Flexible CLI for scrapping politicians' public agendas ❤️ by mdelapenya and friends in Go.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

// Execute execute root command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Error executing command")
	}
}

var chaseCmd = &cobra.Command{
	Use:   "chase",
	Short: "Gets all agendas",
	Long:  "Performs the scrapping and indexing of all agendas",
	Run: func(cmd *cobra.Command, args []string) {
		err := regions.ProcessCLM(context.Background())
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Error processing Agenda CLM")
			return
		}
	},
}
