package linstor



import (
    "context"
    "github.com/LINBIT/golinstor/client"
    log "github.com/sirupsen/logrus"
    "regexp"
    "fmt"
    "golang.org/x/crypto/ssh"
    "time"
    //"os"
    "io/ioutil"
    //"path"
)

type SshConnect struct{
    Sshclient *ssh.Client
    host string
}


type SshConnectList struct {
    Connects []*SshConnect


}

var sshlist = &SshConnectList{}

func SSHConnect(user, host string, port int) (*ssh.Client, error) {
    var (
        addr         string
        clientConfig *ssh.ClientConfig
        client       *ssh.Client
        err          error
    )
         
    //homePath, err := os.UserHomeDir()
    //if err != nil {
      //  return nil, err
    //}
    //key, err := ioutil.ReadFile(path.Join(homePath, ".ssh", "id_rsa"))
    key, err := ioutil.ReadFile("/etc/localssh/id_rsa")
    //key, err := ioutil.ReadFile("/root/.ssh/id_rsa")
    if err != nil {
        return nil, err
    }
    signer, err := ssh.ParsePrivateKey(key)
    if err != nil {
        return nil, err
    }
    
    clientConfig = &ssh.ClientConfig{
        User: user,
        Auth: []ssh.AuthMethod{
            ssh.PublicKeys(signer),
        },
        Timeout:         30 * time.Second,
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    // connet to ssh
    addr = fmt.Sprintf("%s:%d", host, port)

    if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
        return nil, err
    }
    return client, nil
}

func GetSshList(ctx context.Context, c *client.Client) {

	if sshlist.Connects == nil{
        DoSshs(ctx, c)
	}
}

func DoSshs(ctx context.Context, c *client.Client) {

    data := GetNodesIP(ctx, c)

    for _, node := range data {
        reg := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)
        result := reg.FindAllStringSubmatch(node["addr"], -1)
	fmt.Println(result[0][0])
        if result[0][0] != "10.203.1.240" {
            sshclient, err := SSHConnect("root", result[0][0], 22)
            if err != nil{
                sshlist.Connects = nil
                log.Fatal(err) 
            }
            sc := &SshConnect{sshclient, node["name"]}
            sshlist.Connects = append(sshlist.Connects, sc)
        }  
    }
}

func GetDevices(ctx context.Context, c *client.Client) {
    GetSshList(ctx, c)
    for _, cli := range sshlist.Connects{

        session, err := cli.Sshclient.NewSession()
        if err != nil {
            log.Fatal("创建ssh session 失败",err)
        }
        defer session.Close()
        //执行远程命令
        combo,err := session.CombinedOutput("whoami; cd /; ls -al")
        if err != nil {
            log.Fatal("远程执行cmd 失败",err)
        }
        log.Println("命令输出:",string(combo))

    }

}



func GetNodesIP(ctx context.Context, c *client.Client) []map[string]string {

    nodesInfo := []map[string]string{}

    data := GetNodeData(ctx, c)

    for _, node := range data {


        nodeInfo := map[string]string{
            "name":           node["name"],
            "addr":           node["addr"],
        }
        nodesInfo = append(nodesInfo, nodeInfo)

    }
    return nodesInfo
}












