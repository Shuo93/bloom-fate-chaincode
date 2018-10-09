package main

import (
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

type chaincodeImpl struct{}

func (c *chaincodeImpl) register(stub shim.ChaincodeStubInterface, args string) pb.Response {
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

func (c *chaincodeImpl) login(stub shim.ChaincodeStubInterface, args string) pb.Response {
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
	age         string
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
	salary       string
	encryptedKey string
	signature    string
}

func (c *chaincodeImpl) uploadPersonalInfo(stub shim.ChaincodeStubInterface, args string) pb.Response {
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
		m.basic.age + ", " + m.basic.sex + ", " + m.basic.photoHash + ", " +
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

func (c *chaincodeImpl) queryPersonalInfo(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		userID string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}

	sqlStr := "select name, age, sex, location, photo_hash, photo_format, phone, email " +
		"from user_basic where user_id = " + m.userID
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
		m.userID,
		sqlResult[1][0],
		sqlResult[1][1],
		sqlResult[1][2],
		sqlResult[1][3],
		sqlResult[1][4],
		sqlResult[1][5],
		sqlResult[1][6],
		sqlResult[1][7]}

	sqlStr = "select degree, school, encrypted_key, signature " +
		"from user_education where user_id = " + m.userID

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
		"from user_occupation where user_id = " + m.userID
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
		basic      basicMessage
		education  educationMessage
		occupation occupationMessage
	}
	res := response{basic, education, occupation}
	resBytes, err := json.Marshal(res)
	if err != nil {
		return shim.Error("Error, wrong data format")
	}
	return shim.Success(resBytes)
}

