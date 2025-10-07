package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	greetv1 "github.com/yebis0942/grpc-sandbox/gen/proto/greet/v1"
	"github.com/yebis0942/grpc-sandbox/gen/proto/greet/v1/greetv1connect"
)

type GreetServer struct{}

func (s *GreetServer) Greet(
	ctx context.Context,
	req *connect.Request[greetv1.GreetRequest],
) (*connect.Response[greetv1.GreetResponse], error) {
	log.Println("Request headers: ", req.Header())

	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name is required"))
	}

	greeting := fmt.Sprintf("Hello, %s!", req.Msg.Name)
	if req.Msg.Title != nil {
		greeting = fmt.Sprintf("Hello, %s %s!", *req.Msg.Title, req.Msg.Name)
	}

	res := connect.NewResponse(&greetv1.GreetResponse{
		Greeting: greeting,
	})
	res.Header().Set("Greet-Version", "v1")
	return res, nil
}

func main() {
	greeter := &GreetServer{}
	mux := http.NewServeMux()
	path, handler := greetv1connect.NewGreetServiceHandler(greeter)
	mux.Handle(path, handler)
	reflector := grpcreflect.NewStaticReflector(
		"greet.v1.GreetService",
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	log.Println("Starting server on :8085")
	err := http.ListenAndServe(
		"localhost:8085",
		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(mux, &http2.Server{}),
	)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
