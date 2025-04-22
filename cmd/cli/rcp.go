package main

import (
	"fmt"
	"net/rpc"
	"os"

	"github.com/fatih/color"
)

func rpcClient(inMaintenanceMode bool) {
  rcpPort := os.Getenv("RCP_PORT")
	c, err := rpc.Dial("tcp", "127.0.0.1:"+rcpPort)
	if err != nil {
		exitGracefully(err)
	}
	fmt.Println("Connected ...", rcpPort)

  var result string
  err = c.Call("RPCServer.MaintenanceMode", inMaintenanceMode, &result)
  if err != nil {
    exitGracefully(err)
  }
  color.Yellow(result)
}