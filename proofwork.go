package main

import (
	"math/big"
		"bytes"
	"crypto/sha256"
	"math"
)


var (
	maxnonce int32 = math.MaxInt32
)

type proofofwork struct {
	block *Block
	target *big.Int
}
const targetBits = 16
func newproofofwork(b *Block) *proofofwork  {
	target:=big.NewInt(1)
	target.Lsh(target,uint(256-targetBits))
	pow:=&proofofwork{b,target}
	return pow
}

func (pow *proofofwork) preparedata(nonce int32)  []byte{
	data:= bytes.Join(
		[][]byte{
			IntToHex(pow.block.Version),
			pow.block.PrevBlockHash,
			pow.block.Merkleroot,
			IntToHex(pow.block.Time),
			IntToHex(pow.block.Bits),
			IntToHex(nonce)},[]byte{},
		)
	return data
}

func (pow *proofofwork) run() (int32,[]byte)  {
	var nonce int32
	nonce =0
	var currenthash big.Int
	var secondhash [32]byte
	for nonce < maxnonce{
		data:=pow.preparedata(nonce)
		firsthash:=sha256.Sum256(data)
		secondhash = sha256.Sum256(firsthash[:])
		// fmt.Printf("%x\n", secondhash)
		currenthash.SetBytes(secondhash[:])
		if(currenthash.Cmp(pow.target) ==-1){
			break
		}else{
			nonce++
		}
	}
	return nonce,secondhash[:]
}

func (pow *proofofwork) walidate() bool{
	var hashInt  big.Int
	data:=pow.preparedata(pow.block.Nonce)

	fisthash:= sha256.Sum256(data)
	secondhash:= sha256.Sum256(fisthash[:])

	hashInt.SetBytes(secondhash[:])
	isvalid:= hashInt.Cmp(pow.target) ==-1
	return  isvalid
}