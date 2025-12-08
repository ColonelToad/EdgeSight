package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	broker := flag.String("broker", "tcp://localhost:1883", "MQTT broker URL")
	interval := flag.Duration("interval", 2*time.Second, "publish interval")
	noise := flag.Float64("noise", 0.05, "noise level +/- fraction")
	baseTemp := flag.Float64("temp", 22.0, "base temperature C")
	baseHum := flag.Float64("hum", 45.0, "base humidity %")
	basePM := flag.Float64("pm25", 12.0, "base PM2.5")
	basePower := flag.Float64("power", 1200.0, "base power W")
	flag.Parse()

	opts := mqtt.NewClientOptions().AddBroker(*broker).SetClientID("edgesight-sim")
	cli := mqtt.NewClient(opts)
	if token := cli.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	defer cli.Disconnect(50)

	rand.Seed(time.Now().UnixNano())

	for {
		publish(cli, "sensors/temperature", jitter(*baseTemp, *noise))
		publish(cli, "sensors/humidity", jitter(*baseHum, *noise))
		publish(cli, "sensors/pm25", jitter(*basePM, *noise))
		publish(cli, "sensors/power", jitter(*basePower, *noise))
		time.Sleep(*interval)
	}
}

func publish(cli mqtt.Client, topic string, val float64) {
	payload := fmt.Sprintf("%.3f", val)
	cli.Publish(topic, 1, false, payload)
}

func jitter(base, noise float64) float64 {
	return base * (1 + noise*(rand.Float64()*2-1))
}
