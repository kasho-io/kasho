package testutil

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"kasho/proto/kasho/proto"
)

const bufSize = 1024 * 1024

// MockLicenseServer implements a mock licensing server for testing
type MockLicenseServer struct {
	proto.UnimplementedLicenseServer
	ValidResponse     *proto.ValidateLicenseResponse
	LicenseInfoResponse *proto.GetLicenseInfoResponse
}

func (m *MockLicenseServer) ValidateLicense(ctx context.Context, req *proto.ValidateLicenseRequest) (*proto.ValidateLicenseResponse, error) {
	if m.ValidResponse != nil {
		return m.ValidResponse, nil
	}
	return &proto.ValidateLicenseResponse{
		Valid:     true,
		Reason:    "",
		ExpiresAt: 0,
	}, nil
}

func (m *MockLicenseServer) GetLicenseInfo(ctx context.Context, req *proto.GetLicenseInfoRequest) (*proto.GetLicenseInfoResponse, error) {
	if m.LicenseInfoResponse != nil {
		return m.LicenseInfoResponse, nil
	}
	return &proto.GetLicenseInfoResponse{
		CustomerId:   "test-customer",
		CustomerName: "Test Customer",
		Valid:        true,
	}, nil
}

// StartMockServer starts a mock license server for testing
func StartMockServer(t *testing.T, mockServer *MockLicenseServer) (*grpc.ClientConn, func()) {
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	proto.RegisterLicenseServer(s, mockServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	cleanup := func() {
		conn.Close()
		s.Stop()
		lis.Close()
	}

	return conn, cleanup
}