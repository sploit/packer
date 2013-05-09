package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// A Environment is an implementation of the packer.Environment interface
// where the actual environment is executed over an RPC connection.
type Environment struct {
	client *rpc.Client
}

// A EnvironmentServer wraps a packer.Environment and makes it exportable
// as part of a Golang RPC server.
type EnvironmentServer struct {
	env packer.Environment
}

type EnvironmentCliArgs struct {
	Args []string
}

func (e *Environment) Builder(name string) (b packer.Builder, err error) {
	var reply string
	err = e.client.Call("Environment.Builder", name, &reply)
	if err != nil {
		return
	}

	client, err := rpc.Dial("tcp", reply)
	if err != nil {
		return
	}

	b = Builder(client)
	return
}

func (e *Environment) Cli(args []string) (result int, err error) {
	rpcArgs := &EnvironmentCliArgs{args}
	err = e.client.Call("Environment.Cli", rpcArgs, &result)
	return
}

func (e *Environment) Ui() packer.Ui {
	var reply string
	e.client.Call("Environment.Ui", new(interface{}), &reply)

	// TODO: error handling
	client, _ := rpc.Dial("tcp", reply)
	return &Ui{client}
}

func (e *EnvironmentServer) Builder(name *string, reply *string) error {
	builder, err := e.env.Builder(*name)
	if err != nil {
		return err
	}

	// Wrap it
	server := rpc.NewServer()
	RegisterBuilder(server, builder)

	*reply = serveSingleConn(server)
	return nil
}

func (e *EnvironmentServer) Cli(args *EnvironmentCliArgs, reply *int) (err error) {
	*reply, err = e.env.Cli(args.Args)
	return
}

func (e *EnvironmentServer) Ui(args *interface{}, reply *string) error {
	ui := e.env.Ui()

	// Wrap it
	server := rpc.NewServer()
	RegisterUi(server, ui)

	*reply = serveSingleConn(server)
	return nil
}