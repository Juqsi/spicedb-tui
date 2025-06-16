package client

import (
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"spicedb-tui/internal/config"
)

var Client *authzed.Client

func InitClient() error {
	var err error
	Client, err = authzed.NewClient(
		config.Current.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(config.Current.Token),
	)
	return err
}
