package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"encoding/json"

	ms "github.com/envoyproxy/go-control-plane/envoy/service/metrics/v2"
	"github.com/fluent/fluent-logger-golang/fluent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var fluentDLogger *fluent.Fluent

//Struct would implement envoy grpc MetricServiceServer
type server struct {
}

// EnvoyLogFluentd is the json structure which we want to sent to fluentd
// type EnvoyLogFluentd struct {
// 	Name string
// 	Type string
// }

var envoyMetricsPayload map[string]interface{}

//envoyMetricsTagPrefix would be the tag used to send the metrics to fluentd service
const envoyMetricsTagPrefix string = "envoymetrics"

//StreamMetrics receives the metrics from envoy and post extracting the relevant information
// sends it to fluentd server
func (s *server) StreamMetrics(mtserver ms.MetricsService_StreamMetricsServer) error {

	//Acquire the logger during the first call
	if fluentDLogger == nil {
		fluentdErr := acquireFluentDLogger()
		if fluentdErr != nil {
			log.Printf("Error occured for obtaining fluentd logger %v \n", fluentdErr)
			return fluentdErr
		} else {
			log.Printf("Fluentd agent running on : %s : %d \n", fluentDLogger.FluentHost, fluentDLogger.FluentPort)
		}

	}

	//Receives the metrics from envoy
	streamMetricsMessage, err := mtserver.Recv()
	if err != nil {
		log.Println("Error occured : ", err)
		return err
	}

	//Gather all the metrices from current stream
	envoyMetrics := streamMetricsMessage.GetEnvoyMetrics()
	//envoyIdentifier := streamMetricsMessage.GetIdentifier().String()
	for _, metric := range envoyMetrics {

		//Extracts all the relevant information from the metric and post it to fluentd
		// envoyMetricFluentdLog := EnvoyLogFluentd{
		// 	Name: metric.GetName(),
		// 	Type: metric.GetType().String(),
		// }

		inrec, _ := json.Marshal(metric)
		json.Unmarshal(inrec, &envoyMetricsPayload)

		e := fluentDLogger.Post(envoyMetricsTagPrefix, envoyMetricsPayload)
		if e != nil {
			log.Printf("Unable to post metrics data to fluentd for metric %v \n", e)
			return e
		}

	}

	return nil
}

// Obtain the fluentd log instance
func acquireFluentDLogger() error {

	//Implementing retries to obtain fluentd agent
	//Set up the logger for fluentd
	log.Println("Attempting to initiate connection to fluentd server................")
	var fluentdLogerrErr error
	//5 attempts would be made at a gap of 5 seconds to establish connection with the fluentd service
	err := retry(10, 5*time.Second, func() (err error) {
		fluentDLogger, fluentdLogerrErr = fluent.New(fluent.Config{FluentPort: 24224, FluentHost: "fluentd"})
		return fluentdLogerrErr
	})

	return err
}

//Simple retry function. Reference link https://blog.abourget.net/en/2016/01/04/my-favorite-golang-retry-function
func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)

		log.Println("retrying after error:", err)
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

func main() {

	// TCP listener on port 10001
	lis, err := net.Listen("tcp", ":10001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Metric Service listening on %s \n", lis.Addr())

	s := grpc.NewServer()
	ms.RegisterMetricsServiceServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
