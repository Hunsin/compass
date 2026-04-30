// Package statistics implements the gRPC StatisticsService.
package statistics

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	statsLib "github.com/Hunsin/compass/lib/statistics"
	pb "github.com/Hunsin/compass/protocols/gen/go/statistics/v1"
)

// Make sure Service implements the gRPC health server interface.
var _ healthpb.HealthServer = new(Service)

// Service implements the gRPC StatisticsService.
type Service struct {
	pb.UnimplementedStatisticsServiceServer
	model statsLib.Model
}

// New creates a new Service using the given model.
func New(m statsLib.Model) *Service {
	return &Service{model: m}
}

func (s *Service) Check(ctx context.Context, _ *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	res := healthpb.HealthCheckResponse_SERVING
	if err := s.model.Health(ctx); err != nil {
		res = healthpb.HealthCheckResponse_NOT_SERVING
	}
	return &healthpb.HealthCheckResponse{Status: res}, nil
}

func (s *Service) List(context.Context, *healthpb.HealthListRequest) (*healthpb.HealthListResponse, error) {
	return nil, status.Error(codes.Unimplemented, "List method unimplemented")
}

func (s *Service) Watch(*healthpb.HealthCheckRequest, grpc.ServerStreamingServer[healthpb.HealthCheckResponse]) error {
	return status.Error(codes.Unimplemented, "Watch method unimplemented")
}
