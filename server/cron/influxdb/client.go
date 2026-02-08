package influxdb

import (
	"context"
	"log/slog"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type Config struct {
	Token  string `json:"token"`
	Url    string `json:"url"`
	Bucket string `json:"bucket"`
}

type Client struct {
	config Config
	client influxdb2.Client
}

type Temperature struct {
	Room        string  `json:"room"`
	Temperature float64 `json:"temperature"`
	Humidity    int     `json:"humidity"`
	Battery     int     `json:"battery"`
	Co2         int     `json:"co2"`
}

func NewClient(config Config) Client {
	return Client{
		config: config,
		client: influxdb2.NewClient(config.Url, config.Token),
	}
}

func (c Client) WriteTemperature(data Temperature) {
	writeAPI := c.client.WriteAPIBlocking("personal", c.config.Bucket)
	tags := map[string]string{
		"room": data.Room,
	}
	fields := map[string]interface{}{
		"temperature": data.Temperature,
		"humidity":    data.Humidity,
		"battery":     data.Battery,
		"co2":         data.Co2,
	}
	point := write.NewPoint("temperature", tags, fields, time.Now())

	if err := writeAPI.WritePoint(context.Background(), point); err != nil {
		slog.Error("Error for writing data to InfluxDB", "error", err)
	}
}

func (c Client) Close() {
	c.client.Close()
}
