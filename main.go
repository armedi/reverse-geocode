package main

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("Error while reading config file %s", err)
	}

	geojsonFile := viper.GetString("GEOJSON_FILE")

	// read the geojson file
	file, err := os.Open(geojsonFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	// parse the geojson data into a FeatureCollection
	fc, err := geojson.UnmarshalFeatureCollection(contents)
	if err != nil {
		panic(err)
	}

	app := fiber.New()

	app.Get("/area/:coordinate", func(c *fiber.Ctx) error {
		coordinate := c.Params("coordinate")

		parts := strings.Split(strings.TrimPrefix(coordinate, "@"), ",")

		lat, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid coordinate")
		}

		lon, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid coordinate")
		}

		result, err := reverseGeocode(fc.Features, lat, lon)

		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString(err.Error())
		}

		return c.JSON(fiber.Map(result))
	})

	app.Listen(":3000")
}

func reverseGeocode(features []*geojson.Feature, lat float64, lon float64) (map[string]interface{}, error) {
	point := orb.Point{lon, lat}

	numWorkers := runtime.NumCPU()
	jobs := make([]chan *geojson.Feature, numWorkers)
	results := make(chan map[string]interface{}, len(features))

	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < numWorkers; i++ {
		jobs[i] = make(chan *geojson.Feature, len(features)/numWorkers)
		go reverseGeocodeWorker(ctx, point, jobs[i], results)
	}

	// assign jobs to workers using round-robin scheduling
	for i, feature := range features {
		jobs[i%numWorkers] <- feature
	}

	for i := 0; i < len(features); i++ {
		result := <-results
		if result != nil {
			cancel()
			return result, nil
		}
	}

	cancel()

	return nil, errors.New("not found")
}

func reverseGeocodeWorker(ctx context.Context, point orb.Point, jobs <-chan *geojson.Feature, results chan<- map[string]interface{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case feature := <-jobs:
			_, ok := feature.Geometry.(orb.Polygon)
			if ok && feature.Geometry.Bound().Contains(point) {
				results <- feature.Properties
			} else {
				results <- nil
			}
		}
	}
}
