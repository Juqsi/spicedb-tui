package client

import (
	"context"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
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
	go func() {
		_, err = Client.ReadSchema(context.Background(), &v1.ReadSchemaRequest{})
	}()
	return err
}
