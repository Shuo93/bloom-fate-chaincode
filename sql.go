package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
)


func invokeBySQL(stub shim.ChaincodeStubInterface, sqlStr string) error {
	logger.Infof("execute sql: %s", sqlStr)
	if err := stub.PutStateBySql(sqlStr); err != nil {
		logger.Error(err.Error())
		return err
	}
	return nil
}

func deleteBySQL(stub shim.ChaincodeStubInterface, sqlStr string) error {
	logger.Infof("execute sql: %s", sqlStr)
	if err := stub.DeleteStateBySql(sqlStr); err != nil {
		logger.Error(err.Error())
		return err
	}
	return nil
}

func queryBySQL(stub shim.ChaincodeStubInterface, sqlStr string) ([][]string, error) {
	logger.Infof("execute sql: %s", sqlStr)
	sqlResult, err := stub.GetStateBySql(sqlStr)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	logger.Infof("execute sql success")
	var jsonParsed [][]string
	json.Unmarshal(sqlResult, &jsonParsed)
	return jsonParsed, nil
}