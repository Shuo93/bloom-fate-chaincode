package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
	logger.Infof("JSON: %s", args)
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
	logger.Infof("user_name = %s, password = %s", m.Username, m.Password)
	hashBytes := sha256.Sum256([]byte(m.Username))
	userID := hex.EncodeToString(hashBytes[:])
	hashBytes = sha256.Sum256([]byte(m.Password))
	password := hex.EncodeToString(hashBytes[:])
	logger.Infof("user_id = %s, password_hash = %s", userID, password)
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
	sqlStr := "select count(user_id) from account where user_id = '" +
		userID + "' and password = '" + password + "'"
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
	addCreditValue(userID, "1", stub)
	return shim.Success([]byte(userID))
}

type basicMessage struct {
	UserID       string `json:"userId"`
	Name         string `json:"name"`
	Age          string `json:"age"`
	Sex          string `json:"sex"`
	Location     string `json:"location"`
	PhotoHash    string `json:"photoHash"`
	PhotoFormat  string `json:"photoFormat"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Introduction string `json:"introduction"`
}

type educationMessage struct {
	Degree       string `json:"degree"`
	School       string `json:"school"`
	EncryptedKey string `json:"encryptedKey"`
	Signature    string `json:"signature"`
}

type occupationMessage struct {
	Company      string `json:"company"`
	Job          string `json:"job"`
	Salary       string `json:"salary"`
	EncryptedKey string `json:"encryptedKey"`
	Signature    string `json:"signature"`
}

func (cc *BloomFateChaincode) uploadPersonalInfo(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Infof("JSON: %s", args)
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
		"phone, email, introduction, modified_time) values (" + m.Basic.UserID + ", " + m.Basic.Name + ", " +
		m.Basic.Age + ", " + m.Basic.Sex + ", " + m.Basic.Location + ", " + m.Basic.PhotoHash + ", " +
		m.Basic.PhotoFormat + ", " + m.Basic.Phone + ", " + m.Basic.Email + ", " + m.Basic.Introduction + "," + modifiedTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr = "insert into user_education (user_id, degree, school, encrypted_key, signature, modified_time) values (" + m.Basic.UserID +
		", " + m.Education.Degree + ", " + m.Education.School + ", " + m.Education.EncryptedKey + ", " + m.Education.Signature + ", " +
		modifiedTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr = "insert into user_occupation (user_id, company, job, salary, encrypted_key, signature, modified_time) values (" + m.Basic.UserID + ", " +
		m.Occupation.Company + ", " + m.Occupation.Job + ", " + m.Occupation.Salary + ", " + m.Occupation.EncryptedKey + ", " + m.Occupation.Signature + ", " + modifiedTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr = "insert into user_credit (user_id, general, photo, education, occupation, impression, " +
		"other, date_num) values (" + m.Basic.UserID + ", 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0)"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	addCreditValue(m.Basic.UserID, "30", stub)
	return shim.Success(nil)
}

func (cc *BloomFateChaincode) queryPublicKey(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		UserID string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select public_key from account where user_id = '" + m.UserID + "'"
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(sqlResult[1][0]))
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

	sqlStr := "select name, age, sex, location, photo_hash, photo_format, phone, email, introduction " +
		"from user_basic where user_id = '" + m.UserID + "'"
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	if len(sqlResult) < 2 {
		return shim.Success([]byte("no data"))
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
		sqlResult[1][7],
		sqlResult[1][8]}

	sqlStr = "select degree, school, encrypted_key, signature " +
		"from user_education where user_id = '" + m.UserID + "'"

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
		"from user_occupation where user_id = '" + m.UserID + "'"
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
		Basic      basicMessage      `json:"basic"`
		Education  educationMessage  `json:"education"`
		Occupation occupationMessage `json:"occupation"`
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
		UserID   string
		AgeStart string
		AgeEnd   string
		Sex      bool
		Location bool
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select sex, location from user_basic where user_id = '" + m.UserID + "'"
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	mySex := sqlResult[1][0]
	myLocation := sqlResult[1][1]

	sqlStr = "select user_id, name, age, sex, location, " +
		"photo_hash, photo_format, phone, email, introduction from user_basic"
	if m.Sex == true {
		sqlStr += " where sex = '" + mySex + "'"
	} else {
		if mySex == "male" {
			sqlStr += " where sex = 'female'"
		} else {
			sqlStr += " where sex = 'male'"
		}
	}
	if m.Location == true {
		sqlStr += " and location = '" + myLocation + "'"
	}
	if m.AgeStart != "" && m.AgeEnd != "" {
		sqlStr += " and age between '" + m.AgeStart + "' and '" + m.AgeEnd + "'"
	}
	sqlResult, err = queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
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
			r[8],
			r[9]}
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
	// TODO:
	sqlStr := "select user_name from account where user_id = '" + m.SenderID + "'"
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	senderName := sqlResult[1][0]
	sqlStr = "select user_name from account where user_id = '" + m.ReceiverID + "'"
	sqlResult, err = queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	receiverName := sqlResult[1][0]
	sqlStr = "insert into date_list (sendername, receivername, sender_id, receiver_id, location, " +
		"date_time, message, status, send_time) values (" + senderName + ", " + receiverName + ", " + m.SenderID + ", " +
		m.ReceiverID + ", " + m.Location + ", " + m.DateTime + ", " + m.Message + ", " + status + ", " + sendTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	subtractCreditValue(m.SenderID, "5", stub)
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
	sqlStr := "select sendername, receivername, sender_id, receiver_id, location, date_time, message, status, send_time, confirm_time from date_list where " + m.UserType + " = '" + m.UserID + "'"
	if m.Status != "" {
		sqlStr += " and status = '" + m.Status + "'"
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

// Before date. Receiver reply date request with status (pending to approve or reject).
func (cc *BloomFateChaincode) replyDate(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		SenderID   string
		ReceiverID string
		Status     string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "insert into date_list (sender_id, receiver_id, status) values (" +
		m.SenderID + ", " + m.ReceiverID + ", " + m.Status + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	if m.Status == "approve" {
		subtractCreditValue(m.ReceiverID, "5", stub)
	} else if m.Status == "reject" {
		addCreditValue(m.SenderID, "5", stub)
	}
	return shim.Success(nil)
}

// Both sender and receiver do confirmation after date.
//The first :approve to confirm;
//The second: confirm to confirmed.
func (cc *BloomFateChaincode) confirmDate(stub shim.ChaincodeStubInterface, args string) pb.Response {
	type message struct {
		SenderID   string
		ReceiverID string
		Status     string
	}
	b := []byte(args)
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return shim.Error(err.Error())
	}
	sqlStr := "select status from date_list where sender_id = '" +
		m.SenderID + "' and receiver_id = '" + m.ReceiverID + "'"
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
			" values (" + m.SenderID + ", " + m.ReceiverID + ", confirm)"
	}
	if status == "confirm" {
		confirmTime := time.Now().Format("20060102150405")
		sqlStr = "insert into date_list (sender_id, receiver_id, status, confirm_time)" +
			" values (" + m.SenderID + ", " + m.ReceiverID + ", confirmed, " + confirmTime + ")"
		addCreditValue(m.SenderID, "5", stub)
		addCreditValue(m.ReceiverID, "5", stub)
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
	// TODO:
	sqlStr := "select user_name from account where user_id = '" + m.UserID + "'"
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	userName := sqlResult[1][0]
	sqlStr = "select user_name from account where user_id = '" + m.LikerID + "'"
	sqlResult, err = queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	likerName := sqlResult[1][0]
	sqlStr = "insert into like_list (username, likername, user_id, liker_id, created_time) values (" +
		userName + ", " + likerName + ", " + m.UserID + ", " + m.LikerID + ", " + createdTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	addCreditValue(m.LikerID, "1", stub)
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
	sqlStr := "delete from like_list where user_id = '" + m.UserID + "' and liker_id = '" + m.LikerID + "'"
	if err := deleteBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	subtractCreditValue(m.UserID, "1", stub)
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
	sqlStr := "select username, likername, liker_id, created_time from like_list where user_id = '" + m.UserID + "'"
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
	// TODO:
	sqlStr := "select user_name from account where user_id = '" + m.SenderID + "'"
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	senderName := sqlResult[1][0]
	sqlStr = "select user_name from account where user_id = '" + m.ReceiverID + "'"
	sqlResult, err = queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	receiverName := sqlResult[1][0]
	sqlStr = "insert into permission (sendername, receivername, send_id, receiver_id, permission_type, permission_content, " +
		"status, send_time) values (" + senderName + ", " + receiverName + ", " + m.SenderID + ", " +
		m.ReceiverID + ", " + m.PermissionType + ", " + m.PermissionContent + ", " + status + ", " + sendTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	subtractCreditValue(m.SenderID, "5", stub)
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
	sqlStr := "select sendername, receivername, sender_id, receiver_id, permission_type, permission_content," +
		" status, encrypted_key, send_time from permission where " + m.UserType + " = '" + m.UserID + "'"
	if m.Status != "" {
		sqlStr += " and status = '" + m.Status + "'"
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
	if m.Status == "approve" {
		addCreditValue(m.ReceiverID, "2", stub)
	} else if m.Status == "reject" {
		subtractCreditValue(m.SenderID, "5", stub)
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
	sqlStr := "select modified_time from user_" + m.InfoType + "where user_id = '" + m.UserID + "'"
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
	sqlStr := "select * from user_" + m.InfoType + "_history where user_id = '" +
		m.UserID + "' and modified_time = '" + m.ModifiedTime + "'"
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
	// todo
	sqlStr := "select user_name from account where user_id = '" + m.SenderID + "'"
	sqlResult, err := queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	senderName := sqlResult[1][0]
	sqlStr = "select user_name from account where user_id = '" + m.ReceiverID + "'"
	sqlResult, err = queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	receiverName := sqlResult[1][0]
	sqlStr = "insert into date_measure (sendername, receivername, sender_id, receiver_id, general, photo, " +
		"education, occupation, impression, other, created_time) values (" + senderName + ", " + receiverName + ", " +
		m.SenderID + ", " + m.ReceiverID + ", " + m.General + ", " + m.Photo + ", " + m.Education + ", " +
		m.Occupation + ", " + m.Impression + ", " + m.Other + ", " + createdTime + ")"
	if err := invokeBySQL(stub, sqlStr); err != nil {
		return shim.Error(err.Error())
	}
	// TODO: calculate average values in user_credit
	sqlStr = "select general, photo, education, occupation, impression, other, date_num from user_credit where user_id = '" + m.ReceiverID + "'"
	sqlResult, err = queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	//sqlResult: 0: general, 1: photo, 2: education, 3: occupation, 4: impression, 5: other
	dateNum, _ := strconv.Atoi(sqlResult[1][6])
	values := [6]string{m.General, m.Photo, m.Education, m.Occupation, m.Impression, m.Other}
	for i, r := range sqlResult[1][:6] {
		if values[i], err = updateCredit(values[i], dateNum, r); err != nil {
			return shim.Error(err.Error())
		}
	}
	sqlStr = "insert into user_credit (user_id, general, photo, education, occupation, impression, other, date_num) values (" +
		m.ReceiverID + ", " + values[0] + ", " + values[1] + ", " + values[2] + ", " + values[3] + ", " + values[4] + ", " +
		values[5] + ", " + strconv.Itoa(dateNum+1) + ")"
	sqlResult, err = queryBySQL(stub, sqlStr)
	if err != nil {
		return shim.Error(err.Error())
	}
	addCreditValue(m.SenderID, "2", stub)
	return shim.Success(nil)
}

func updateCredit(prevValueStr string, dateNum int, valueStr string) (string, error) {
	prevValue, err := strconv.ParseFloat(prevValueStr, 64)
	if err != nil {
		return prevValueStr, err
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return prevValueStr, err
	}
	result := (prevValue*float64(dateNum) + float64(value)) / float64(dateNum+1)
	return fmt.Sprintf("%.1f", result), nil
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
		"where user_id = '" + m.UserID + "'"
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
	sqlStr := "select credit_value from account where user_id = '" + userID + "'"
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
