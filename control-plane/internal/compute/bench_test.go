package compute

import (
	"context"
	"fmt"
	"testing"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

func BenchmarkRegisterHost(b *testing.B) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	for b.Loop() {
		_, err := svc.RegisterHost(ctx, "host1:9090", models.HostResources{
			MemoryMb: 16384, CpuMillicores: 8000, DiskMb: 102400,
		}, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPlaceWorkspace(b *testing.B) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	svc.RegisterHost(ctx, "host1:9090", models.HostResources{
		MemoryMb: 1 << 30, CpuMillicores: 1 << 20, DiskMb: 1 << 30,
	}, nil)

	for b.Loop() {
		_, _, err := svc.PlaceWorkspace(ctx, 512, 500, 1024, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListHosts_Scaling(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			svc := NewService(newMockRepo())
			ctx := context.Background()

			for i := 0; i < n; i++ {
				svc.RegisterHost(ctx, fmt.Sprintf("host%d:9090", i), models.HostResources{
					MemoryMb: 16384, CpuMillicores: 8000, DiskMb: 102400,
				}, nil)
			}

			b.ResetTimer()
			for b.Loop() {
				_, err := svc.ListHosts(ctx, "")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDeregisterHost(b *testing.B) {
	ctx := context.Background()

	for b.Loop() {
		b.StopTimer()
		svc := NewService(newMockRepo())
		host, _ := svc.RegisterHost(ctx, "host1:9090", models.HostResources{
			MemoryMb: 16384, CpuMillicores: 8000, DiskMb: 102400,
		}, nil)
		b.StartTimer()

		err := svc.DeregisterHost(ctx, host.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPlaceWorkspace_LargeFleet(b *testing.B) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	for i := 0; i < 1000; i++ {
		svc.RegisterHost(ctx, fmt.Sprintf("host%d:9090", i), models.HostResources{
			MemoryMb: int64(1024 + i), CpuMillicores: int32(1000 + i), DiskMb: int64(10240 + i),
		}, nil)
	}

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.PlaceWorkspace(ctx, 512, 500, 1024, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}
