//rpc_client_crud
package main
	
import (
	"fmt"
	"net/rpc"
	"strings"	
	"strconv"
	"bufio"
	"os"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"	
)

type Args struct {
	name string
	price int
}

var (
	clientIp = "localhost"
	clientPort = "1234"
	client *rpc.Client
	//prefixs to log
	clientPrefix string = "[CLIENT]--"
	errorPrefix string = "[ERROR]--"
	clientPrivateKey *rsa.PrivateKey
	clientPublicKey *rsa.PublicKey
	svPublicKey rsa.PublicKey
	debug = false
)

//funcs to call rpc the server
//create
func callCreate(name string, price string) bool {
	var reply bool
	price = strings.TrimSpace(price)
	fmt.Println(clientPrefix, "Creating product name = " + name +" price = "+ price)
	err := client.Call("Product.Create", name+":"+price , &reply)
	if err != nil {
		fmt.Println(clientPrefix, "Error:", err)
	}
	fmt.Println(clientPrefix, "Server called.")
	if !reply{
		fmt.Println(clientPrefix, "Not ok.")
	}else{
		fmt.Println(clientPrefix, "Ok.")
	}
	return reply
	
}
//read
func callRead(id int) string{
	var reply string
	fmt.Println(clientPrefix, "Trying get product id= ", id)
	err := client.Call("Product.Read", id , &reply)
	if err != nil {
		fmt.Println(clientPrefix, "Error:", err)
	}
	fmt.Println(clientPrefix, "Server called.")	
	if reply == "NE"{
		fmt.Println(clientPrefix, "Not exists.")
	}else if reply == "NA"{
		fmt.Println(clientPrefix, "Not ok.")
	}else{
		fmt.Println(clientPrefix, "Geted.")
	}
	return reply
}
//update
func callUpdate(id string, name string, price string) bool{
	var reply bool
	price = strings.TrimSpace(price)
	fmt.Println(clientPrefix, "Updating product id = ", id ," name = ", name, "  price = ", price, ".")
	err := client.Call("Product.Update", id+":"+name+":"+price , &reply)
	if err != nil {
		fmt.Println(clientPrefix, "Error:", err)
	}
	fmt.Println(clientPrefix, "Server called.")
	if !reply{
		fmt.Println(clientPrefix, "Not ok.")
	}else{
		fmt.Println(clientPrefix, "Ok.")
	}
	return reply
	
}
//delete
func callDelete(id int) bool{
	var reply bool
	fmt.Println(clientPrefix, "Deleting product id = ", id ,".")
	err := client.Call("Product.Delete", id, &reply)
	if err != nil {
		fmt.Println(clientPrefix, "Error:", err)
	}
	fmt.Println(clientPrefix, "Server called.")
	if !reply{
		fmt.Println(clientPrefix, "Not ok.")
	}else{
		fmt.Println(clientPrefix, "Ok.")
	}
	return reply
}
//Change the public keys
func callHandshake(){
	fmt.Println(clientPrefix, "Handshake inited")
	var reply *rsa.PublicKey
	err := client.Call("Product.Handshake", clientPublicKey , &reply)
	if err != nil{
		fmt.Println(errorPrefix, "Handshake error: ", err)
		os.Exit(0)
	}
	svPublicKey = *reply
	fmt.Println(clientPrefix, "Server public key getted")
	fmt.Println(clientPrefix, "Client public key has sent")
}
//Login
func callLogin(login string, pass string) bool{
	var reply bool
	fmt.Println(clientPrefix, "Trying login: " +login + " pass: "+pass)	
	err := client.Call("Product.Login", encrypt(login+":"+pass) , &reply)
	if err != nil {
		fmt.Println(clientPrefix, "Error:", err)
	}
	fmt.Println(clientPrefix, "Server called.")	
	if reply{
		fmt.Println(clientPrefix, "Authentication successful")
	}else{
		fmt.Println(clientPrefix, "Authentication fail")
	}
	return reply
}
//Debug
func callDebug(state int) bool{
	var reply bool
	if state == 1{
		fmt.Println(clientPrefix, "Debug server: ON ")	
	}else if state == 0{
		fmt.Println(clientPrefix, "Debug server: OFF ")	
	}
	
	err := client.Call("Product.DebugController", state , &reply)
	if err != nil {
		fmt.Println(clientPrefix, "Error:", err)
	}
	return reply
}
func callQuit() bool{
	var reply bool
	err := client.Call("Product.Quit", 0 , &reply)
	if err != nil {
		fmt.Println(clientPrefix, "Error:", err)
	}
	return reply
}
//Interpreter for client input commands
func commandInterpreter(command string) (bool,string ){
	if strings.HasPrefix(command, ".") == false{		
		fmt.Println(errorPrefix, "Not a command.")
		return false, ""
	}else{		
		//create command
		//waits a string ".create name price "
		if (strings.HasPrefix(command, ".CREATE")) || (strings.HasPrefix(command, ".create")) {		
			separateString := strings.Split(command, " ")
			if len(separateString) < 3 {
				fmt.Println(errorPrefix, "Create error, few arguments.")
				return false, ""
			}else if len(separateString) > 3 {
				fmt.Println(errorPrefix, "Create error, many arguments.")
				return false, ""
			}else{
				callCreate(separateString[1], separateString[2])
			}
			return true, ""
		//read command
		//waits a string ".read id"
		}else if (strings.HasPrefix(command, ".READ")) || (strings.HasPrefix(command, ".read")) {
			separateString := strings.Split(command, " ")
			if len(separateString) < 2 {
				fmt.Println(errorPrefix, "Read error, few arguments.")
				return false, ""
			}else if len(separateString) > 2 {
				fmt.Println(errorPrefix, "Read error, many arguments.")
				return false, ""
			}else{
				id_ , err := strconv.Atoi(strings.TrimSpace(separateString[1]))
				if err != nil{
					fmt.Println(errorPrefix, "Read error, its not a number.")
				}else{
					response := callRead(id_)
					fmt.Println(clientPrefix, response )
					return true, response
				}				
			}		
		//update command	
		//waits a string ".update id name price"
		}else if(strings.HasPrefix(command, ".UPDATE")) || (strings.HasPrefix(command, ".update")) {
			separateString := strings.Split(command, " ")
			if len(separateString) < 4 {
				fmt.Println(errorPrefix, "Create error, few arguments.")
				return false, ""
			}else if len(separateString) > 4 {
				fmt.Println(errorPrefix, "Create error, many arguments.")
				return false, ""
			}else{
				callUpdate(separateString[1], separateString[2], separateString[3])
			}
			return true, ""
		//delete comand
		//waits a string ".delete id"
		}else if(strings.HasPrefix(command, ".DELETE")) || (strings.HasPrefix(command, ".delete")) {
			separateString := strings.Split(command, " ")		
			if len(separateString) < 2 {
				fmt.Println(errorPrefix, "Delete error, few arguments.")
				return false, ""
			}else if len(separateString) > 2 {
				fmt.Println(errorPrefix, "Delete error, many arguments.")
				return false, ""
			}else{
				id_ , err := strconv.Atoi(strings.TrimSpace(separateString[1]))
				if err != nil{
					fmt.Println(errorPrefix, "Delete error, its not a number.")
				}else{
					callDelete( id_ )
					return true, ""
				}	
			}
		//login comand
		//waits a string ".login login password"
		}else if (strings.HasPrefix(command, ".LOGIN")) || (strings.HasPrefix(command, ".login")){
			separateString := strings.Split(command, " ")
			if len(separateString) < 3 {
				fmt.Println(errorPrefix, "Login error, few arguments.")
				return false, ""
			}else if len(separateString) > 3 {
				fmt.Println(errorPrefix, "Login error, many arguments.")
				return false, ""
			}else{
				loginID := strings.TrimSpace(separateString[2])
				callLogin(separateString[1], loginID)
			}
			return true, ""
		//Turn on the debug cripyt, from client and server
		}else if (strings.HasPrefix(command, ".DEBUG")) || (strings.HasPrefix(command, ".debug")){
			fmt.Println(clientPrefix, "Debug mode: ON")
			callDebug(1)
			debug = true
		//Turn off the debug cripyt, from client and server
		}else if (strings.HasPrefix(command, ".NODEBUG")) || (strings.HasPrefix(command, ".nodebug")){
			fmt.Println(clientPrefix, "Debug mode: OFF")
			callDebug(0)
			debug = false
		//Quit application
		}else if (strings.HasPrefix(command, ".QUIT")) || (strings.HasPrefix(command, ".quit")) {
			fmt.Println(clientPrefix, "Quiting.")
			callQuit()
			return false, "quit"
		//Not a command
		}else{
			fmt.Println(errorPrefix, "Not a valid command.")
			return false, ""
		}
	}
	return true, ""
}
// Generate RSA Keys
func initializeKeys(){
	fmt.Println(clientPrefix, "Generating private key.")
	probablePrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(errorPrefix, err.Error)
		os.Exit(1)
	}
	fmt.Println(clientPrefix, "Generating public key.")
	probablePublicKey := &probablePrivateKey.PublicKey
	clientPrivateKey = probablePrivateKey
	clientPublicKey = probablePublicKey
}
//Connect with rpc
func rpcConnect(){
	var err error
	fmt.Println(clientPrefix, "Trying Rpc dial from " + clientIp+":"+clientPort)
	client, err = rpc.Dial("tcp", clientIp+":"+clientPort)
	if err != nil {
		fmt.Println(errorPrefix, "Dialing error:", err)
		os.Exit(0)
	}
	fmt.Println(clientPrefix, "Rpc dial ok")
	
}
func encrypt(initial string) []byte{
	fmt.Println(clientPrefix, "Cripyting login/pass")
	message := []byte(initial)
	label := []byte("")
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, &svPublicKey, message, label)
	if err != nil {
		fmt.Println(errorPrefix, "Encrypt error: ", err)
		os.Exit(1)
	}
	if debug{
		fmt.Println(clientPrefix, "Initial message: " + initial)
		fmt.Print(clientPrefix, "Encrypted message: ")
		fmt.Println(string(ciphertext))
	}
	fmt.Println(clientPrefix, "Cripyt ok")
	return ciphertext	
}

func main(){
	//Initialize public and private keys
	initializeKeys()
	//Connect with rpc
	rpcConnect()
	//Handshake keys
	callHandshake()
	
	for true {
		fmt.Println(clientPrefix, "Enter a command.")
		fmt.Print(clientPrefix)
		var sentence string
		
		reader := bufio.NewReader(os.Stdin)
		
		sentence, _ = reader.ReadString('\n')

		_, res := commandInterpreter(sentence)
		sentence = ""
		if res == "quit" {
			break
		}
	}
	
	fmt.Println(clientPrefix, "BYE.")
}
