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

	numWorkers := runtime.NumCPU()
	featuresSet := make([][]*geojson.Feature, numWorkers)

	for i, feature := range fc.Features {
		featuresSet[i%numWorkers] = append(featuresSet[i%numWorkers], feature)
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

		result, err := reverseGeocode(featuresSet, lat, lon)

		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString(err.Error())
		}

		return c.JSON(fiber.Map(result))
	})

	app.Listen(":3000")
}

func reverseGeocode(featuresSet [][]*geojson.Feature, lat float64, lon float64) (map[string]interface{}, error) {
	point := orb.Point{lon, lat}

	results := make(chan map[string]interface{}, len(featuresSet))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, features := range featuresSet {
		jobs := make(chan *geojson.Feature, len(features))

		go func(features []*geojson.Feature) {
			for _, job := range features {
				select {
				case <-ctx.Done():
					return
				default:
					jobs <- job
				}
			}
		}(features)

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case job := <-jobs:
					_, ok := job.Geometry.(orb.Polygon)
					if ok && job.Geometry.Bound().Contains(point) {
						results <- job.Properties
					} else {
						results <- nil
					}
				}
			}
		}()
	}

	for i := 0; i < len(featuresSet); i++ {
		for j := 0; j < len(featuresSet[i]); j++ {
			result := <-results
			if result != nil {
				return result, nil
			}
		}
	}

	return nil, errors.New("not found")
}
