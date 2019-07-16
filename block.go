package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"fmt"
	"encoding/hex"
			"strconv"
	"time"
)

type Block struct{
	Version int32
	PrevBlockHash []byte
	Merkleroot []byte
	Hash []byte
	Time int32
	Bits int32
	Nonce int32
	Transactions []*transaction
	Height int32
}

func(b *Block) Serialize() []byte{
	var encoded bytes.Buffer
	enc:=gob.NewEncoder(&encoded)
	err:= enc.Encode(b)
	if err!= nil{
		log.Panic(err)
	}
	return encoded.Bytes()
}

//反序列化
func deserializeblock(d []byte) *Block{
	var block Block

	decode :=gob.NewDecoder(bytes.NewReader(d))
	err := decode.Decode(&block)
	if err!=nil{
		log.Panic(err)
	}
	return &block
}

//计算困难度
func CalculateTargetFast(bits []byte) []byte {

	var result []byte
	//第一个字节  计算指数
	exponent := bits[:1]
	fmt.Printf("%x\n", exponent)

	//计算后面3个系数
	coeffient := bits[1:]
	fmt.Printf("%x\n", coeffient)

	//将字节，他的16进制为"18"  转化为了string "18"
	str := hex.EncodeToString(exponent) //"18"
	fmt.Printf("str=%s\n", str)
	//将字符串18转化为了10进制int64 24
	exp, _ := strconv.ParseInt(str, 16, 8)

	fmt.Printf("exp=%d\n", exp)
	//拼接，计算出目标hash
	result = append(bytes.Repeat([]byte{0x00}, 32-int(exp)), coeffient...)
	result = append(result, bytes.Repeat([]byte{0x00}, 32-len(result))...)

	return result
}

func (b *Block) createmerkeltreeroot(transation []*transaction){
	var transhash  [][]byte
	for _,tx := range  transation{
		transhash = append(transhash,tx.hash())
	}
	mtree:= NewMerkleTree(transhash)

	b.Merkleroot = mtree.RootNode.Data
}
/*
func testcreatemerkletreeroot(){
	block := &Block{
		2,
		[]byte{},
		[]byte{},
		[]byte{},
		int32(time.Now().Unix()),
		404454260,
		0,
		[]*transaction{},
	}

	txin:=txinput{[]byte{}, -1 , nil}
	txout:= newtxoutput(subsidy,"first")
	tx := transaction{nil,[]txinput{txin},[]txoutput{*txout}}



	txin2:=txinput{[]byte{}, -1 , nil}
	txout2:= newtxoutput(subsidy,"first")
	tx2 := transaction{nil,[]txinput{txin2},[]txoutput{*txout2}}

	var transactionhash []*transaction
	transactionhash = append(transactionhash,&tx,&tx2)

	block.createmerkeltreeroot(transactionhash)

	fmt.Printf("%x\n", block.Merkleroot)
}
*/
func (b*Block) bSerialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(b)
	if err != nil {

		fmt.Println("--------\n")
		log.Panic(err)
	}
	fmt.Printf("%x\n", encoded.Bytes())
	return encoded.Bytes()
}
func (b *Block) debserialize() []byte{
	var encoded bytes.Buffer
	enc := gob.NewDecoder(&encoded)
	err := enc.Decode(b)
	if err == nil {

		fmt.Println("--------\n")
		log.Panic(err)
	}
	fmt.Printf("%x\n", encoded.Bytes())
	return encoded.Bytes()
}

func (b*Block) String(){
	fmt.Printf("version:%s\n",strconv.FormatInt(int64(b.Version),10))
	fmt.Printf("prev blockhash:%x\n",b.PrevBlockHash)
	fmt.Printf("merkleroot hash:%x\n",b.Merkleroot )
	fmt.Printf("hash:%x\n",b.Hash)
	fmt.Printf("time:%s\n",strconv.FormatInt(int64(b.Time),10))
	fmt.Printf("bits:%s\n",strconv.FormatInt(int64(b.Bits),10))
	fmt.Printf("nonce:%s\n\n",strconv.FormatInt(int64(b.Nonce),10))
	//fmt.Println(b.Transactions)
}
func testserialize(){
	block := &Block{
		2,
		[]byte{},
		[]byte{},
		[]byte{},
		int32(time.Now().Unix()),
		404454260,
		0,
		[]*transaction{},
		0,
	}
	deblock:= deserializeblock(block.bSerialize())
	deblock.String()

}

func newgenisblock(ptransaction []* transaction) *Block{
	block := &Block{
		2,
		[]byte{},
		[]byte{},
		[]byte{},
		int32(time.Now().Unix()),
		404454260,
		0,
		ptransaction,
		0,
	}
	pow:=newproofofwork(block)
	nonce,hash:= pow.run()
	block.Nonce = nonce
	block.Hash = hash
	block.String()
	return block
}
func testpow(){
	block := &Block{
		2,
		[]byte{},
		[]byte{},
		[]byte{},
		int32(time.Now().Unix()),
		404454260,
		0,
		[]*transaction{},
		0,
	}
	pow:=newproofofwork(block)
	nonce,_:= pow.run()
	block.Nonce = nonce
	fmt.Println(pow.walidate())
}

func newblock(ptransaction []*transaction,preblockhash []byte,height int32) *Block {
	block := &Block{
		2,
		preblockhash,
		[]byte{},
		[]byte{},
		int32(time.Now().Unix()),
		404454260,
		0,
		ptransaction,
		height,
	}
	pow:=newproofofwork(block)
	nonce,hash:= pow.run()
	block.Nonce = nonce
	block.Hash = hash
	block.String()
	return block
}
