package internal

import (
	"microservicesDemo/L3-etcd/kitex_gen/post/postservice"
	"microservicesDemo/L3-etcd/kitex_gen/user/userservice"
	"time"

	"github.com/cloudwego/kitex/client"
	etcd "github.com/kitex-contrib/registry-etcd"
)

type Clients struct {
	UserClient userservice.Client
	PostClient postservice.Client
}

func InitClients() Clients {
	// 创建etcd解析器
	r, err := etcd.NewEtcdResolver([]string{"http://127.0.0.1:2379"})
	if err != nil {
		panic(err)
	}

	userClient, err := userservice.NewClient("user-service", client.WithResolver(r), client.WithRPCTimeout(5*time.Second))
	if err != nil {
		panic(err)
	}

	postClient, err := postservice.NewClient("post-service", client.WithResolver(r), client.WithRPCTimeout(5*time.Second))
	if err != nil {
		panic(err)
	}

	return Clients{
		UserClient: userClient,
		PostClient: postClient,
	}
}
