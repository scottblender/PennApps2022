package main

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"

	"github.com/sonr-io/sonr/pkg/crypto/mpc"
	"github.com/sonr-io/sonr/pkg/motor"
	"github.com/sonr-io/sonr/third_party/types/common"
	mt "github.com/sonr-io/sonr/third_party/types/motor"
	bt "github.com/sonr-io/sonr/x/bucket/types"
	"github.com/sonr-io/sonr/x/schema/types"
)

const (
	APP_ACCT_ADDR   = "snr1hy7kqutdw8kwr8635vr6fjn2n8h47rff5ce0nt"
	APP_AES_KEY_LOC = "C:/Users/scott/Desktop/PennApps_2022/aes"
	password        = "Y0urP@ssWd123"
)

var (
	aes_key = make([]byte, 0)
)

func main() {
	mtr, _ := motor.EmptyMotor(&mt.InitializeRequest{
		DeviceId: "unique_device_id", // this field must be unique to the device
	}, common.DefaultCallback())
	fmt.Println(mtr.GetDeviceID())
	aesKey, err := mpc.NewAesKey()
	aes_key = aesKey
	if err != nil {
		fmt.Printf("Error encountered.%s", err)
	}
	fmt.Println(string(aesKey))
	// Store the AES key generated on disk using the `ioutil` go pkg
	req := mt.CreateAccountRequest{
		Password:  password,
		AesDscKey: aesKey,
	}
	if err != nil {
		fmt.Printf("Error encountered. %s", err)
	}
	mtr.CreateAccount(req)
	address := mtr.Address
	LogIn(mtr, address)
	response1, response2 := CreateSchemafunction(mtr)
	patient_object := CreatePatient(mtr, response1.WhatIs.Did)
	doctor_object := CreateDoctor(mtr, response2.WhatIs.Did)
	fmt.Print(patient_object.Reference.Cid)
	fmt.Print(doctor_object.Reference.Cid)
}

func LogIn(mtr motor.MotorNode, address string) {
	fmt.Println(len(aes_key))
	_, err := mtr.Login(mt.LoginRequest{Did: address, Password: password, AesPskKey: aes_key})
	if err != nil {
		fmt.Printf("Error encountered. %s", err)
	}
	fmt.Println(mtr.GetAddress())
}

func CreateSchemafunction(mtr motor.MotorNode) (mt.CreateSchemaResponse, mt.CreateSchemaResponse) {
	patient_schema := mt.CreateSchemaRequest{
		Label: "patient-schema",
		Fields: map[string]types.SchemaKind{
			"patient_name":          types.SchemaKind_STRING,
			"date_of_birth":         types.SchemaKind_STRING,
			"medical_issue":         types.SchemaKind_STRING,
			"primary_care_provider": types.SchemaKind_STRING,
		},
		Metadata: map[string]string{},
	}

	doctor_schema := mt.CreateSchemaRequest{
		Label: "doctor-schema",
		Fields: map[string]types.SchemaKind{
			"name":                types.SchemaKind_STRING,
			"healthcare_facility": types.SchemaKind_STRING,
			"patient_list":        types.SchemaKind_LIST,
		},
		Metadata: map[string]string{},
	}
	response, err := mtr.CreateSchema(patient_schema)
	if err != nil {
		fmt.Printf("Error encountered. %s", err)
	}
	response2, err2 := mtr.CreateSchema(doctor_schema)
	if err2 != nil {
		fmt.Printf("Error encountered. %s", err)
	}
	return response, response2
}
func CreatePatient(mtr motor.MotorNode, schemaDid string) mt.UploadObjectResponse {
	patient_builder, err := mtr.NewObjectBuilder(schemaDid)
	if err != nil {
		fmt.Printf("Error encountered. %s", err)
	}
	patient_builder.Set("patient_name", "Bob")
	patient_builder.Set("date_of_birth", "04/05/1990")
	patient_builder.Set("medical_issue", "COVID-19")
	patient_builder.Set("primary_care_provider", "Dr. Smith")
	response, err := patient_builder.Upload()
	if err != nil {
		fmt.Printf("Error encountered. %s", err)
	}
	return *response
}
func CreateDoctor(mtr motor.MotorNode, schemaDid string) mt.UploadObjectResponse {
	doctor_builder, err := mtr.NewObjectBuilder(schemaDid)
	if err != nil {
		fmt.Printf("Error encountered. %s", err)
	}
	doctor_builder.Set("name", "Dr. Smith")
	doctor_builder.Set("healthcare_facility", "UPenn")
	patient_list := list.New()
	patient_list.PushFront("Bob")
	patient_list.PushFront("John")
	doctor_builder.Set("patient_list", patient_list)
	response, err2 := doctor_builder.Upload()
	if err2 != nil {
		fmt.Printf("Error encountered. %s", err)
	}
	return *response
}
func CreateBuckets(mtr motor.MotorNode, patient_id string, doctor_id string, patient_schemaDid string, doctor_schemaDid string) {
	req := mt.CreateBucketRequest{
		Creator:    mtr.GetAddress(),
		Label:      "patient_bucket",
		Visibility: bt.BucketVisibility_PUBLIC,
		Role:       bt.BucketRole_USER,
		Content:    make([]*bt.BucketItem, 0),
	}
	req.Content = append(req.Content, &bt.BucketItem{
		Name:      "patient_information",
		Uri:       patient_id,
		Type:      bt.ResourceIdentifier_CID,
		SchemaDid: patient_schemaDid,
	})
	req.Content = append(req.Content, &bt.BucketItem{
		Name:      "doctor_information",
		Uri:       doctor_id,
		Type:      bt.ResourceIdentifier_CID,
		SchemaDid: doctor_schemaDid,
	})
	site_bucket, err := mtr.CreateBucket(context.Background(), req)
	if err != nil {
		fmt.Printf("Error encountered. %s", err)
	}
	for _, c := range site_bucket.GetContent() {
		content := make(map[string]interface{})
		fmt.Println(json.Unmarshal(c.Item, &content))
	}
}
