package main

import (
	"fmt"
	"log"
	"net"
	"time"

	accesslog "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v2"
	"github.com/fluent/fluent-logger-golang/fluent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//Struct would implement envoy grpc AccessLogServer
type accesslogserver struct {
}

// EnvoyAccessLogFluentd is the json structure which we want to sent to fluentd
type EnvoyAccessLogFluentd struct {
	RouteName       string
	UpstreamCluster string
}

var fluentDLogger *fluent.Fluent
var envoyAccessLogFluentd EnvoyAccessLogFluentd

func (als *accesslogserver) StreamAccessLogs(alserver accesslog.AccessLogService_StreamAccessLogsServer) error {

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

	//Receives the accesslogs from envoy
	accesslogStreamMessage, err := alserver.Recv()
	if err != nil {
		log.Println("Error occured : ", err)
		return err
	}

	envoyAccessLogs := accesslogStreamMessage.GetHttpLogs()
	envoyAccessLogsIdentifier := accesslogStreamMessage.GetIdentifier()

	for _, httpLog := range envoyAccessLogs.GetLogEntry() {

		commonProps := httpLog.GetCommonProperties()
		envoyAccessLogFluentd = EnvoyAccessLogFluentd{
			RouteName:       commonProps.GetRouteName(),
			UpstreamCluster: commonProps.GetUpstreamCluster(),
		}

		e := fluentDLogger.Post(envoyAccessLogsIdentifier.GetLogName(), envoyAccessLogFluentd)
		if e != nil {
			log.Printf("Unable to post access Log %v", e)
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
	err := retry(5, 5*time.Second, func() (err error) {
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

	// TCP listener on port 10002
	lis, err := net.Listen("tcp", ":10002")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Access Log Service listening on %s \n", lis.Addr())

	s := grpc.NewServer()

	accesslog.RegisterAccessLogServiceServer(s, &accesslogserver{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
