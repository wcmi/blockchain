package main

import (
	"fmt"
	"net"
	"log"
	"io/ioutil"
	"bytes"
	"encoding/gob"
	"io"
)

const nodeversion = 0x00

var nodeaddress string

var blocktransaction  [][]byte
const  commonlength  = 12

type Version struct {
	Version int
	Blockheight  int32
	Addrfrom  string
}

var knownodes =[] string{"localhost:30000"}

func startserver(nodeid, mineraddress string, bc  *blockchain)  {
	nodeaddress = fmt.Sprintf("localhost:%s",nodeid)
	ln, err :=  net.Listen("tcp", nodeaddress)
	defer ln.Close()

	if nodeaddress != knownodes[0]{
		sendversion(knownodes[0],bc)
	}

	for   {
		conn, err2 := ln.Accept()
		if err2 != nil{
			log.Panic(err)
		}
		go handleconnect(conn, bc)
	}
}

func handleconnect(conn net.Conn, bc *blockchain)  {
	request , err := ioutil.ReadAll(conn)
	if err != nil{
		log.Panic(err)
	}
	command := bytestocommand(request[:commonlength])
	fmt.Println(command)
	switch command {
	case "version":
		fmt.Printf("\nstr : 获取version\n")
		handleversion(request, bc)
	case "getblocks":
		fmt.Printf("get getblocks msg\n")
		handlegetblock(request,bc)
	case "inv":
		handleinv(request,bc)
	case "getdata":
		handlegetdata(request, bc)
	case "block":
		handleblock(request,bc)
	}
}
type blocksend struct {
	Addfrom string
	Block []byte
}
func handleblock(request []byte, bc *blockchain) {
	var buff  bytes.Buffer
	var payload blocksend

	buff.Write(request[commonlength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)

	if err != nil {
		log.Panic(err)
	}

	blockdata:= payload.Block
	block :=  deserializeblock(blockdata)
	bc.addblock(block)

	if len(blocktransaction) > 0 {
		blockhash := blocktransaction[0]
		sendgetdata(payload.Addfrom,"block", blockhash)
		blocktransaction =  blocktransaction[1:]
	}else {
		set := utxoset{bc}
		set.reindex()
	}

}
func handlegetdata(request []byte, bc *blockchain) {
	var buff bytes.Buffer
	var payload  getdata

	buff.Write(request[commonlength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)

	if err != nil {
		log.Panic(err)
	}
	if payload.Type == "block"  {
		fmt.Printf("payload id :%x \n",payload.Id)
		block, err := bc.getblock([]byte(payload.Id))
		if err != nil {
			log.Panic(err)
		}

		fmt.Println("getdatablock",payload.Addfrom)
		sendblock(payload.Addfrom, &block)
	}
}
func sendblock(addr string, block *Block)  {
	fmt.Println("send block",addr)
	data:= blocksend{nodeaddress,block.Serialize()}
	payload := gobencode(data)
	request := append(commandtobytes("block"), payload...)
	senddata(addr, request)
}

func handleinv(request []byte, bc *blockchain)  {
	var buff  bytes.Buffer
	var payload inv
	buff.Write(request[commonlength:])
	dec:= gob.NewDecoder(&buff)
	err:= dec.Decode(&payload)

	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("recieve inventory %d,%s", len(payload.Items),payload.Type)

	if payload.Type ==  "block"  {
		blocktransaction = payload.Items
		blockhash:= payload.Items[0]
		sendgetdata(payload.Addfrom,"block", blockhash)
		newintransit := [][]byte{}

		for _, b := range  blocktransaction{
			if bytes.Compare(b, blockhash) != 0{
				newintransit = append(newintransit,b)
			}
		}
		blocktransaction = newintransit
	}
}

type getdata struct{
	Addfrom string
	Type string
	Id []byte
}
func sendgetdata(addr string, kind string, id []byte) {
	payload := gobencode(getdata{nodeaddress,kind, id})
	request:= append(commandtobytes("getdata"),payload...)
	senddata(addr , request)
}

func handlegetblock(request []byte, bc *blockchain) {
	var buff bytes.Buffer
	var payload getblocks
	buff.Write(request[commonlength:])
	dec := gob.NewDecoder(&buff)
	err:= dec.Decode(&payload)
	if err!= nil {
		log.Panic(err)
	}
	block:= bc.getblockhash()
	fmt.Println("sendenv", payload.Addrfrom)
	sendinv(payload.Addrfrom, "block", block)
}

type inv struct {
	Addfrom string
	Type string
	Items [][]byte
}

func sendinv(addr string, kind string, items [][]byte) {
	inventroy := inv{nodeaddress,kind,items }
	payload:=  gobencode(inventroy)
	request := append(commandtobytes("inv"),payload...)
	senddata(addr, request)
}

func (ver *Version) String(){
	fmt.Printf("Version:%d\n",ver.Version)
	fmt.Printf("blockHeight:%d\n",ver.Blockheight)
	fmt.Printf("AddrFrom:%s\n",ver.Addrfrom)
}
func nodeisknow(addr string) bool  {
	for _, node := range  knownodes{
		if node == addr{
			return true
		}
	}
	return false
}

type getblocks struct {
	Addrfrom string
}

func sendgetblock(address string)  {
	payload := gobencode(getblocks{nodeaddress})
	request:= append(commandtobytes("getblocks"),payload...)
	senddata(address,request)
}
func handleversion(request []byte, bc *blockchain)  {
	var  buff  bytes.Buffer
	var  payload Version
	buff.Write(request[commonlength:])
	dec:= gob.NewDecoder(&buff)
	err:= dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	payload.String()
	myblockheight := bc.getblockheight()
	foreignerblockheight := payload.Blockheight

	if myblockheight < foreignerblockheight{
		fmt.Printf("send get block msg\n")
		sendgetblock(payload.Addrfrom)
	}else{
		sendversion(payload.Addrfrom,bc)
	}
	if !nodeisknow(payload.Addrfrom){
		knownodes = append(knownodes, payload.Addrfrom)
	}
}

func senddata(addr string, data[] byte)  {
	con, err:= net.Dial("tcp",addr)
	if err != nil {
		fmt.Printf("%s is no available",addr)
		var updatenodes []string
		for _, node := range knownodes{
			if node != addr{
				updatenodes = append(updatenodes, node)
			}
		}
		knownodes= updatenodes
	}
	defer  con.Close()
	_, err = io.Copy(con, bytes.NewReader(data))
	if err != nil{
		log.Panic(err)
	}
}
func sendversion(addr string, bc  *blockchain)  {
	blockheight := bc.getblockheight()
	payload := gobencode(Version{nodeversion,blockheight, nodeaddress})
	request := append(commandtobytes("version"),payload...)
	senddata(addr , request)
}

func gobencode(data  interface{}) []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return  buff.Bytes()
}

func commandtobytes(command string) []byte {
	var bytes [commonlength]byte
	for i, c := range command{
		bytes[i] = byte(c)
	}
	return bytes[:]
}

func bytestocommand(bytes []byte) string  {
	var command []byte
	for _, b := range  bytes{
		if b!= 0x00 {
			command = append(command,b)
		}
	}
	return fmt.Sprintf("%s",command)
}