package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	gRPC "github.com/VictoriousAnnro/gRPCServ/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*To run server, open 2 terminals and use these commands:
go run .\server\server.go
go run .\client\client.go
- In separate terminals! And order is important!*/

// Same principle as in client. Flags allows for user specific arguments/values
var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", "5400", "Tcp server")

var server gRPC.GetTimeClient   //the server
var ServerConn *grpc.ClientConn //the server connection

func main() {
	//parse flag/arguments
	flag.Parse()

	fmt.Println("--- CLIENT APP ---")

	//log to file instead of console
	//setLog()

	//connect to server and close the connection when program closes
	fmt.Println("--- join Server ---")
	ConnectToServer()
	defer ServerConn.Close()

	//start the biding
	parseInput()
}

// connect to server
func ConnectToServer() {

	//dial options
	//the server is not using TLS, so we use insecure credentials
	//(should be fine for local testing but not in the real world)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	//use context for timeout on the connection
	timeContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel() //cancel the connection when we are done

	//dial the server to get a connection to it
	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, *serverPort)
	conn, err := grpc.DialContext(timeContext, fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}

	// makes a client from the server connection and saves the connection
	// and prints rather or not the connection was is READY
	server = gRPC.NewGetTimeClient(conn)
	ServerConn = conn
	log.Println("the connection is: ", conn.GetState().String())
}

func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Type 0 to get the current time")
	fmt.Println("--------------------")

	//Infinite loop to listen for clients input.
	for {
		fmt.Print("-> ")

		//Read input into var input and any errors into err
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input) //Trim input

		if !conReady(server) {
			log.Printf("Client %s: something was wrong with the connection to the server :(", *clientsName)
			continue
		}

		//Convert string to int64, return error if the int is larger than 32bit or not a number
		val, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			continue
		}
		GetTheTime(val)
	}
}

func GetTheTime(val int64) {
	//create amount type
	request := &gRPC.Request{
		ClientName: *clientsName,
		Value:      val, //cast from int to int32
	}

	//Make gRPC call to server with amount, and recieve acknowlegdement back.
	ack, err := server.GetTime(context.Background(), request)
	if err != nil {
		log.Printf("Client %s: no response from the server, attempting to reconnect", *clientsName)
		log.Println(err)
	}

	fmt.Print("Success, the time is: ", ack.Timestring)
}

// Function which returns a true boolean if the connection to the server is ready, and false if it's not.
func conReady(s gRPC.GetTimeClient) bool {
	return ServerConn.GetState().String() == "READY"
}
