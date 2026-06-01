package grpchandler

import pbiamv1 "github.com/tuannm99/podzone/pkg/api/proto/iam/v1"

type IAMServer struct {
	*IAMCommandServer
	*IAMQueryServer
	iamUnimplemented
}

var _ pbiamv1.IAMServiceServer = (*IAMServer)(nil)

type iamUnimplemented struct {
	pbiamv1.UnimplementedIAMServiceServer
	pbiamv1.UnimplementedIAMCommandServiceServer
	pbiamv1.UnimplementedIAMQueryServiceServer
}

func NewIAMServer(commandServer *IAMCommandServer, queryServer *IAMQueryServer) *IAMServer {
	return &IAMServer{
		IAMCommandServer: commandServer,
		IAMQueryServer:   queryServer,
	}
}
