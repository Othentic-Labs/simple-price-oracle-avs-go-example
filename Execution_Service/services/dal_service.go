package services

import (
	"Execution_Service/config"
	"Execution_Service/utils"
	"encoding/binary"
	"encoding/hex"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
)

func Init() {
	config.Init()
}

type Params struct {
	proofOfTask      string
	data             string
	taskDefinitionId int
	performerAddress string
	signature        string
	signatureType    string
}

func SendTask(proofOfTask string, data string, taskDefinitionId int) {

	arguments := abi.Arguments{
		{Type: abi.Type{T: abi.StringTy}},
		{Type: abi.Type{T: abi.BytesTy}},
		{Type: abi.Type{T: abi.AddressTy}},
		{Type: abi.Type{T: abi.UintTy}},
	}

	dataPacked, err := arguments.Pack(
		proofOfTask,
		[]byte(data),
		common.HexToAddress(config.PerformerAddress),
		big.NewInt(int64(taskDefinitionId)),
	)
	if err != nil {
		log.Println("error occured while encoding")
		log.Println(err)
	}
	messageHash := crypto.Keccak256Hash(dataPacked).String()
	log.Println("Private key from config:", config.PrivateKey)
	log.Println("Private key length:", len(config.PrivateKey))
	signingKey, err := utils.GetSigningKey(config.PrivateKey)
	if err != nil {
		log.Println("Error getting signing key:", err)
		return
	}
	serializedSignature, err := utils.Sign(&signingKey, messageHash)
	if err != nil {
		log.Println("error occured while signing")
		log.Println(err)
	}
	log.Println("serializedSignature")
	keyBytes := signingKey.Bytes()
	log.Println("Signing key as hex:", hex.EncodeToString(keyBytes[:]))
	bigIntValue := new(big.Int)
	signingKey.BigInt(bigIntValue)
	log.Println("Signing key as big.Int:", bigIntValue)
	log.Println("Signing key as bytes:", keyBytes)

	// Print as uint32 array for comparison with JS
	log.Println("Signing key as uint32 array:")
	for i := 0; i < 8; i++ {
		u := binary.LittleEndian.Uint32(keyBytes[i*4 : (i+1)*4])
		log.Printf("  %d", u)
	}

	log.Println("Serialized signature:", serializedSignature)

	client, err := rpc.Dial(config.OTHENTIC_CLIENT_RPC_ADDRESS)
	if err != nil {
		log.Println(err)
	}

	params := Params{
		proofOfTask:      proofOfTask,
		data:             "0x" + hex.EncodeToString([]byte(data)),
		taskDefinitionId: taskDefinitionId,
		performerAddress: config.PerformerAddress,
		signature:        serializedSignature,
		signatureType:    "bls",
	}

	response := makeRPCRequest(client, params)
	log.Println("API response:", response)
}

func makeRPCRequest(client *rpc.Client, params Params) interface{} {
	var result interface{}

	err := client.Call(&result, "sendTask", params.proofOfTask, params.data, params.taskDefinitionId, params.performerAddress, params.signature, params.signatureType)
	if err != nil {
		log.Println(err)
	}
	return result
}
