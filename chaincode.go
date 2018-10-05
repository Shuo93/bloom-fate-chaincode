package main

import (
	"crypto/sha256"
	"encoding/json"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("BloomFate")

func register(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		username string
		password string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	hashBytes := sha256.Sum256([]byte(m.username))
	userID := string(hashBytes[:])
	hashBytes = sha256.Sum256([]byte(m.password))
	password := string(hashBytes[:])
	initValue := "50"
	createdTime := time.Now().Format("20060102150405")
	sqlStr := "insert into account (user_id, user_name, password, credit_value, created_time) " +
		"values (" + userID + ", " + m.username + ", " + password +
		", " + initValue + ", " + createdTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func login(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		username string
		password string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	hashBytes := sha256.Sum256([]byte(m.username))
	userID := string(hashBytes[:])
	hashBytes = sha256.Sum256([]byte(m.password))
	password := string(hashBytes[:])
	sqlStr := "select count(user_id) from account where name_id = " +
		userID + " and password = " + password
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	if len(sqlResult) < 2 {
		return shim.Error("Error, no info found")
	}
	num, err := strconv.Atoi(sqlResult[1][0])
	if err != nil {
		return shim.Error(err.Error())
	}
	if num == 0 {
		return shim.Error("Error, no account found")
	}
	if num != 1 {
		return shim.Error("Error, more than one account registered")
	}
	return shim.Success(nil)
}

func invokeBySQL(stub shim.ChaincodeStubInterface, sqlStr string) error {
	logger.Infof("execute sql: %s" + sqlStr)
	// if err := stub.PutStateBySql(sqlStr); err != nil {
	// 	logger.Errorf("execute sql error occur: " + err)
	// 	return err
	// }
	logger.Infof("execute sql success")
	return nil
}

func queryBySQL(stub shim.ChaincodeStubInterface, sqlStr string) ([][]string, error) {
	logger.Infof("execute sql: %s" + sqlStr)
	// sqlResult, err := stub.GetStateBySql(sqlStr)
	// if err != nil {
	// 	logger.Errorf("execute sql error occur: " + err)
	// 	return nil, err
	// }
	logger.Infof("execute sql success")
	var jsonParsed [][]string
	// json.Unmarshal(sqlResult, &jsonParsed)
	return jsonParsed, nil
}
