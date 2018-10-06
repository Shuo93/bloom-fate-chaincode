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
		username  string
		password  string
		publicKey string
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
	sqlStr := "insert into account (user_id, user_name, password, credit_value, public_key, created_time) " +
		"values (" + userID + ", " + m.username + ", " + password +
		", " + initValue + ", " + m.publicKey + ", " + createdTime + ")"
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
	return shim.Success([]byte(userID))
}

type basicMessage struct {
	userID      string
	name        string
	age         int64
	sex         string
	location    string
	photoHash   string
	photoFormat string
	phone       string
	email       string
}

type educationMessage struct {
	degree       string
	school       string
	encryptedKey string
	signature    string
}

type occupationMessage struct {
	company      string
	job          string
	salary       int64
	encryptedKey string
	signature    string
}

func uploadPersonalInfo(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		basic      basicMessage
		education  educationMessage
		occupation occupationMessage
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	modifiedTime := time.Now().Format("20060102150405")
	sqlStr := "insert into user_basic (user_id, name, age, sex, location, photo_hash, photo_format, " +
		"phone, email, modified_time) values (" + m.basic.userID + ", " + m.basic.name + ", " +
		strconv.FormatInt(m.basic.age, 10) + ", " + m.basic.sex + ", " + m.basic.photoHash + ", " +
		m.basic.photoFormat + ", " + m.basic.phone + ", " + m.basic.email + ", " + modifiedTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}

	sqlStr = "insert into user_credit (user_id, general, photo, education, occupation, impression, )" +
		"other, date_num) values (" + m.basic.userID + ", 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func queryPersonalInfo(stub shim.ChaincodeStubInterface, args string) pb.Response {
	
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
