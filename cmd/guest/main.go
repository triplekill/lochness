package main

import (
	"encoding/json"
	"fmt"

	"code.google.com/p/go-uuid/uuid"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	server  = "http://localhost:18000/"
	jsonout = false
	t       = "application/json"
)

type (
	jmap map[string]interface{}
)

func (j jmap) ID() string {
	return j["id"].(string)
}

func (j jmap) String() string {
	buf, err := json.Marshal(&j)
	if err != nil {
		return ""
	}
	return string(buf)
}

func (g jmap) Print() {
	if jsonout {
		fmt.Println(g)
	} else {
		fmt.Println(g.ID())
	}
}

func assertID(id string) {
	if uuid := uuid.Parse(id); uuid == nil {
		log.WithField("id", id).Fatal("invalid id")
	}
}

func assertSpec(spec string) {
	j := jmap{}
	if err := json.Unmarshal([]byte(spec), &j); err != nil {
		log.WithFields(log.Fields{
			"spec":  spec,
			"error": err,
		}).Fatal("invalid spec")
	}
}

func help(cmd *cobra.Command, _ []string) {
	cmd.Help()
}

func getGuests(c *client) []jmap {
	ret := c.getMany("guests", "guests")
	guests := make([]jmap, len(ret))
	for i := range ret {
		guests[i] = ret[i]
	}
	return guests
}

func getGuest(c *client, id string) jmap {
	return c.get("guest", "guests/"+id)
}

func createGuest(c *client, spec string) jmap {
	return c.post("guest", "guests", spec)
}

func modifyGuest(c *client, id string, spec string) jmap {
	return c.patch("guest", "guests/"+id, spec)
}

func deleteGuest(c *client, id string) jmap {
	return c.del("hypervisor", "guests/"+id)
}

func list(cmd *cobra.Command, ids []string) {
	c := newClient(server)
	guests := []jmap{}

	if len(ids) == 0 {
		guests = getGuests(c)
	} else {
		for _, id := range ids {
			assertID(id)
			guests = append(guests, getGuest(c, id))
		}
	}

	for _, guest := range guests {
		guest.Print()
	}
}

func create(cmd *cobra.Command, specs []string) {
	c := newClient(server)
	for _, spec := range specs {
		assertSpec(spec)
		guest := createGuest(c, spec)
		guest.Print()
	}
}

func modify(cmd *cobra.Command, args []string) {
	c := newClient(server)
	if len(args)%2 != 0 {
		log.WithField("num", len(args)).Fatal("expected an even number of args")
	}
	for i := 0; i < len(args); i += 2 {
		id := args[i]
		assertID(id)
		spec := args[i+1]
		assertSpec(spec)

		guest := modifyGuest(c, id, spec)
		guest.Print()
	}
}

func del(cmd *cobra.Command, ids []string) {
	c := newClient(server)
	for _, id := range ids {
		assertID(id)
		guest := deleteGuest(c, id)
		guest.Print()
	}
}

func main() {
	root := &cobra.Command{
		Use:   "guest",
		Short: "guest is the cli interface to waheela",
		Run:   help,
	}
	root.PersistentFlags().BoolVarP(&jsonout, "jsonout", "j", jsonout, "output in json")
	root.PersistentFlags().StringVarP(&server, "server", "s", server, "server address to connect to")

	cmdList := &cobra.Command{
		Use:   "list [<id>...]",
		Short: "List the guests",
		Run:   list,
	}

	cmdCreate := &cobra.Command{
		Use:   "create <spec>...",
		Short: "Create guests",
		Long:  `Create new guest(s) using "spec"(s) as the initial values. Where "spec" is a valid json string.`,
		Run:   create,
	}

	cmdModify := &cobra.Command{
		Use:   "modify (<id> <spec>)...",
		Short: "Modify guests",
		Long:  `Modify given guest(s). Where "spec" is a valid json string.`,
		Run:   modify,
	}

	cmdDelete := &cobra.Command{
		Use:   "delete <id>...",
		Short: "Delete guests",
		Run:   del,
	}

	root.AddCommand(cmdList, cmdCreate, cmdModify, cmdDelete)
	root.Execute()
}
