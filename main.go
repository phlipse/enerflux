package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
)

// command line flags
var (
	energyUsername    string
	energyPassword    string
	energyCustomer    string
	energyAPI         string
	workDir           string
	interval          int
	influxAddress     string
	influxUsername    string
	influxPassword    string
	influxDatabase    string
	influxMeasurement string
	influxTagMeter    string
	useNewState       bool
	printCount        bool
)

// constants used during execution
const (
	// min interval duration in seconds
	intervalMin = 5
	// name of state file
	stateFile = "/state.json"
)

func init() {
	// get command line flags
	flag.StringVar(&energyUsername, "u", "admin", "Username of getfresh.energy API.")
	flag.StringVar(&energyPassword, "p", "secret", "Password of getfresh.energy API.")
	flag.StringVar(&energyCustomer, "c", "00000", "Customer of getfresh.energy API.")
	flag.StringVar(&energyAPI, "a", "https://www.getfresh.energy", "URL of getfresh.energy API.")
	flag.StringVar(&workDir, "w", ".", "Path to working directory, has to be writable by executing user.")
	flag.IntVar(&interval, "i", 30, "Interval to query getfresh.energy API.")
	flag.StringVar(&influxAddress, "t", "http://127.0.0.1:8086", "URL of inflxuDB.")
	flag.StringVar(&influxUsername, "j", "admin", "Username of inflxuDB.")
	flag.StringVar(&influxPassword, "k", "secret", "Password of inflxuDB.")
	flag.StringVar(&influxDatabase, "d", "freshenergy", "Target database in inflxuDB.")
	flag.StringVar(&influxMeasurement, "m", "energy", "Measurement name in inflxuDB.")
	flag.StringVar(&influxTagMeter, "l", "", "Optional meter location for metric tag.")
	flag.BoolVar(&useNewState, "s", false, "Use new state even if there is a persistent one.")
	flag.BoolVar(&printCount, "z", false, "Print out count of data points that will be written to influxdb.")
	flag.Parse()

	// adjust interval if set too low
	if interval < intervalMin {
		fmt.Printf("[WARN] interval set too low, setting it to %d\n", intervalMin)
		interval = intervalMin
	}
}

func main() {
	// create a new influxdb client
	influxClient, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     influxAddress,
		Username: influxUsername,
		Password: influxPassword,
	})
	if err != nil {
		fmt.Printf("[ERROR] could not create influxdb client: %v\n", err)
		os.Exit(1)
	}
	// try to ping influxdb, establish first connection
	_, _, err = influxClient.Ping(0)
	if err != nil {
		fmt.Printf("[ERROR] could not establish connection to influxdb: %v\n", err)
		os.Exit(1)
	}
	// set metric meter tag, location of smart meter
	tags := map[string]string{"meter": influxTagMeter}

	// get new fresh energy client for API access
	energyClient := NewFreshEnergyClient(energyUsername, energyPassword, energyCustomer)
	// persist state at shutdown
	defer func() {
		if err := energyClient.PersistState(); err != nil {
			fmt.Printf("[WARN] could not persist state: %v\n", err)
		}
	}()

	// make channel for SIGINT, gracefull shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	// make ticker
	tick := time.NewTicker(time.Second * time.Duration(interval))

	// do stuff
	for {
		select {
		case <-sig:
			// shutdown gracefully
			tick.Stop()
			return
		case <-tick.C:
			// get energy content from API
			err := energyClient.Get()
			if err != nil {
				fmt.Printf("[WARN] could not retrieve energy content from api: %v\n", err)
				continue
			}

			// create new metric batch
			batch, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
				Database:  influxDatabase,
				Precision: "s",
			})
			if err != nil {
				fmt.Printf("[WARN] could not create new metric batch: %v\n", err)
				continue
			}

			// create a point for each reading from API and add to batch
			for _, read := range energyClient.API.Readings {
				fields := map[string]interface{}{
					"power":         read.Power,
					"powerPhase1":   read.PowerPhase1,
					"powerPhase2":   read.PowerPhase2,
					"powerPhase3":   read.PowerPhase3,
					"energyReading": read.EnergyReading,
				}

				point, err := influxdb.NewPoint(influxMeasurement, tags, fields, read.DateTime.In(time.Local))
				if err != nil {
					fmt.Printf("[WARN] could not create data point: %v\n", err)
					continue
				}

				batch.AddPoint(point)
			}

			// print out number of points in batch that will be written to influxdb
			if printCount {
				fmt.Printf("[INFO] write %d data points to influxdb\n", len(batch.Points()))
			}

			// write batch to influxdb
			if err := influxClient.Write(batch); err != nil {
				fmt.Printf("[WARN] could not write batch to influxdb: %v\n", err)
				continue
			}
		}
	}
}
