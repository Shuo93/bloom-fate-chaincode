package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func register(stub shim.ChaincodeStubInterface, args string) pb.Response {
	fmt.Printf("Init method is called.")
	return shim.Success(nil)
}

func login(stub shim.ChaincodeStubInterface, arg string) pb.Response {
	return shim.Success(nil)
}
