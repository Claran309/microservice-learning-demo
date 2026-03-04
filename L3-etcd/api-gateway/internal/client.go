package internal

import (
	"microservicesDemo/L3-etcd/kitex_gen/post/postservice"
	"microservicesDemo/L3-etcd/kitex_gen/user/userservice"
)

type Clients struct {
	UserClient userservice.Client
	PostClient postservice.Client
}

func InitClients() Clients {
	userClient, err := userservice.NewClient("user")
	if err != nil {
		panic(err)
	}

	postClient, err := postservice.NewClient("post")
	if err != nil {
		panic(err)
	}

	return Clients{
		UserClient: userClient,
		PostClient: postClient,
	}
}
