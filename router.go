package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// BloomFateChaincode is the chaincode of bloom fate
type BloomFateChaincode struct{}

// Init callback representing the invocation of a chaincode
func (t *BloomFateChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Printf("Init method is called.")
	return shim.Success(nil)
}

// Invoke callback representing the invocation of a chaincode
func (t *BloomFateChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "register":
		return register(stub, args[0])
	case "login":
		return login(stub, args[0])
	}
	return shim.Success(nil)
}
