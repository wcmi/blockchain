package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"bytes"
)

const version = byte(0x00)

type Wallet struct{
	PrivateKey ecdsa.PrivateKey
	Publickey []byte
}

func newkeypair() (ecdsa.PrivateKey,[]byte)  {
	curve:= elliptic.P256()
	private,err:= ecdsa.GenerateKey(curve,rand.Reader)
	if err!= nil {
		fmt.Println("error")
	}
	pubkey:= append(private.PublicKey.X.Bytes(),private.Y.Bytes()...)
	return *private, pubkey
}

func hashpubkey(pubkey []byte) []byte  {
	pubkeyhash256 := sha256.Sum256(pubkey)
	pipemd160hasher:= ripemd160.New()
	_,err := pipemd160hasher.Write(pubkeyhash256[:])
	if err != nil {
		fmt.Println("error")
	}
	publicpipe160:= pipemd160hasher.Sum(nil)
	return  publicpipe160
}

func checksum(payload []byte) []byte {
	firsthash := sha256.Sum256(payload)
	secondhash:= sha256.Sum256(firsthash[:])
	pchecksum := secondhash[:4]
	return pchecksum
}

func ValidateAddress(address []byte) bool{
	pubkeyHash := Base58Decode(address)
	actualCheckSum := pubkeyHash[len(pubkeyHash)-4:]
	publickeyHash  := pubkeyHash[1:len(pubkeyHash)-4]
	targetChecksum := checksum(append([]byte{0x00},publickeyHash...))
	return bytes.Compare(actualCheckSum,targetChecksum)==0
}


func (w Wallet) GetAddress() []byte{

	pubkeyHash:= hashpubkey(w.Publickey)
	versionPayload := append([]byte{version},pubkeyHash...)
	check:=checksum(versionPayload)
	fullPayload := append(versionPayload,check...)
	//返回地址
	address:=Base58Encode(fullPayload)
	return address
}

func newwallet() *Wallet{

	private,public:=newkeypair()

	wallet := Wallet{private,public}

	return &wallet
}