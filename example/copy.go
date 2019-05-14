package example

import (
	"github.com/julisman/ssh/ssh"
	"fmt"
)


func main(){
	client := ssh.New(ssh.SSHConfig{"<your ip>", "<your cert>"})
	client.Connect(ssh.CERT_PUBLIC_KEY_FILE)
	err = client.CopyFile("<origin path>", "<dest path>")
	if err != nil {
		fmt.Errorln("Error SSH")
	}
	client.Close()
}
