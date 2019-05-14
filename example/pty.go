package example

import (
	"github.com/julisman/ssh"
	"fmt"
)


func main(){
	client := ssh.New(ssh.SSHConfig{"<your ip>", "<your cert>"})
	client.Connect(ssh.CERT_PUBLIC_KEY_FILE)

	for {
		client.Gui()
	}
}
