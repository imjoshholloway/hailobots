package robots

import (
	"fmt"
	"math"
	"time"

	"github.com/imjoshholloway/hailobots/reporter"
)

const (
	nearbyStationProximity = 0.35
	earthRadius            = 6371.0 // km
)

// New creates a new Robot and starts processing the Points
// passed into it
func New(id int, stations map[string]Point, listener chan *reporter.TrafficReport) *Robot {
	r := &Robot{
		ID:       id,
		Next:     make(chan *RoutePoint, 10),
		Out:      listener,
		Stations: stations,
		Shutdown: make(chan bool),
	}

	// Process our RoutePoints as they come in and generate the TrafficReport
	// if we need to and return the Report to the Channel
	go r.Run()
	return r
}

// Robot represents our worker, it moves from Point to Point which is provided by
// the Dispatcher and creates TrafficReports where needed
type Robot struct {
	ID       int
	Next     chan *RoutePoint
	Current  *RoutePoint
	Last     *RoutePoint
	Stations map[string]Point
	Out      chan *reporter.TrafficReport
	Shutdown chan bool
}

// Point represents the lat&lon on a Route
type Point struct {
	Lat float64
	Lon float64
}

// RoutePoint represents the point on a route, it includes the datetime of when
// the Robot was at the point & the next RoutePoint
type RoutePoint struct {
	Time time.Time
	Point
}

// Run starts the the Robot processing the Points passed to it by the
// Dispatcher
func (r *Robot) Run() {
	for {
		select {
		case <-r.Shutdown:
			fmt.Printf("Robot: %d received shutdown signal\n", r.ID)
			break
		case current, ok := <-r.Next:
			if current == nil && !ok {
				r.Shutdown <- true
				return
			}
			r.Current = current

			if r.Last != nil && r.Current.Point == r.Last.Point {
				fmt.Printf("Robot: %d location is the same, not moving\n", r.ID)
				break
			}

			if r.Last == nil {
				fmt.Printf("Robot: %d starting at point (lat/lon): %v/%v\n", r.ID, r.Current.Point.Lat, r.Current.Point.Lon)
			} else {
				fmt.Printf("Robot: %d moving from point (lat/lon): %v/%v to point (lat/lon): %v/%v. Time: %v \n", r.ID, r.Last.Point.Lat, r.Last.Point.Lon, r.Current.Point.Lat, r.Current.Point.Lon, r.Current.Time)
			}

			// Generate the TrafficReport and publish it to the TrafficReportsChannel
			if report := r.GenerateTrafficReport(); report != nil {
				r.Out <- report
			}
			r.Last = current
		}
	}
}

// GenerateTrafficReport finds any nearbyStations, and generate a TrafficReport
// if any are found
func (r Robot) GenerateTrafficReport() *reporter.TrafficReport {
	nearby := r.findNearbyStations(r.Current)

	if len(nearby) == 0 {
		return nil
	}

	for station, distance := range nearby {
		fmt.Printf("Robot: %v - %v Station is %vkm away from point (lat/lon): %v/%v\n", r.ID, station, distance, r.Current.Point.Lat, r.Current.Point.Lon)
	}

	var robotSpeed float64
	var distance float64
	if r.Last != nil {
		timeTaken := float64(r.Current.Time.Unix() - r.Last.Time.Unix())
		distance = distanceInKm(r.Last.Point, r.Current.Point)
		robotSpeed = speed(distance, timeTaken)
	}

	return reporter.NewTrafficReport(
		r.ID,
		r.Current.Time,
		robotSpeed,
		getTrafficCondition(robotSpeed, distance),
	)
}

// findNearbyStations takes a RoutePoint and returns a list of the Stations
// found within the nearbyStationProximity
func (r *Robot) findNearbyStations(point *RoutePoint) map[string]float64 {
	nearby := make(map[string]float64)
	for name, station := range r.Stations {
		distance := distanceInKm(point.Point, station)
		if distance < nearbyStationProximity {
			nearby[name] = distance
		}
	}

	return nearby
}

// getTrafficCondition is a helper func to get the TrafficCondition based
// on the speed / distance travelled. Completely random but designed to try and
// give a good spread of HEAVY/MODERATE/LIGHT results
func getTrafficCondition(rSpeed float64, distance float64) reporter.TrafficCondition {
	if (distance < 2 && rSpeed < 24.14) || (distance > 4 && rSpeed < 40.23) {
		return reporter.TrafficConditionHeavy
	} else if distance < 1 && rSpeed < 32.19 {
		return reporter.TrafficConditionModerate
	}

	return reporter.TrafficConditionLight
}

// speed calculates the speed based on the distance
// and the time in seconds it took to go that far
func speed(distance float64, timeTaken float64) float64 {
	if distance == 0 || timeTaken == 0 {
		return 0.00
	}

	return (distance / (timeTaken / 3600.00))
}

// distanceInKm calculates the distance in km between two points:
// http://andrew.hedges.name/experiments/haversine
func distanceInKm(x, y Point) float64 {
	dLon := deg2Rad(y.Lon - x.Lon)
	dLat := deg2Rad(y.Lat - x.Lat)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(deg2Rad(x.Lat))*
		math.Cos(deg2Rad(y.Lat))*math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := earthRadius * c

	return d
}

// deg2Rad is a helper func to convert degrees to radians
// (saves us having to do the calculation everywhere)
func deg2Rad(deg float64) float64 {
	return deg * (math.Pi / 180)
}
