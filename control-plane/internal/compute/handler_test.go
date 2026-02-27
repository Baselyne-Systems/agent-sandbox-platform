package compute

import (
	"context"
	"testing"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/models"
	pb "github.com/baselyne/agent-sandbox-platform/control-plane/pkg/gen/compute/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestHandler() *Handler {
	return NewHandler(NewService(newMockRepo()))
}

func registerTestHost(t *testing.T, h *Handler) *pb.Host {
	t.Helper()
	resp, err := h.RegisterHost(context.Background(), &pb.RegisterHostRequest{
		Address: "runtime.host1:50052",
		TotalResources: &pb.HostResources{
			MemoryMb:      16384,
			CpuMillicores: 8000,
			DiskMb:        102400,
		},
	})
	if err != nil {
		t.Fatalf("RegisterHost: %v", err)
	}
	return resp.Host
}

func TestHandler_RegisterHost_Success(t *testing.T) {
	h := newTestHandler()
	host := registerTestHost(t, h)

	if host.HostId == "" {
		t.Error("expected host ID")
	}
	if host.Address != "runtime.host1:50052" {
		t.Errorf("address = %q, want 'runtime.host1:50052'", host.Address)
	}
	if host.Status != pb.HostStatus_HOST_STATUS_READY {
		t.Errorf("status = %v, want READY", host.Status)
	}
	if host.TotalResources.MemoryMb != 16384 {
		t.Errorf("total memory = %d, want 16384", host.TotalResources.MemoryMb)
	}
	if host.AvailableResources.MemoryMb != 16384 {
		t.Errorf("available memory = %d, want 16384", host.AvailableResources.MemoryMb)
	}
	if host.ActiveSandboxes != 0 {
		t.Errorf("active_sandboxes = %d, want 0", host.ActiveSandboxes)
	}
	if host.LastHeartbeat == nil {
		t.Error("expected last_heartbeat timestamp")
	}
}

func TestHandler_RegisterHost_NilResources(t *testing.T) {
	h := newTestHandler()
	_, err := h.RegisterHost(context.Background(), &pb.RegisterHostRequest{
		Address:        "host:50052",
		TotalResources: nil,
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestHandler_RegisterHost_InvalidResources(t *testing.T) {
	h := newTestHandler()
	_, err := h.RegisterHost(context.Background(), &pb.RegisterHostRequest{
		Address: "host:50052",
		TotalResources: &pb.HostResources{
			MemoryMb:      0,
			CpuMillicores: 8000,
			DiskMb:        1024,
		},
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestHandler_DeregisterHost_Success(t *testing.T) {
	h := newTestHandler()
	host := registerTestHost(t, h)

	_, err := h.DeregisterHost(context.Background(), &pb.DeregisterHostRequest{
		HostId: host.HostId,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify status changed to offline
	resp, _ := h.ListHosts(context.Background(), &pb.ListHostsRequest{
		Status: pb.HostStatus_HOST_STATUS_OFFLINE,
	})
	if len(resp.Hosts) != 1 {
		t.Errorf("offline hosts = %d, want 1", len(resp.Hosts))
	}
}

func TestHandler_DeregisterHost_NotFound(t *testing.T) {
	h := newTestHandler()
	_, err := h.DeregisterHost(context.Background(), &pb.DeregisterHostRequest{
		HostId: "nonexistent",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}

func TestHandler_ListHosts_Success(t *testing.T) {
	h := newTestHandler()
	registerTestHost(t, h)
	registerTestHost(t, h)

	resp, err := h.ListHosts(context.Background(), &pb.ListHostsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Hosts) != 2 {
		t.Errorf("hosts = %d, want 2", len(resp.Hosts))
	}
}

func TestHandler_ListHosts_StatusFilter(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	host := registerTestHost(t, h)
	registerTestHost(t, h)
	h.DeregisterHost(ctx, &pb.DeregisterHostRequest{HostId: host.HostId})

	resp, err := h.ListHosts(ctx, &pb.ListHostsRequest{
		Status: pb.HostStatus_HOST_STATUS_READY,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Hosts) != 1 {
		t.Errorf("ready hosts = %d, want 1", len(resp.Hosts))
	}
}

func TestHandler_PlaceWorkspace_Success(t *testing.T) {
	h := newTestHandler()
	registerTestHost(t, h)

	resp, err := h.PlaceWorkspace(context.Background(), &pb.PlaceWorkspaceRequest{
		MemoryMb:      1024,
		CpuMillicores: 500,
		DiskMb:        2048,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.HostId == "" {
		t.Error("expected host ID")
	}
	if resp.RuntimeEndpoint == "" {
		t.Error("expected runtime endpoint")
	}
}

func TestHandler_PlaceWorkspace_NoCapacity(t *testing.T) {
	h := newTestHandler()
	// No hosts registered — should fail with no capacity

	_, err := h.PlaceWorkspace(context.Background(), &pb.PlaceWorkspaceRequest{
		MemoryMb:      1024,
		CpuMillicores: 500,
		DiskMb:        2048,
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.ResourceExhausted {
		t.Errorf("code = %v, want ResourceExhausted", st.Code())
	}
}

func TestHandler_HostStatusConversion_RoundTrip(t *testing.T) {
	tests := []struct {
		proto pb.HostStatus
		model models.HostStatus
	}{
		{pb.HostStatus_HOST_STATUS_READY, models.HostStatusReady},
		{pb.HostStatus_HOST_STATUS_DRAINING, models.HostStatusDraining},
		{pb.HostStatus_HOST_STATUS_OFFLINE, models.HostStatusOffline},
	}
	for _, tt := range tests {
		got := protoHostStatusToModel(tt.proto)
		if got != tt.model {
			t.Errorf("protoHostStatusToModel(%v) = %q, want %q", tt.proto, got, tt.model)
		}
		back := modelHostStatusToProto(got)
		if back != tt.proto {
			t.Errorf("modelHostStatusToProto(%q) = %v, want %v", got, back, tt.proto)
		}
	}
}
