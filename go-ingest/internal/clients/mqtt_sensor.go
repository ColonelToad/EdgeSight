package clients

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTSensorReading holds last values seen on subscribed topics.
type MQTTSensorReading struct {
	Temperature float64
	Humidity    float64
	PM25        float64
	Power       float64
}

// MQTTSensorClient subscribes to sensor topics and returns the latest readings.
type MQTTSensorClient struct {
	broker   string
	clientID string
	topics   []string
	timeout  time.Duration
}

// NewMQTTSensorClient creates a new client.
func NewMQTTSensorClient(broker string) *MQTTSensorClient {
	return &MQTTSensorClient{
		broker:   broker,
		clientID: "edgesight-ingest",
		topics: []string{
			"sensors/temperature",
			"sensors/humidity",
			"sensors/pm25",
			"sensors/power",
		},
		timeout: 3 * time.Second,
	}
}

// FetchReadings connects, subscribes, waits briefly for messages, and returns the latest values.
func (c *MQTTSensorClient) FetchReadings() (*MQTTSensorReading, error) {
	if c.broker == "" {
		return nil, fmt.Errorf("mqtt broker not configured")
	}

	opts := mqtt.NewClientOptions().AddBroker(c.broker).SetClientID(c.clientID)
	mc := mqtt.NewClient(opts)

	if token := mc.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("mqtt connect: %w", token.Error())
	}
	defer mc.Disconnect(50)

	reading := &MQTTSensorReading{}
	mu := sync.Mutex{}
	var wg sync.WaitGroup

	handler := func(_ mqtt.Client, msg mqtt.Message) {
		mu.Lock()
		switch msg.Topic() {
		case "sensors/temperature":
			reading.Temperature = parseFloatBytes(msg.Payload())
		case "sensors/humidity":
			reading.Humidity = parseFloatBytes(msg.Payload())
		case "sensors/pm25":
			reading.PM25 = parseFloatBytes(msg.Payload())
		case "sensors/power":
			reading.Power = parseFloatBytes(msg.Payload())
		}
		mu.Unlock()
	}

	for _, t := range c.topics {
		wg.Add(1)
		if token := mc.Subscribe(t, 1, func(cl mqtt.Client, m mqtt.Message) {
			handler(cl, m)
			wg.Done()
		}); token.Wait() && token.Error() != nil {
			return nil, fmt.Errorf("mqtt subscribe %s: %w", t, token.Error())
		}
	}

	// Wait up to timeout for first messages; then return whatever was received.
	waitCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
	case <-time.After(c.timeout):
	}

	return reading, nil
}

func parseFloatBytes(b []byte) float64 {
	v, _ := strconv.ParseFloat(string(b), 64)
	return v
}
