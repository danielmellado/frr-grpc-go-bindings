/*
 * Copyright 2021 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/danielmellado/frr-grpc-go-bindings/frr"
	"google.golang.org/grpc"
)

var(
	serverAddr = flag.String("server_addr", "localhost:50051", "The server address in the format of host:port")
)

func main(){
	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil{
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	NorthBoundClient := pb.NewNorthboundClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var frrNBClientPath = []string{"/frr-routing:routing/control-plane-protocols/control-plane-protocol/frr-bgp:bgp"}
	frrConfigRequest := new(pb.GetRequest)
	frrConfigRequest.Type = pb.GetRequest_CONFIG
	frrConfigRequest.Encoding = pb.Encoding_JSON
	frrConfigRequest.WithDefaults = true
	frrConfigRequest.Path = frrNBClientPath
	frrNBClient, err := NorthBoundClient.Get(ctx, frrConfigRequest)
	fmt.Println("Request: ", frrConfigRequest.String())

	waitc := make(chan struct{})
	go func() {
		for {
			in, err := frrNBClient.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to get data: %v", err)
			}
			fmt.Printf("data %v", in)
		}
	}()


	frrConfigResponse, err := frrNBClient.Recv()
	if err != nil {
		log.Fatalf("Error getting response %v", err)
	}
	fmt.Println("Timestamp", frrConfigResponse.Timestamp )
}

