package main

import (
	"context"
	"flag"
	kubernetes "github.com/gavinbunney/terraform-provider-kubectl/kubernetes"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	tf5server "github.com/hashicorp/terraform-plugin-go/tfprotov5/server"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"google.golang.org/grpc"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debuggable", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{ProviderFunc: kubernetes.Provider}
	grpcProviderFunc := func() tfprotov5.ProviderServer {
		return schema.NewGRPCProviderServer(kubernetes.Provider())
	}

	if debugMode {
		plugin.Debug(context.Background(), "gavinbunney/kubectl", opts)
	}

	// taken from github.com/hashicorp/terraform-plugin-sdk/v2@v2.3.0/plugin/serve.go
	// configured to allow larger message sizes than 4mb
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: plugin.Handshake,
		VersionedPlugins: map[int]goplugin.PluginSet{
			5: {
				plugin.ProviderPluginName: &tf5server.GRPCProviderPlugin{
					GRPCProvider: func() tfprotov5.ProviderServer {
						return grpcProviderFunc()
					},
				},
			},
		},
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			return grpc.NewServer(append(opts,
				grpc.MaxSendMsgSize(64<<20 /* 64MB */),
				grpc.MaxRecvMsgSize(64<<20 /* 64MB */))...)
		},
		Logger: opts.Logger,
		Test:   opts.TestConfig,
	})
}
