package main

func main(){

	bc:= newblockchain("12ZiBnmnkqrjWWfhV9zgF5w1xz8nvh7aaR")
	cli:= CLI{bc}
	cli.run()
	//testblockdb()
	//newgenisblock()
	//testserialize()
	//testpow()
	//testcreatemerkletreeroot()
	/*


	wallet := newwallet()


	fmt.Printf("私钥：%x\n",wallet.PrivateKey.D.Bytes())
	fmt.Printf("公钥：%x\n", wallet.Publickey)
	fmt.Printf("地址：%x\n", wallet.GetAddress())
	address,_:=hex.DecodeString("3146536551465558616f7631635a434d4e424a343834707663616754676765473936")
	fmt.Printf("%d\n",ValidateAddress(address))
*/

}

