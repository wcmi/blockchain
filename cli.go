package main

import (
	"os"
	"fmt"
	"flag"
	"log"
	)

type CLI struct {
	bc *blockchain
}

func (cli *CLI) addblock()  {
	cli.bc.minblock([]*transaction{})
}

func (cli *CLI) printchain(){
	cli.bc.printblockchain()
}

func (cli *CLI)validatearges()  {
	if len(os.Args) <1{
		fmt.Println("<1")
		os.Exit(1)
	}
	fmt.Println(os.Args)
}

func (cli *CLI)useage()  {
	fmt.Println("addblock\n")
	fmt.Println("printchain\n")
}

func (cli *CLI)send(from, to string, amount int)  {
	tx:= newutxotransaction(from, to, amount,cli.bc)
	newblock := cli.bc.minblock([]*transaction{tx})

	set := utxoset{cli.bc}
	set.update(newblock)
	fmt.Println("success send\n")
}

func (cli* CLI)createwallet()  {
	wallets, _:= newwallets()
	address:= wallets.createwallet()
	wallets.savetofile()
	fmt.Printf("your wallet address %s\n",address)
}
func (cli* CLI) listaddress()  {
	wallets, err:= newwallets()
	if err != nil {
		log.Panic(err)
	}
	addresses:= wallets.getwalletsadress()
	for _,address := range  addresses{
		fmt.Println(address)
	}
}

func (cli *CLI) getblockheight()  {
	fmt.Println(cli.bc.getblockheight())
}

func (cli *CLI)run()  {
	cli.validatearges()
	addblockcmd:=flag.NewFlagSet("addblock",flag.ExitOnError)
	printchaincmd:= flag.NewFlagSet("printchain",flag.ExitOnError)

	getbalancemcd:= flag.NewFlagSet("getbalance",flag.ExitOnError)
	getbalanceaddress:= getbalancemcd.String("address","","")
	sendcmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendfrom := sendcmd.String("from","", "")
	sendto := sendcmd.String("to","","")
	sendcmdvlaue:= sendcmd.Int("amount",  0,"")

	startnodecmd := flag.NewFlagSet("startnode", flag.ExitOnError)
	startnodeminner:= startnodecmd.String("minner", "","minner address")

	createwalletcmd := flag.NewFlagSet("createwallet",flag.ExitOnError)
	listaddresscmd:= flag.NewFlagSet("listaddress",flag.ExitOnError)

	getblockheight := flag.NewFlagSet("getblockheight",flag.ExitOnError)
	switch os.Args[1] {
	case "startnode":
		err:= startnodecmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getblockheight":
		err:= getblockheight.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err:= createwalletcmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddress":
		err:= listaddresscmd.Parse(os.Args[2:])
		if err!=nil  {
			log.Panic(err)
		}
	case "send":
		err:=sendcmd.Parse(os.Args[2:])
		if err!=nil {
			log.Panic(err)
		}
	case "getbalance":
		err:=getbalancemcd.Parse(os.Args[2:])
		if err!=nil {
			log.Panic(err)
		}
	case "addblock":
		err:= addblockcmd.Parse(os.Args[2:])
		if err!=nil {
			log.Panic(err)
		}
	case "printchain":
		err:= printchaincmd.Parse(os.Args[2:])
		if err!=nil {
			log.Panic(err)
		}
	default:
		cli.useage()
		os.Exit(1)
	}
	if addblockcmd.Parsed() {
		cli.addblock()
	}
	if printchaincmd.Parsed(){
		cli.printchain()
	}
	if getbalancemcd.Parsed(){
		if *getbalanceaddress == ""{
			os.Exit(1)
		}
		cli.getbalance(*getbalanceaddress)
	}

	if sendcmd.Parsed(){
		if *sendcmdvlaue <= 0 || *sendfrom == "" || *sendto == ""{
			os.Exit(1)
		}
		fmt.Println(*sendfrom,*sendto, *sendcmdvlaue)
		cli.send(*sendfrom,*sendto, *sendcmdvlaue)
	}

	if createwalletcmd.Parsed() {
		cli.createwallet()
	}
	if listaddresscmd.Parsed() {
		cli.listaddress()
	}
	if getblockheight.Parsed() {
		cli.getblockheight()
	}

	if startnodecmd.Parsed() {
		nodeid := os.Getenv("NODE_ID")
		if nodeid == "" {
			startnodecmd.Usage()
			os.Exit(1)
		}
		cli.startnode(nodeid, *startnodeminner)
	}
}

func (cli *CLI) startnode(nodeid string, minneraddress string)  {
	fmt.Printf("starting node :%s\n", nodeid)
	if len(minneraddress) > 0{
		if ValidateAddress([]byte(minneraddress)){
			fmt.Println("minner is on ", minneraddress)
		}else{
			log.Panic("error minner address")
		}

	}
	startserver(nodeid, minneraddress,cli.bc)
}

func (cli *CLI)getbalance(address string)()  {
	balance:= 0

	decodeaddress := Base58Decode([]byte(address))
	pubkeyhash := decodeaddress[1:len(decodeaddress) - 4]

	set := utxoset{cli.bc}
	//utxos := cli.bc.findutxo(pubkeyhash)
	utxos := set.findutxobypubkeyhash(pubkeyhash)

	for _,out:= range utxos {
		balance+= out.Value
	}
	fmt.Printf("balance of %s:%d\n",address,balance)
}