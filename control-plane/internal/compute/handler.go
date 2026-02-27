package compute

import (
	"context"
	"errors"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/compute/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler implements the ComputePlaneServiceServer gRPC interface.
type Handler struct {
	pb.UnimplementedComputePlaneServiceServer
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterHost(ctx context.Context, req *pb.RegisterHostRequest) (*pb.RegisterHostResponse, error) {
	res := req.GetTotalResources()
	if res == nil {
		return nil, status.Error(codes.InvalidArgument, "total_resources is required")
	}
	host, err := h.svc.RegisterHost(ctx, req.GetAddress(), models.HostResources{
		MemoryMb:      res.GetMemoryMb(),
		CpuMillicores: res.GetCpuMillicores(),
		DiskMb:        res.GetDiskMb(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.RegisterHostResponse{Host: hostToProto(host)}, nil
}

func (h *Handler) DeregisterHost(ctx context.Context, req *pb.DeregisterHostRequest) (*pb.DeregisterHostResponse, error) {
	if err := h.svc.DeregisterHost(ctx, req.GetHostId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.DeregisterHostResponse{}, nil
}

func (h *Handler) ListHosts(ctx context.Context, req *pb.ListHostsRequest) (*pb.ListHostsResponse, error) {
	statusFilter := protoHostStatusToModel(req.GetStatus())
	hosts, err := h.svc.ListHosts(ctx, statusFilter)
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbHosts := make([]*pb.Host, len(hosts))
	for i := range hosts {
		pbHosts[i] = hostToProto(&hosts[i])
	}
	return &pb.ListHostsResponse{Hosts: pbHosts}, nil
}

func (h *Handler) PlaceWorkspace(ctx context.Context, req *pb.PlaceWorkspaceRequest) (*pb.PlaceWorkspaceResponse, error) {
	hostID, address, err := h.svc.PlaceWorkspace(ctx, req.GetMemoryMb(), req.GetCpuMillicores(), req.GetDiskMb())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.PlaceWorkspaceResponse{
		HostId:          hostID,
		RuntimeEndpoint: address,
	}, nil
}

func (h *Handler) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	host, err := h.svc.Heartbeat(ctx, req.GetHostId(), models.HostResources{
		MemoryMb:      req.GetAvailableMemoryMb(),
		CpuMillicores: req.GetAvailableCpuMillicores(),
		DiskMb:        req.GetAvailableDiskMb(),
	}, req.GetActiveSandboxes())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.HeartbeatResponse{
		Status: modelHostStatusToProto(host.Status),
	}, nil
}

// --- converters ---

func hostToProto(h *models.Host) *pb.Host {
	return &pb.Host{
		HostId:  h.ID,
		Address: h.Address,
		Status:  modelHostStatusToProto(h.Status),
		TotalResources: &pb.HostResources{
			MemoryMb:      h.TotalResources.MemoryMb,
			CpuMillicores: h.TotalResources.CpuMillicores,
			DiskMb:        h.TotalResources.DiskMb,
		},
		AvailableResources: &pb.HostResources{
			MemoryMb:      h.AvailableResources.MemoryMb,
			CpuMillicores: h.AvailableResources.CpuMillicores,
			DiskMb:        h.AvailableResources.DiskMb,
		},
		ActiveSandboxes: h.ActiveSandboxes,
		LastHeartbeat:   timestamppb.New(h.LastHeartbeat),
	}
}

func modelHostStatusToProto(s models.HostStatus) pb.HostStatus {
	switch s {
	case models.HostStatusReady:
		return pb.HostStatus_HOST_STATUS_READY
	case models.HostStatusDraining:
		return pb.HostStatus_HOST_STATUS_DRAINING
	case models.HostStatusOffline:
		return pb.HostStatus_HOST_STATUS_OFFLINE
	default:
		return pb.HostStatus_HOST_STATUS_UNSPECIFIED
	}
}

func protoHostStatusToModel(s pb.HostStatus) models.HostStatus {
	switch s {
	case pb.HostStatus_HOST_STATUS_READY:
		return models.HostStatusReady
	case pb.HostStatus_HOST_STATUS_DRAINING:
		return models.HostStatusDraining
	case pb.HostStatus_HOST_STATUS_OFFLINE:
		return models.HostStatusOffline
	default:
		return ""
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, ErrHostNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, ErrNoCapacity):
		return status.Error(codes.ResourceExhausted, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
