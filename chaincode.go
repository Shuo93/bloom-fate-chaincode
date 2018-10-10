package main

import (
	"encoding/hex"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("BloomFate")

const timestampFormat = "20060102150405"

func (cc *BloomFateChaincode) register(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		Username  string
		Password  string
		PublicKey string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	hashBytes := sha256.Sum256([]byte(m.Username))
	userID := hex.EncodeToString(hashBytes[:])
	hashBytes = sha256.Sum256([]byte(m.Password))
	password := hex.EncodeToString(hashBytes[:])
	initValue := "50"
	createdTime := time.Now().Format("20060102150405")
	sqlStr := "insert into account (user_id, user_name, password, credit_value, public_key, created_time) " +
		"values (" + userID + ", " + m.Username + ", " + password +
		", " + initValue + ", " + m.PublicKey + ", " + createdTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (cc *BloomFateChaincode) login(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		Username string
		Password string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	hashBytes := sha256.Sum256([]byte(m.Username))
	userID := hex.EncodeToString(hashBytes[:])
	hashBytes = sha256.Sum256([]byte(m.Password))
	password := hex.EncodeToString(hashBytes[:])
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
	UserID      string
	Name        string
	Age         string
	Sex         string
	Location    string
	PhotoHash   string
	PhotoFormat string
	Phone       string
	Email       string
}

type educationMessage struct {
	Degree       string
	School       string
	EncryptedKey string
	Signature    string
}

type occupationMessage struct {
	Company      string
	Job          string
	Salary       string
	EncryptedKey string
	Signature    string
}

func (cc *BloomFateChaincode) uploadPersonalInfo(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		Basic      basicMessage
		Education  educationMessage
		Occupation occupationMessage
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	modifiedTime := time.Now().Format("20060102150405")
	sqlStr := "insert into user_basic (user_id, name, age, sex, location, photo_hash, photo_format, " +
		"phone, email, modified_time) values (" + m.Basic.UserID + ", " + m.Basic.Name + ", " +
		m.Basic.Age + ", " + m.Basic.Sex + ", " + m.Basic.PhotoHash + ", " +
		m.Basic.PhotoFormat + ", " + m.Basic.Phone + ", " + m.Basic.Email + ", " + modifiedTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}

	sqlStr = "insert into user_credit (user_id, general, photo, education, occupation, impression, )" +
		"other, date_num) values (" + m.Basic.UserID + ", 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (cc *BloomFateChaincode) queryPersonalInfo(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		UserID string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}

	sqlStr := "select name, age, sex, location, photo_hash, photo_format, phone, email " +
		"from user_basic where user_id = " + m.UserID
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	if len(sqlResult) < 2 {
		return shim.Error("Error, no data for the user")
	}
	if len(sqlResult) > 2 {
		return shim.Error("Error, redundant data for the user")
	}
	basic := basicMessage{
		m.UserID,
		sqlResult[1][0],
		sqlResult[1][1],
		sqlResult[1][2],
		sqlResult[1][3],
		sqlResult[1][4],
		sqlResult[1][5],
		sqlResult[1][6],
		sqlResult[1][7]}

	sqlStr = "select degree, school, encrypted_key, signature " +
		"from user_education where user_id = " + m.UserID

	sqlResult, err = queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	if len(sqlResult) < 2 {
		return shim.Error("Error, no data for the user")
	}
	if len(sqlResult) > 2 {
		return shim.Error("Error, redundant data for the user")
	}
	education := educationMessage{
		sqlResult[1][0],
		sqlResult[1][1],
		sqlResult[1][2],
		sqlResult[1][3]}

	sqlStr = "select company, job, salary, encrypted_key, signature " +
		"from user_occupation where user_id = " + m.UserID
	sqlResult, err = queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	if len(sqlResult) < 2 {
		return shim.Error("Error, no data for the user")
	}
	if len(sqlResult) > 2 {
		return shim.Error("Error, redundant data for the user")
	}
	occupation := occupationMessage{
		sqlResult[1][0],
		sqlResult[1][1],
		sqlResult[1][2],
		sqlResult[1][3],
		sqlResult[1][4]}

	type response struct {
		Basic      basicMessage
		Education  educationMessage
		Occupation occupationMessage
	}
	res := response{basic, education, occupation}
	resBytes, err := json.Marshal(res)
	if err != nil {
		return shim.Error("Error, wrong data format")
	}
	return shim.Success(resBytes)
}

func (cc *BloomFateChaincode) queryPersonList(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		AgeStart string
		AgeEnd   string
		Sex      string
		Location string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select user_id, name, age, sex, location, " +
		"photo_hash, photo_format, phone, email from user_basic"
	if m.Sex != "" {
		sqlStr += " where sex = " + m.Sex
	}
	if m.Location != "" {
		sqlStr += " and location = " + m.Location
	}
	if m.AgeStart != "" && m.AgeEnd != "" {
		sqlStr += " and age between " + m.AgeStart + " and " + m.AgeEnd
	}
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	if len(sqlResult) < 2 {
		return shim.Error("Error, no data for the user")
	}

	size := len(sqlResult) - 1
	res := make([]basicMessage, size)
	for i, r := range sqlResult[1:] {
		basic := basicMessage{
			r[0],
			r[1],
			r[2],
			r[3],
			r[4],
			r[5],
			r[6],
			r[7],
			r[8]}
		res[i] = basic
	}
	resBytes, err := json.Marshal(res)
	if err != nil {
		return shim.Error("Error, wrong data format")
	}
	return shim.Success(resBytes)
}

func (cc *BloomFateChaincode) modifyPersonalInfo(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		UserID       string
		InfoType     string
		Feild        string
		Value        string
		EncryptedKey string
		Signature    string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	modifiedTime := time.Now().Format("20060102150405")
	var sqlStr string
	if m.EncryptedKey == "" || m.Signature == "" {
		sqlStr = "insert into user_" + m.InfoType + " (user_id, " + m.Feild + ", modified_time) values (" +
			m.UserID + ", " + m.Value + ", " + modifiedTime + ")"
	} else {
		sqlStr = "insert into user_" + m.InfoType + " (user_id, " + m.Feild + ", encrypted_key, signature, modified_time) values (" +
			m.UserID + ", " + m.Value + ", " + m.EncryptedKey + ", " + m.Signature + ", " + modifiedTime + ")"
	}
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (cc *BloomFateChaincode) sendDate(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		SenderID   string
		ReceiverID string
		Location   string
		DateTime   string
		Message    string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	status := "pending"
	sendTime := time.Now().Format("20060102150405")
	sqlStr := "insert into date_list (sender_id, receiver_id, location, " +
		"date_time, message, status, send_time) values (" + m.SenderID + ", " + m.ReceiverID +
		", " + m.Location + ", " + m.DateTime + ", " + m.Message + ", " + status + ", " + sendTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (cc *BloomFateChaincode) queryDate(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		UserType string
		UserID   string
		Status   string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select sender_id, receiver_id, location, date_time, message, status, send_time, confirm_time from date_list where " + m.UserType + " = " + m.UserID
	if m.Status != "" {
		sqlStr += " and status = " + m.Status
	}
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(convertSQLResultToJSON(sqlResult))
}

func convertSQLResultToJSON(sqlResult [][]string) []byte {
	var buf bytes.Buffer
	buf.WriteString("[")
	for i := 1; i < len(sqlResult); i++ {
		if i != 1 {
			buf.WriteString(",")
		}
		buf.WriteString("{")
		for j, key := range sqlResult[0] {
			if j != 0 {
				buf.WriteString(",")
			}
			buf.WriteString("\"")
			buf.WriteString(key)
			buf.WriteString("\"")
			buf.WriteString(":")
			buf.WriteString("\"")
			buf.WriteString(sqlResult[i][j])
			buf.WriteString("\"")
		}
		buf.WriteString("}")
	}
	buf.WriteString("]")
	return buf.Bytes()
}

func (cc *BloomFateChaincode) replyDate(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		SenderID   string
		ReveiverID string
		Status     string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "insert into date_list (sender_id, receiver_id, status) values (" +
		m.SenderID + ", " + m.ReveiverID + ", " + m.Status + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (cc *BloomFateChaincode) confirmDate(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		SenderID   string
		ReveiverID string
		Status     string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select status from date_list where sender_id = " +
		m.SenderID + " and receiver_id = " + m.ReveiverID
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	status := sqlResult[1][0]
	if status == "confirmed" || status == "reject" || status == "pending" {
		return shim.Error("Error, wrong date status: " + status)
	}
	if status == "approve" {
		sqlStr = "insert into date_list (sender_id, receiver_id, status)" +
			" values (" + m.SenderID + ", " + m.ReveiverID + ", " + "confirm" + ")"
	}
	if status == "confirm" {
		confirmTime := time.Now().Format("20060102150405")
		sqlStr = "insert into date_list (sender_id, receiver_id, status, confirm_time)" +
			" values (" + m.SenderID + ", " + m.ReveiverID + ", " + "confirmed" + ", " + confirmTime + ")"
	}
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (cc *BloomFateChaincode) like(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		UserID  string
		LikerID string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	createdTime := time.Now().Format("20060102150405")
	sqlStr := "insert into like_list (user_id, liker_id, created_time) values (" +
		m.UserID + ", " + m.LikerID + ", " + createdTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (cc *BloomFateChaincode) unlike(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		UserID  string
		LikerID string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "delete from like_list where user_id = " + m.UserID + " and liker_id = " + m.LikerID
	if err := deleteBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (cc *BloomFateChaincode) queryLikeList(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		UserID string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select liker_id, created_time from like_list where user_id = " + m.UserID
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(convertSQLResultToJSON(sqlResult))
}

func (cc *BloomFateChaincode) sendPermission(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		SenderID          string
		ReceiverID        string
		PermissionType    string
		PermissionContent string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sendTime := time.Now().Format(timestampFormat)
	status := "pending"
	sqlStr := "insert into permission (send_id, receiver_id, permission_type, permission_content, " +
		"status, send_time) values (" + m.SenderID + ", " + m.ReceiverID + ", " + m.PermissionType +
		", " + m.PermissionContent + ", " + status + ", " + sendTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (cc *BloomFateChaincode) queryPermession(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		UserType string
		UserID   string
		Status   string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select sender_id, receiver_id, permission_type, permission_content," +
		" status, encrypted_key, send_time from permission where " + m.UserType + " = " + m.UserID
	if m.Status != "" {
		sqlStr += " and status = " + m.Status
	}
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(convertSQLResultToJSON(sqlResult))
}

func (cc *BloomFateChaincode) replyPermession(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		SenderID          string
		ReceiverID        string
		PermissionContent string
		Status            string
		EncryptedKey      string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	confirmTime := time.Now().Format(timestampFormat)
	sqlStr := "insert into permission (sender_id, receiver_id, permission_content," +
		" status, encrypted_key, confirm_time) values (" + m.SenderID + ", " + m.ReceiverID + ", " +
		m.PermissionContent + ", " + m.Status + ", " + m.EncryptedKey + ", " + confirmTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (cc *BloomFateChaincode) queryModifyRecord(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		UserID   string
		InfoType string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select modified_time from user_" + m.InfoType + "where user_id = " + m.UserID 
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(convertSQLResultToJSON(sqlResult))
}

func (cc *BloomFateChaincode) queryModifyRecordByTime(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		UserID       string
		InfoType     string
		ModifiedTime string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select * from user_" + m.InfoType + "_history where user_id = " +
		m.UserID + " and modified_time = " + m.ModifiedTime
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(convertSQLResultToJSON(sqlResult))
}

func (cc *BloomFateChaincode) measureCredit(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		SenderID   string
		ReceiverID string
		General    string
		Photo      string
		Education  string
		Occupation string
		Impression string
		Other      string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	createdTime := time.Now().Format(timestampFormat)
	sqlStr := "insert into date_measure (sender_id, receiver_id, general, photo, " +
		"education, occupation, impression, other, created_time) values (" + m.SenderID + ", " +
		m.ReceiverID + ", " + m.General + ", " + m.Photo + ", " + m.Education + ", " +
		m.Occupation + ", " + m.Impression + ", " + m.Other + ", " + createdTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (cc *BloomFateChaincode) queryCredit(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		UserID string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select general, photo, education, occupation, impression, other, date_num from user_credit " +
		"where user_id = " + m.UserID
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(convertSQLResultToJSON(sqlResult))
}

func exchangeCreditValue(sender string, receiver string, value string, stub shim.ChaincodeStubInterface) error {
	if err := subtractCreditValue(sender, value, stub); err != nil {
		return err
	}
	if err := addCreditValue(receiver, value, stub); err != nil {
		return err
	}
	return nil
}

func addCreditValue(userID string, valueStr string, stub shim.ChaincodeStubInterface) error {
	balance, value, err := getBalanceAndValue(userID, valueStr, stub)
	if err != nil {
		return err
	}
	balance += value
	return changeCreditValue(userID, balance, stub)
}

func subtractCreditValue(userID string, valueStr string, stub shim.ChaincodeStubInterface) error {
	balance, value, err := getBalanceAndValue(userID, valueStr, stub)
	if err != nil {
		return err
	}
	if balance < value {
		return errors.New("value have to be smaller than balance")
	}
	balance -= value
	return changeCreditValue(userID, balance, stub)
}

func changeCreditValue(userID string, balance float64, stub shim.ChaincodeStubInterface) error {
	sqlStr := "insert into account (user_id, credit_value) values (" + userID +
		", " + fmt.Sprintf("%.1f", balance) + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return err
	}
	return nil
}

func getBalanceAndValue(userID string, valueStr string, stub shim.ChaincodeStubInterface) (float64, float64, error) {
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, 0, err
	}
	sqlStr := "select credit_value from account where user_id = " + userID
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return 0, value, err
	}
	balance, err := strconv.ParseFloat(sqlResult[1][0], 64)
	if err != nil {
		return 0, value, err
	}
	return balance, value, err
}

func invokeBySQL(stub shim.ChaincodeStubInterface, sqlStr string) error {
	logger.Infof("execute sql: %s" + sqlStr)
	return stub.PutStateBySql(sqlStr)
}

func deleteBySQL(stub shim.ChaincodeStubInterface, sqlStr string) error {
	logger.Infof("execute sql: %s" + sqlStr)
	return stub.DeleteStateBySql(sqlStr)
}

func queryBySQL(stub shim.ChaincodeStubInterface, sqlStr string) ([][]string, error) {
	logger.Infof("execute sql: %s" + sqlStr)
	sqlResult, err := stub.GetStateBySql(sqlStr)
	if err != nil {
		return nil, err
	}
	logger.Infof("execute sql success")
	var jsonParsed [][]string
	json.Unmarshal(sqlResult, &jsonParsed)
	return jsonParsed, nil
}
