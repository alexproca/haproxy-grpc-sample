package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	// The Protobuf generated file
	creator "app/codenamecreator"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

var addressHTTP = os.Getenv("SERVER_ADDRESS_HTTP")

func main() {
	address := os.Getenv("SERVER_ADDRESS")
	crt := os.Getenv("TLS_CERT")

	creds, err := credentials.NewClientTLSFromFile(crt, "")
	if err != nil {
		log.Fatalf("Failed to load TLS certificate")
	}

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Did not connect, %v", err)
	}
	defer conn.Close()

	client := creator.NewCodenameCreatorClient(conn)
	ctx := context.Background()

	// Optional: add some metadata
	ctx = metadata.AppendToOutgoingContext(ctx, "mysecretpassphrase", "abc123")

	// getCodenamesStreamingExample(ctx, client)
	// getSingleCodenameAndExitExample(ctx, client, "Science")

	for {
		getHTTPHello()
		getSingleCodenameAndExitExample(ctx, client, "Science")
		time.Sleep(5 * time.Second)
	}
}

func getHTTPHello() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if resp, err := http.Get(addressHTTP); err == nil {
		response, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Hello result: %s", response)
	} else {
		log.Printf("Could not get result, %v", err)
	}
}

func getSingleCodenameAndExitExample(ctx context.Context, client creator.CodenameCreatorClient, category string) {
	result, err := client.GetCodename(ctx, &creator.NameRequest{Category: category})
	if err != nil {
		log.Printf("Could not get result, %v", err)
	}

	log.Printf("Codename result: %s", result)
}

func getCodenamesStreamingExample(ctx context.Context, client creator.CodenameCreatorClient) {
	fmt.Println("Generating codenames...")
	stream, err := client.KeepGettingCodenames(ctx)

	if err != nil {
		log.Fatalf("Could not get stream, %v", err)
	}

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatalf("%v", err)
			}

			log.Printf("Server: %s, Received: %s\n", in.Server, in.Name)
		}
	}()

	category := "Science"
	for {
		if category == "Science" {
			category = "Animals"
		} else {
			category = "Science"
		}

		err := stream.Send(&creator.NameRequest{Category: category})
		if err != nil {
			log.Fatalf("%v", err)
		}
		time.Sleep(10 * time.Second)
	}
}