func (c *chaincodeImpl) queryPersonList(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		ageStart string
		ageEnd   string
		sex      string
		location string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select user_id, name, age, sex, location, " +
		"photo_hash, photo_format, phone, email from user_basic"
	if m.sex != "" {
		sqlStr += " where sex = " + m.sex
	}
	if m.location != "" {
		sqlStr += " and location = " + m.location
	}
	if m.ageStart != "" && m.ageEnd != "" {
		sqlStr += " and age between " + m.ageStart + " and " + m.ageEnd
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

func (c *chaincodeImpl) modifyPersonalInfo(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		userID       string
		infoType     string
		feild        string
		value        string
		encryptedKey string
		signature    string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	modifiedTime := time.Now().Format("20060102150405")
	var sqlStr string
	if m.encryptedKey == "" || m.signature == "" {
		sqlStr = "insert into user_" + m.infoType + " (user_id, " + m.feild + ", modified_time) values (" +
			m.userID + ", " + m.value + ", " + modifiedTime + ")"
	} else {
		sqlStr = "insert into user_" + m.infoType + " (user_id, " + m.feild + ", encrypted_key, signature, modified_time) values (" +
			m.userID + ", " + m.value + ", " + m.encryptedKey + ", " + m.signature + ", " + modifiedTime + ")"
	}
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (c *chaincodeImpl) sendDate(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		senderID   string
		receiverID string
		location   string
		dateTime   string
		message    string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	status := "pending"
	sendTime := time.Now().Format("20060102150405")
	sqlStr := "insert into date_list (sender_id, receiver_id, location, " +
		"date_time, message, status, send_time) values (" + m.senderID + ", " + m.receiverID +
		", " + m.location + ", " + m.dateTime + ", " + m.message + ", " + status + ", " + sendTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (c *chaincodeImpl) queryDate(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		userType string
		userID   string
		status   string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select sender_id, receiver_id, location, date_time, message, status, send_time, confirm_time from date_list where " + m.userType + " = " + m.userID
	if m.status != "" {
		sqlStr += " and status = " + m.status
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

func (c *chaincodeImpl) replyDate(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		senderID   string
		reveiverID string
		status     string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "insert into date_list (sender_id, receiver_id, status) values (" +
		m.senderID + ", " + m.reveiverID + ", " + m.status + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (c *chaincodeImpl) confirmDate(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		senderID   string
		reveiverID string
		status     string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select status from date_list where sender_id = " +
		m.senderID + " and receiver_id = " + m.reveiverID
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
			" values (" + m.senderID + ", " + m.reveiverID + ", " + "confirm" + ")"
	}
	if status == "confirm" {
		confirmTime := time.Now().Format("20060102150405")
		sqlStr = "insert into date_list (sender_id, receiver_id, status, confirm_time)" +
			" values (" + m.senderID + ", " + m.reveiverID + ", " + "confirmed" + ", " + confirmTime + ")"
	}
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (c *chaincodeImpl) like(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		userID  string
		likerID string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	createdTime := time.Now().Format("20060102150405")
	sqlStr := "insert into like_list (user_id, liker_id, created_time) values (" +
		m.userID + ", " + m.likerID + ", " + createdTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (c *chaincodeImpl) unlike(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		userID  string
		likerID string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "delete from like_list where user_id = " + m.userID + " and liker_id = " + m.likerID
	if err := deleteBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (c *chaincodeImpl) queryLikeList(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		userID string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select liker_id, created_time from like_list where user_id = " + m.userID
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(convertSQLResultToJSON(sqlResult))
}

func (c *chaincodeImpl) sendPermission(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		senderID          string
		receiverID        string
		permissionType    string
		permissionContent string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sendTime := time.Now().Format(timestampFormat)
	status := "pending"
	sqlStr := "insert into permission (send_id, receiver_id, permission_type, permission_content, " +
		"status, send_time) values (" + m.senderID + ", " + m.receiverID + ", " + m.permissionType +
		", " + m.permissionContent + ", " + status + ", " + sendTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (c *chaincodeImpl) queryPermession(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		userType string
		userID   string
		status   string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select sender_id, receiver_id, permission_type, permission_content," +
		" status, encrypted_key, send_time from permission where " + m.userType + " = " + m.userID
	if m.status != "" {
		sqlStr += " and status = " + m.status
	}
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(convertSQLResultToJSON(sqlResult))
}

func (c *chaincodeImpl) replyPermession(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		senderID          string
		receiverID        string
		permissionContent string
		status            string
		encryptedKey      string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	confirmTime := time.Now().Format(timestampFormat)
	sqlStr := "insert into permission (sender_id, receiver_id, permission_content," +
		" status, encrypted_key, confirm_time) values (" + m.senderID + ", " + m.receiverID + ", " +
		m.permissionContent + ", " + m.status + ", " + m.encryptedKey + ", " + confirmTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (c *chaincodeImpl) queryModifyRecord(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		userID   string
		infoType string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select modified_time from user_" + m.infoType
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(convertSQLResultToJSON(sqlResult))
}

func (c *chaincodeImpl) queryModifyRecordByTime(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		userID       string
		infoType     string
		modifiedTime string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select * from user_" + m.infoType + "_history where user_id = " +
		m.userID + " and modified_time = " + m.modifiedTime
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(convertSQLResultToJSON(sqlResult))
}

func (c *chaincodeImpl) measureCredit(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		senderID   string
		receiverID string
		general    string
		photo      string
		education  string
		occupation string
		impression string
		other      string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	createdTime := time.Now().Format(timestampFormat)
	sqlStr := "insert into date_measure (sender_id, receiver_id, general, photo, " +
		"education, occupation, impression, other, created_time) values (" + m.senderID + ", " +
		m.receiverID + ", " + m.general + ", " + m.photo + ", " + m.education + ", " +
		m.occupation + ", " + m.impression + ", " + m.other + ", " + createdTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (c *chaincodeImpl) queryCredit(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		userID string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select general, photo, education, occupation, impression, other, date_num from user_credit " +
		"where user_id = " + m.userID
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
	// if err := stub.PutStateBySql(sqlStr); err != nil {
	// 	logger.Errorf("execute sql error occur: " + err)
	// 	return err
	// }
	logger.Infof("execute sql success")
	return nil
}

func deleteBySQL(stub shim.ChaincodeStubInterface, sqlStr string) error {
	logger.Infof("execute sql: %s" + sqlStr)
	// if err := stub.DeleteStateBySql(sqlStr); err != nil {
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
