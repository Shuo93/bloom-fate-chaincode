package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func main() {
	err := shim.Start(new(BloomFateChaincode))
	if err != nil {
		fmt.Printf("Error starting Fate chaincode: %s", err)
	}
}
