package api

import (
	"context"
	"log"
	"net"
	"net/http"
	"therealbroker/api/proto"
	module "therealbroker/internal/broker"
	"therealbroker/metrics"
	br "therealbroker/pkg/broker"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const port = "8080"
const metric_port = "8081"

type Server struct {
	proto.UnimplementedBrokerServer
}

var broker = module.NewModule()

func ConvertError(err error) error {

	switch err {
	case br.ErrUnavailable:
		return status.Error(codes.Unavailable, err.Error())
	case br.ErrInvalidID:
		return status.Error(codes.InvalidArgument, err.Error())
	case br.ErrExpiredID:
		return status.Error(codes.InvalidArgument, err.Error())
	case nil:
		return nil
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *Server) Publish(ctx context.Context, request *proto.PublishRequest) (*proto.PublishResponse, error) {

	start_time := time.Now()
	defer metrics.Latency.WithLabelValues("publish").Observe(float64(time.Since(start_time).Nanoseconds()))

	msg := br.Message{
		Body:       string(request.Body),
		Expiration: time.Duration(request.ExpirationSeconds) * time.Second,
	}

	ID, err := broker.Publish(ctx, request.Subject, msg)
	if err != nil {
		metrics.FailedCalls.WithLabelValues("publish").Inc()
		return nil, ConvertError(err)
	}

	metrics.SuccessfullCalls.WithLabelValues("publish").Inc()
	return &proto.PublishResponse{Id: int32(ID)}, ConvertError(err)
}

func (s *Server) Subscribe(request *proto.SubscribeRequest, stream proto.Broker_SubscribeServer) error {

	start_time := time.Now()
	defer metrics.Latency.WithLabelValues("subscribe").Observe(float64(time.Since(start_time).Nanoseconds()))

	ctx := stream.Context()
	topic, err := broker.Subscribe(ctx, request.Subject)
	if err != nil {
		metrics.FailedCalls.WithLabelValues("subscribe").Inc()
		return ConvertError(err)
	}
	metrics.Subscriptions.Inc()
	for message := range topic {
		response := &proto.MessageResponse{Body: message.Body}
		if err := stream.Send(response); err != nil {
			metrics.FailedCalls.WithLabelValues("subscribe").Inc() // should it be here?
			return ConvertError(err)
		}
	}

	metrics.SuccessfullCalls.WithLabelValues("subscribe").Inc()
	metrics.Subscriptions.Dec()
	return nil
}

func (s *Server) Fetch(ctx context.Context, request *proto.FetchRequest) (*proto.MessageResponse, error) {

	start_time := time.Now()
	defer metrics.Latency.WithLabelValues("fetch").Observe(float64(time.Since(start_time).Nanoseconds()))

	msg, err := broker.Fetch(ctx, request.Subject, int(request.Id))
	if err != nil {
		metrics.FailedCalls.WithLabelValues("fetch").Inc()
		return nil, ConvertError(err)
	}
	metrics.SuccessfullCalls.WithLabelValues("fetch").Inc()

	return &proto.MessageResponse{Body: msg.Body}, ConvertError(err)
}

func StartGrpcServer() {

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":"+metric_port, nil))
	}()

	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	proto.RegisterBrokerServer(server, &Server{})
	server.Serve(listen)
}
