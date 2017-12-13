//rpc_server_crud
package main

import (
    "fmt"
    "net"
	"net/rpc"
	"strings"
	"strconv"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"	
	"os"
	
)

//Struct to product
type Product struct {
	id 		int
	name, price	string
}

var (
	serverIp = "localhost"
	serverPort = "1234"
	//"database" int to product
	database map[int]Product
	//next id to database
	globalNextId int

	//prefixs to log
	dbPrefix string = "[DATABASE]--"
	serverPrefix string = "[SERVER]--"
	errorPrefix string = "[ERROR]--"

	serverPrivateKey *rsa.PrivateKey
	serverPublicKey *rsa.PublicKey
	cliPublicKey rsa.PublicKey

	adminID = "admin"
	adminPASS = "chatuba"

	authenticated = false
	svDebug = false
)

//func to create a new product into database
//gets a string type -> "name:price" ,ex ->"arroz:27"
func (pro *Product) Create(prod string, reply *bool) error {	
	if !authenticated{
		fmt.Println(errorPrefix, "Not authenticated")
		*reply = false
		return nil
	}
	separateString := strings.Split(prod, ":")
	newProduct := Product{
		id: 	globalNextId,
		name: 	separateString[0],
		price: 	separateString[1],
	}
	put(globalNextId, newProduct)
	globalNextId++
	fmt.Println(serverPrefix,"Global ID incresed, new = ", globalNextId)
    *reply = true
    return nil
}
//func to read a product in database
//gets a int type -> "id"
func (pro *Product) Read(id int, reply *string) error{
	if !authenticated{
		fmt.Println(errorPrefix, "Not authenticated")
		*reply = "NA"
		return nil
	}
	var returnProduct 	Product
	var ex 				bool
	returnProduct, ex = get(id)
	if ex == false{
		fmt.Println(dbPrefix, "The product with id = ", id ," not exists.")
		*reply = "NE"
		return nil
	}
	*reply = "name: " + returnProduct.name + " | price: " + returnProduct.price +"."
	return nil	
}
//func to update a product in database
//gets a string type -> "id:name:price", ex -> "1:arroz:2"
func (pro *Product) Update(prod string, reply *bool) error{
	if !authenticated{
		fmt.Println(errorPrefix, "Not authenticated")
		*reply = false
		return nil
	}
	separateString := strings.Split(prod, ":")
	id_, _ :=	strconv.Atoi(separateString[0])
	newProduct := Product{
		id: 	id_,
		name: 	separateString[1],
		price: 	separateString[2],
	}
	update(id_, newProduct)	
    *reply = true
	return nil
}
//func to read a product in database
//gets a int type -> "id"
func (pro *Product) Delete(id int, reply *bool) error{
	if !authenticated{
		fmt.Println(errorPrefix, "Not authenticated")
		*reply = false
		return nil
	}
	remove(id)
	*reply = true
	return nil
}
//Handshake to change public keys
func (pro *Product) Handshake(publicKey rsa.PublicKey, reply *rsa.PublicKey) error{	
	fmt.Println(serverPrefix, "Handshake inited")
	cliPublicKey = publicKey
	*reply = *serverPublicKey
	fmt.Println(serverPrefix, "Client public key getted")
	fmt.Println(serverPrefix, "Server public key has sent")
	return nil
}
//Check the login
func (pro *Product) Login(data []byte, reply *bool) error{
//func (pro *Product) Login(data string, reply *bool) error{
	message := decrypt(data)
	separateString := strings.Split(message, ":")
	login := separateString[0]
	pass := separateString[1]
	fmt.Println(serverPrefix, "Trying login: " +login +" pass: " + pass)
	if (login == adminID) && (pass == adminPASS){
		fmt.Println(serverPrefix, "Login successful")
		authenticated = true
		*reply = true
		return nil
	}else{
		fmt.Println(errorPrefix, "Login error, wrong datas")
		*reply = false
		return nil
	}	
}
//Control the debug
func (pro *Product) DebugController(state int, reply *bool) error{
	if state == 1{
		fmt.Println(serverPrefix, "Debug mode: ON")
		svDebug = true
	}else if state == 0{
		fmt.Println(serverPrefix, "Debug mode: OFF")
		svDebug = false
	}
	*reply = true
	return nil
}
//Quit client
func (pro *Product) Quit(state int, reply *bool) error {
	authenticated = false
	*reply = true
	return nil
}
// Generate RSA Keys
func initializeServerKeys(){
	fmt.Println(serverPrefix, "Generating private key.")
	probablePrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(errorPrefix, err.Error)
		os.Exit(1)
	}
	fmt.Println(serverPrefix, "Generating public key.")
	probablePublicKey := &probablePrivateKey.PublicKey
	serverPrivateKey = probablePrivateKey
	serverPublicKey = probablePublicKey
}
func decrypt(ciphertext []byte) string {
	// Decrypt Message
	fmt.Println(serverPrefix, "Decrypting login/pass")
	hash := sha256.New()
	label := []byte("")

	plainText, err := rsa.DecryptOAEP(hash, rand.Reader, serverPrivateKey, ciphertext, label)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(serverPrefix, "Decrypt ok")
	if svDebug{
		fmt.Print(serverPrefix, "Encrypted message: ")
		fmt.Println(string(ciphertext))
		fmt.Println(serverPrefix, "Original message: " +string(plainText))
	
	}
	
	return string(plainText)
}

