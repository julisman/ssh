package example

import (
	"github.com/julisman/ssh"
	"fmt"
)


func main(){
	client := ssh.New(ssh.SSHConfig{"<your ip>", "<your cert>"})
	client.Connect(ssh.CERT_PUBLIC_KEY_FILE)
	result := client.RunCmd("ls")
	fmt.Println(result)
	client.Close()
}
