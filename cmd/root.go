package cmd

import (
	"context"
	"time"

	"github.com/mdelapenya/cansino/indexers"
	"github.com/mdelapenya/cansino/models"
	"github.com/mdelapenya/cansino/regions"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var date string

func init() {
	getCmd.Flags().StringVarP(&date, "date", "d", "Today", "Sets the date to be run (yyyy-MM-dd)")

	rootCmd.AddCommand(chaseCmd)
	rootCmd.AddCommand(getCmd)
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
		err := processRegion(context.Background())
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Error processing Agenda CLM")
			return
		}
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets an agenda",
	Long:  "Performs the scrapping and indexing of an agenda, identified by the region and day",
	Run: func(cmd *cobra.Command, args []string) {
		layout := "2006-01-02"
		t, err := time.Parse(layout, date)
		if err != nil {
			log.WithFields(log.Fields{
				"date": date,
			}).Fatal("Wrong date format. Please use yyyy-MM-dd")
		}

		clm := regions.NewAgendaCLM(t.Day(), int(t.Month()), t.Year())

		processAgenda(context.Background(), clm)
	},
}

func processAgenda(ctx context.Context, a *models.Agenda) error {
	a.Scrap(context.Background())

	indexer, _ := indexers.GetIndexer("elasticsearch")
	for _, event := range a.Events {
		err := indexer.Index(context.Background(), event)
		if err != nil {
			log.WithFields(log.Fields{
				"agendaID": a.ID,
				"date":     a.Date,
				"error":    err,
			}).Errorf("error indexing event")
			return err
		}
	}

	return nil
}

// processRegion processes all entities from the beginning to the end
func processRegion(ctx context.Context) error {
	start := regions.HistoricalStartDate.ToDate()
	end := time.Now()

	for rd := regions.RangeDate(start, end); ; {
		date := rd()
		if date.IsZero() {
			break
		}

		clm := regions.NewAgendaCLM(date.Day(), int(date.Month()), date.Year())

		processAgenda(context.Background(), clm)
	}

	return nil
}
