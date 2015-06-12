package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/lochness"
)

type (
	// Refresher writes out the dhcp configuration files hypervisors.conf and
	// guests.conf, given a fetcher
	Refresher struct {
		Domain string
	}

	// templateHelper is used for inserting values into the templates
	templateHelper struct {
		Domain      string
		Hypervisors []hypervisorHelper
		Guests      []guestHelper
	}

	// hypervisorHelper is used for inserting a hypervisor's values into the template
	hypervisorHelper struct {
		ID      string
		MAC     string
		IP      string
		Gateway string
		Netmask string
	}

	// guestHelper is used for inserting a guest's values into the template
	guestHelper struct {
		ID      string
		MAC     string
		IP      string
		Gateway string
		CIDR    string
	}
)

var hypervisorsTemplate = `
# Auto generated by cdhcpd, do not edit

group hypervisors {
    option domain-name "nodes.{{.Domain}}";
    if exists user-class and option user-class = "iPXE" {
        filename "http://ipxe.services.{{.Domain}}:8888/ipxe/${net0/ip}";
    } else {
        next-server tftp.services.{{.Domain}};
        filename "undionly.kpxe";
    }
{{range $h := .Hypervisors}}
    host {{$h.ID}} {
        hardware ethernet  {{$h.MAC}};
        fixed-address      {{$h.IP}};
        option routers     {{$h.Gateway}};
        option subnet-mask {{$h.Netmask}};
    }
{{end}}
}
`

var guestsTemplate = `
# Auto generated by cdhcpd, do not edit

group guests {
    option domain-name "guests.{{.Domain}}";
{{range $g := .Guests}}
    host {{$g.ID}} {
        hardware ethernet  {{$g.MAC}};
        fixed-address      {{$g.IP}};
        option routers     {{$g.Gateway}};
        option subnet-mask {{$g.CIDR}};
    }
{{end}}
}
`

// NewRefresher creates a new refresher
func NewRefresher(domain string) *Refresher {
	return &Refresher{
		Domain: domain,
	}
}

// genHypervisorsConf writes the hypervisors config
func (r *Refresher) genHypervisorsConf(w io.Writer, hypervisors map[string]*lochness.Hypervisor) error {
	vals := new(templateHelper)
	vals.Domain = r.Domain

	// Sort keys
	hkeys := make([]string, len(hypervisors))
	i := 0
	for id := range hypervisors {
		hkeys[i] = id
		i++
	}
	sort.Strings(hkeys)

	// Loop through and build up the templateHelper
	for _, id := range hkeys {
		hv := hypervisors[id]
		vals.Hypervisors = append(vals.Hypervisors, hypervisorHelper{
			ID:      hv.ID,
			MAC:     strings.ToUpper(hv.MAC.String()),
			IP:      hv.IP.String(),
			Gateway: hv.Gateway.String(),
			Netmask: hv.Netmask.String(),
		})
	}

	// Execute template
	t, err := template.New("hypervisors.conf").Parse(hypervisorsTemplate)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "template.Parse",
		}).Error("Could not parse hypervisors.conf template")
		return err
	}
	if err = t.Execute(w, vals); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "template.Execute",
		}).Error("Could not execute hypervisors.conf template")
		return err
	}
	return nil
}

// genGuestsConf writes the guests config
func (r *Refresher) genGuestsConf(w io.Writer, guests map[string]*lochness.Guest, subnets map[string]*lochness.Subnet) error {
	vals := new(templateHelper)
	vals.Domain = r.Domain

	// Sort guest keys
	gkeys := make([]string, len(guests))
	i := 0
	for id := range guests {
		gkeys[i] = id
		i++
	}
	sort.Strings(gkeys)

	// Loop through and build up the templateHelper
	for _, id := range gkeys {
		g := guests[id]
		if g.HypervisorID == "" || g.SubnetID == "" {
			continue
		}
		s, ok := subnets[g.SubnetID]
		if !ok {
			continue
		}
		mask := s.CIDR.Mask
		vals.Guests = append(vals.Guests, guestHelper{
			ID:      g.ID,
			MAC:     strings.ToUpper(g.MAC.String()),
			IP:      g.IP.String(),
			Gateway: s.Gateway.String(),
			CIDR:    fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]),
		})
	}

	// Execute template
	t, err := template.New("guests.conf").Parse(guestsTemplate)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "template.Parse",
		}).Error("Could not parse guests.conf template")
		return err
	}
	if err = t.Execute(w, vals); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "template.Execute",
		}).Error("Could not execute guests.conf template")
		return err
	}
	return nil
}