func main(){
	
	//db	
	init_db()
	//Initialize public and private keys
	initializeServerKeys()

	//Interface to rpc
    interf := new(Product)
	rpc.Register(interf)
	//log
	fmt.Println(serverPrefix,"RPC register interface.")
	//TCP connection
	//Address
	tcpAddr, err := net.ResolveTCPAddr("tcp", ""+ serverIp + ":" + serverPort)
	if err != nil {
		fmt.Println(errorPrefix,"Error on solve TCP address.")	
	}
	//Listener
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println(errorPrefix,"Error on listen TCP address.")	
	}
	//log
	fmt.Println(serverPrefix, "Server working on", serverIp,":",serverPort)
    //Accept connections
    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
		rpc.ServeConn(conn)
		fmt.Println(serverPrefix, "Connection accepted, ", conn, ".")
	}
}
//database funcs
//instantiates the database
func init_db() {
	globalNextId = 1
	fmt.Println(dbPrefix,"Database initiated.")
	database = make(map[int]Product)
	fmt.Println(dbPrefix, "Database working.")
}
// inserts a new product in databese or updates its
func put(id int, pro Product) {
	if !authenticated{
		fmt.Println(errorPrefix, "Not authenticated")
		os.Exit(0)
	}
	fmt.Println(dbPrefix,"Puting product { id = ",id," | name = ", pro.name  ,"}")
	database[id] = pro
	fmt.Println(dbPrefix,"Inserted.")
}
//update a product from the id
func update(id int, pro Product) {
	if !authenticated{
		fmt.Println(errorPrefix, "Not authenticated")
		os.Exit(0)
	}
	fmt.Println(dbPrefix,"Updating product { id = ",id," | name = ", pro.name  ,"}")
	database[id] = pro
	fmt.Println(dbPrefix,"Updated.")
}
// get the product from the id, and return if the produc in the database
func get(id int) (Product, bool) {
	if !authenticated{
		fmt.Println(errorPrefix, "Not authenticated")
		os.Exit(0)
	}
	fmt.Println(dbPrefix,"Getting product { id = ",id,"}")
	pro, ex := database[id]
	fmt.Println(dbPrefix, "Got.")
    return pro, ex
}
//remove the product from the id
func remove(id int){
	if !authenticated{
		fmt.Println(errorPrefix, "Not authenticated")
		os.Exit(0)
	}
	fmt.Println(dbPrefix,"Deleting product { id = ",id, "}")
	delete(database, id)
	fmt.Println(dbPrefix,"Deleted.")
}