package balancer

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(&exampleResolverBuilder{})
	/*
	   //注册的时候将Scheme => builder保存到m
	   func Register(b Builder) {
	       m[b.Scheme()] = b
	   }
	*/
}

const (
	exampleScheme      = "example"
	exampleServiceName = "lb.example.grpc.io"
)

var addrs = []string{"localhost:50051", "localhost:50052"}

type exampleResolverBuilder struct{}

func (*exampleResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &exampleResolver{
		target: target,
		cc:     cc,
		addrsStore: map[string][]string{
			exampleServiceName: addrs,
		},
	}
	r.start()
	return r, nil
}
func (*exampleResolverBuilder) Scheme() string { return exampleScheme }

type exampleResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

func (r *exampleResolver) start() {
	addrStrs := r.addrsStore[r.target.Endpoint]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{Addr: s}
	}

	//UpdateState()将addr更新到cc，也就是外部的连接中，供其他接口使用。
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}

// 被 gRPC 调用，以尝试再次解析目标名称。只用于提示，可忽略该方法。
func (*exampleResolver) ResolveNow(o resolver.ResolveNowOptions) {}

// 关闭resolver
func (*exampleResolver) Close() {}

func maintest() {
	// grpc.Dial() 会调用Scheme=>builder 的Build() 方法，之后调用r.start()
	//...
	roundrobinConn, err := grpc.Dial(
		// Target{Scheme:exampleScheme,Endpoint:exampleServiceName}
		fmt.Sprintf("%s:///%s", exampleScheme, exampleServiceName),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	//...
}
