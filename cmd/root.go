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
var regionParam string

var availableRegionNames = []string{
	"Castilla-La Mancha", "Extremadura", "Madrid",
}

var availableRegions = map[string]*models.Region{}

func init() {
	getCmd.Flags().StringVarP(&date, "date", "d", "Today", "Sets the date to be run (yyyy-MM-dd)")
	getCmd.Flags().StringVarP(&regionParam, "region", "r", "all", "Sets the region to be run")

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
		for _, region := range availableRegions {
			err := processRegion(context.Background(), region)
			if err != nil {
				log.WithFields(log.Fields{
					"error":  err,
					"region": region,
				}).Error("Error processing Agenda")
				return
			}
		}
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets an agenda",
	Long:  "Performs the scrapping and indexing of an agenda, identified by the region and day",
	Run: func(cmd *cobra.Command, args []string) {
		t := time.Now()
		if date != "Today" {
			layout := "2006-01-02"
			parsedDate, err := time.Parse(layout, date)
			if err != nil {
				log.WithFields(log.Fields{
					"date": date,
				}).Fatal("Wrong date format. Please use yyyy-MM-dd")
			}
			t = parsedDate
		}

		if regionParam != "all" {
			if !contains(availableRegionNames, regionParam) {
				log.WithFields(log.Fields{
					"region":    regionParam,
					"available": availableRegionNames,
				}).Fatal("Region not supported.")
			}

			region, err := regions.RegionFactory(regionParam)
			if err != nil {
				log.WithFields(log.Fields{
					"error":  err,
					"region": regionParam,
				}).Fatal("Cannot initialise region")
			}
			availableRegions[regionParam] = region
		} else {
			for _, regionName := range availableRegionNames {
				region, err := regions.RegionFactory(regionName)
				if err != nil {
					log.WithFields(log.Fields{
						"error":  err,
						"region": regionName,
					}).Fatal("Cannot initialise regions")
				}
				availableRegions[regionName] = region
			}
		}

		for _, region := range availableRegions {
			err := processAgenda(context.Background(), region, t.Day(), int(t.Month()), t.Year())
			if err != nil {
				log.WithFields(log.Fields{
					"date":   date,
					"region": region,
				}).Fatal("Error retrieving the agenda for one day")
			}
		}
	},
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func processAgenda(ctx context.Context, region *models.Region, day int, month int, year int) error {
	log.WithFields(log.Fields{
		"day":    day,
		"month":  month,
		"year":   year,
		"region": region.Name,
	}).Info("Processing agenda")

	agenda, err := regions.AgendaFactory(region, day, month, year)
	if err != nil {
		return err
	}

	agenda.Scrap(context.Background())

	indexer, _ := indexers.GetIndexer("elasticsearch")
	for _, event := range agenda.Events {
		err := indexer.Index(context.Background(), event)
		if err != nil {
			log.WithFields(log.Fields{
				"agendaID": agenda.ID,
				"date":     agenda.Date,
				"error":    err,
			}).Errorf("error indexing event")
			return err
		}
	}

	return nil
}

// processRegion processes all entities for a region, from the beginning to the end
func processRegion(ctx context.Context, region *models.Region) error {
	start := region.StartDate.ToDate()
	end := time.Now()

	for rd := regions.RangeDate(start, end); ; {
		date := rd()
		if date.IsZero() {
			break
		}

		err := processAgenda(context.Background(), region, date.Day(), int(date.Month()), date.Year())
		if err != nil {
			return err
		}
	}

	return nil
}
