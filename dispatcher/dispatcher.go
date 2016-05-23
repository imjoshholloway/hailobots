package dispatcher

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/imjoshholloway/hailobots/robots"
)

const (
	// InstructionShutdown is our value for the SHUTDOWN instruction
	InstructionShutdown Instruction = "SHUTDOWN"
	dateFormat                      = "2006-01-02 15:04:05"
)

// Instruction provides a base type for Instructions sent within the program
type Instruction string

// New creates our dispatcher & starts processing the sources,
// passing the Points received to each Robot
func New(robots map[int]*robots.Robot, sources []io.Reader) *Dispatcher {
	d := &Dispatcher{
		Robots:      robots,
		Sources:     sources,
		Instruction: make(chan Instruction),
		Terminate:   make(chan bool),
	}

	go d.Run()
	return d
}

// Dispatcher controls the Robot's execution, loading Points from the sources and
// passing them to each Robot. It also provides communication between the Robots
// and the Reporter
type Dispatcher struct {
	Robots map[int]*robots.Robot
	// Readers is a map of the file Reader for each robot where key = the Robot.ID
	// and the value = the io.Reader instance
	Sources     []io.Reader
	Instruction chan Instruction
	Terminate   chan bool
}

// Run starts the Dispatcher, controlling the instructions and processing of
// the sources
func (d *Dispatcher) Run() {
	for {
		select {
		case inst, ok := <-d.Instruction:
			if !ok {
				log.Fatal("Dispatcher: error processing instruction")
			}
			if inst == InstructionShutdown {
				fmt.Printf("Dispatcher: received SHUTDOWN signal\n")
				for _, r := range d.Robots {
					fmt.Printf("Dispatcher: shutting down robot: %d\n", r.ID)
					close(r.Next)
					// Whilst the robot is still running, don't kill it!
					<-r.Shutdown
				}
				d.Terminate <- true
			}
		}
	}
}

// Process feeds data from the dispatcher Sources to each of the Robots
func (d *Dispatcher) Process() {
	var wg sync.WaitGroup
	for _, reader := range d.Sources {
		wg.Add(1)
		go func(reader io.Reader) {
			defer wg.Done()
			csvReader := csv.NewReader(reader)
			csvReader.FieldsPerRecord = 4
			for {
				line, err := csvReader.Read()
				if err == io.EOF {
					log.Printf("Dispatcher: end of source reached")
					break
				}
				// TODO: handle errors
				if err != nil {
					log.Printf("Dispatcher: error processing line in source: %v", err)
					continue
				}

				id, err := strconv.Atoi(line[0])
				if err != nil {
					log.Printf("Dispatcher: skipping line - error processing ID from source: %v", err)
					continue
				}

				lat, err := strconv.ParseFloat(line[1], 64)
				if err != nil {
					log.Printf("Dispatcher: skipping line - error processing Lat from source: %v", err)
					continue
				}

				lon, err := strconv.ParseFloat(line[2], 64)
				if err != nil {
					log.Printf("Dispatcher: skipping line - error processing Lon from source: %v", err)
					continue
				}

				routeTime, err := time.Parse(dateFormat, line[3])
				if err != nil {
					log.Printf("Dispatcher: skipping line - error processing Time from source: %v", err)
					continue
				}

				robot, ok := d.Robots[id]

				if !ok {
					log.Printf("Dispatcher: skipping line - Robot: %d not found: %v", id, err)
					continue
				}

				if routeTime.Hour() == 8 && routeTime.Minute() == 10 {
					fmt.Printf("Dispatcher: Robot: %d reached 8:10. Terminating \n", id)
					robot.Shutdown <- true
					break
				}

				point := robots.Point{Lat: lat, Lon: lon}
				routePoint := &robots.RoutePoint{
					Time:  routeTime,
					Point: point,
				}

				robot.Next <- routePoint
				fmt.Printf("Dispatcher: sent Robot: %d to point (lat/lon): %v/%v @ Time: %v \n", robot.ID, routePoint.Point.Lat, routePoint.Point.Lon, routePoint.Time)
			}
		}(reader)

		wg.Wait()

	}

	fmt.Println("Dispatcher: finished loading sources. Shutting down")
	d.Instruction <- InstructionShutdown
}
