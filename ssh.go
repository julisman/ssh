package ssh

import (
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"time"
	"fmt"
	"io"
	"path"
	"log"
	"os"
)

const (
	CERT_PASSWORD = 1
	CERT_PUBLIC_KEY_FILE = 2
	DEFAULT_TIMEOUT = 3 // second
)

type SSHConfig struct {
	IP  string
	Cert string
	User string
	Port int
}

type SSH struct{
	Ip string
	User string
	Cert string //password or key file path
	Port int
	session *ssh.Session
	client *ssh.Client
}


func (ssh_client *SSH) readPublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func (ssh_client *SSH) Connect(mode int) {

	var ssh_config *ssh.ClientConfig
	var auth  []ssh.AuthMethod
	if mode == CERT_PASSWORD {
		auth = []ssh.AuthMethod{ssh.Password(ssh_client.Cert)}
	} else if mode == CERT_PUBLIC_KEY_FILE {
		auth = []ssh.AuthMethod{ssh_client.readPublicKeyFile(ssh_client.Cert)}
	} else {
		log.Println("does not support mode: ", mode)
		return
	}

	ssh_config = &ssh.ClientConfig{
		User: ssh_client.User,
		Auth: auth,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout:time.Second * DEFAULT_TIMEOUT,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ssh_client.Ip, ssh_client.Port), ssh_config)
	if err != nil {
		fmt.Println(err)
		return
	}

	session, err := client.NewSession()
	if err != nil {
		fmt.Println(err)
		client.Close()
		return
	}

	ssh_client.session = session
	ssh_client.client  = client
}

func (ssh_client *SSH) RunCmd(cmd string) string {
	out, err := ssh_client.session.CombinedOutput(cmd)
	if err != nil {
		fmt.Println(err)
	}
	return string(out)
}

func (ssh_client *SSH) Start(cmd string) error {
	err := ssh_client.session.Start(cmd)
	if err != nil {
		return err
		fmt.Println(err)
	}
	return nil
}

func (ssh_client *SSH) Wait()  {
	err := ssh_client.session.Wait()
	if err != nil {
		fmt.Println(err)
	}
}

func (ssh_client *SSH) StdinPipe()  io.WriteCloser{
	r, err := ssh_client.session.StdinPipe()
	if err != nil {
		fmt.Println(err)
	}
	return r
}

func (ssh_client *SSH) CopyFile(filePath, destinationPath string)  error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	s, err := f.Stat()
	if err != nil {
		return err
	}
	return ssh_client.copy(s.Size(), s.Mode().Perm(), path.Base(filePath), f, destinationPath)

}

func (ssh_client *SSH) copy(size int64, mode os.FileMode, fileName string, contents io.Reader, destination string) error {
	defer ssh_client.Close()
	w, err := ssh_client.session.StdinPipe()

	if err != nil {
		return err
	}

	if err := ssh_client.session.Start(fmt.Sprintf("scp -t %v", destination)); err != nil {
		w.Close()
		return err
	}

	errors := make(chan error)

	go func() {
		errors <- ssh_client.session.Wait()
	}()

	fmt.Fprintf(w, "C%#o %d %s\n", mode, size, fileName)
	io.Copy(w, contents)
	fmt.Fprint(w, "\x00")
	w.Close()

	return <-errors
}

func (ssh_client *SSH) Close() {
	ssh_client.session.Close()
	ssh_client.client.Close()
}

func New(cfg SSHConfig) *SSH{

	if cfg.User == "" {
		cfg.User = "root"
	}

	if cfg.Port == 0 {
		cfg.Port = 22
	}

	client := &SSH{
		Ip: cfg.IP,
		User : cfg.User,
		Port:cfg.Port,
		Cert:cfg.Cert,
	}
	return client
}

func (ssh_client *SSH) Gui() error{
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	err := ssh_client.session.RequestPty("xterm", 80, 40, modes)
	if err != nil {
		return err
	}

	stdBuf, err := ssh_client.session.StdoutPipe()
	if err != nil {
		return err
	}
	go io.Copy(os.Stdout, stdBuf)

	stdinBuf, err := ssh_client.session.StdinPipe()
	if err != nil {
		return err
	}
	go io.Copy(stdinBuf, os.Stdin)

	err = ssh_client.session.Shell()
	if err != nil {
		return err
	}

	stderr, err := ssh_client.session.StderrPipe()

	if err != nil {
		return err
	}

	go io.Copy(os.Stderr, stderr)


	return nil

}
