package gapi

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcMetadataUserAgentHeader = "grpcgateway-user-agent"
	grpcMetadataClientIPHeader  = "x-forwarded-for"
	userAgentHeader             = "user-agent"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

func (server *Server) ExtractMetadata(ctx context.Context) *Metadata {
	meta := &Metadata{}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userAgents := md.Get(grpcMetadataUserAgentHeader); len(userAgents) > 0 {
			meta.UserAgent = userAgents[0]
		}

		if userAgents := md.Get(userAgentHeader); len(userAgents) > 0 {
			meta.UserAgent = userAgents[0]
		}

		if clientIPs := md.Get(grpcMetadataClientIPHeader); len(clientIPs) > 0 {
			meta.ClientIP = clientIPs[0]
		}
	}

	if p, ok := peer.FromContext(ctx); ok {
		meta.ClientIP = p.Addr.String()
	}

	return meta
}
