package reporter

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	// TrafficConditionHeavy is our value for HEAVY traffic conditions
	TrafficConditionHeavy TrafficCondition = "HEAVY"
	// TrafficConditionModerate is our value for MODERATE traffic conditions
	TrafficConditionModerate TrafficCondition = "MODERATE"
	// TrafficConditionLight is our value for LIGHT traffic conditions
	TrafficConditionLight TrafficCondition = "LIGHT"

	dateFormat = "2006-01-02 15:04:05"
)

// NewTrafficReport creates a traffic report for the robotID specified
func NewTrafficReport(
	id int,
	ts time.Time,
	speed float64,
	traffic TrafficCondition,
) *TrafficReport {
	return &TrafficReport{
		RobotID: id,
		Time:    ts,
		Speed:   speed,
		Traffic: traffic,
	}
}

// TrafficReport represents the data for a traffic report
type TrafficReport struct {
	RobotID int
	Time    time.Time
	Speed   float64
	Traffic TrafficCondition
}

// TrafficCondition provides a base type for TrafficConditions reported by the
// program
type TrafficCondition string

// saveTrafficReport writes all the TrafficReport records to a CSV file
func SaveTrafficReport(reports chan *TrafficReport, reportPath string) error {
	writer, err := os.OpenFile(reportPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Unable to create traffic-report.csv: %v", err)
		return err
	}

	csvWriter := csv.NewWriter(writer)
	for {
		select {
		case r, ok := <-reports:
			if !ok {
				break
			}
			record := []string{
				strconv.Itoa(r.RobotID),
				r.Time.Format(dateFormat),
				strconv.FormatFloat(r.Speed, 'f', 2, 64),
				string(r.Traffic),
			}

			if err := csvWriter.Write(record); err != nil {
				log.Fatalf("Error writing data to traffic-report.csv: %v", err)
			}
			csvWriter.Flush()
		}
	}

	defer writer.Close()
	return nil
}
