package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/imjoshholloway/hailobots/dispatcher"
	"github.com/imjoshholloway/hailobots/reporter"
	"github.com/imjoshholloway/hailobots/robots"
)

// populateStationsList takes a io.Reader instance and
// creates each of the stations in the Stations slice
func populateStationsList(stationsCSV io.Reader) (map[string]robots.Point, error) {

	stations := make(map[string]robots.Point)
	reader := csv.NewReader(stationsCSV)
	reader.FieldsPerRecord = 3

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Unable to process line in tube.csv: %v", err)
			return nil, err
		}

		lat, err := strconv.ParseFloat(line[1], 64)
		if err != nil {
			log.Fatalf("Unable to parse lat from tube.csv: %v", err)
			return nil, err
		}
		lon, err := strconv.ParseFloat(line[2], 64)
		if err != nil {
			log.Fatalf("Unable to parse lon from tube.csv: %v", err)
			return nil, err
		}

		fmt.Printf("Loaded station: %v \t\t Lat/Lon: %v, %v\n", line[0], lat, lon)
		stations[line[0]] = robots.Point{
			Lat: lat,
			Lon: lon,
		}
	}

	return stations, nil
}

func main() {

	// TODO: Move to flags
	robotIDs := [2]int{5937, 6043}
	stationsCSVPath := "data/tube.csv"
	robotPointsPath := "data"
	reportFilePath := "traffic-report.csv"

	bots := make(map[int]*robots.Robot)
	sources := make([]io.Reader, 0, len(robotIDs))

	// Open our stations CSV file and populate the list of stations with
	// it.
	stationsCSV, err := os.Open(stationsCSVPath)
	if err != nil {
		log.Fatalf("Unable to load tube.csv: %v", err)
	}

	stations, err := populateStationsList(stationsCSV)
	defer stationsCSV.Close()

	trafficReports := make(chan *reporter.TrafficReport)

	// Loop over each of the robotID's we have, load the CSV files
	// and create the Robot instances
	for _, id := range robotIDs {
		csvReader, err := os.Open(fmt.Sprintf("%s/%d.csv", robotPointsPath, id))
		if err != nil {
			log.Printf("No CSV file found for Robot: %d", id)
			continue
		}
		sources = append(sources, csvReader)
		bots[id] = robots.New(id, stations, trafficReports)
	}

	go reporter.SaveTrafficReport(trafficReports, reportFilePath)

	d := dispatcher.New(bots, sources)
	go d.Process()

	fmt.Printf("Dispatcher Terminated: %v\n", <-d.Terminate)
}
