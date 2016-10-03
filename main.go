package main

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/k-sone/snmpgo"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

const (
	oidSysName            = "1.3.6.1.2.1.1.5"
	oidIfDescr            = "1.3.6.1.2.1.2.2.1.2"
	oidCDPCacheEntry      = "1.3.6.1.4.1.9.9.23.1.2.1.1"
	oidCDPCacheDeviceid   = "1.3.6.1.4.1.9.9.23.1.2.1.1.6"
	oidCDPCacheDevicePort = "1.3.6.1.4.1.9.9.23.1.2.1.1.7"
	oidCDPCacheAddress    = "1.3.6.1.4.1.9.9.23.1.2.1.1.4"
	cacheAddress          = 4
	cacheDeviceid         = 6
	cacheDevicePort       = 7
)

// Информация о версии приложения
var (
	Version = "0.0.1"
)

type cdpNeighbor struct {
	LName   string
	LIP     string
	LIfName string
	RName   string
	RIP     string
	RIfName string
}

func compactIfName(ifname string) string {
	if strings.HasPrefix(ifname, "Ten") {
		return strings.Replace(ifname, "TenGigabitEthernet", "Te ", -1)
	}
	if strings.HasPrefix(ifname, "Gig") {
		return strings.Replace(ifname, "GigabitEthernet", "Gi ", -1)
	}
	if strings.HasPrefix(ifname, "Fas") {
		return strings.Replace(ifname, "FastEthernet", "Fa ", -1)
	}
	return ifname
}

func prepareOids() (snmpgo.Oids, error) {
	return snmpgo.NewOids([]string{
		oidCDPCacheDeviceid,
		oidCDPCacheDevicePort,
		oidCDPCacheAddress,
		oidIfDescr,
		oidSysName,
	})
}

func getCDPNeighbors(host, community string) (map[string]cdpNeighbor, error) {
	nei := map[string]cdpNeighbor{}

	snmpAddr := net.JoinHostPort(host, "161")

	snmp, err := snmpgo.NewSNMP(snmpgo.SNMPArguments{
		Version:   snmpgo.V2c,
		Address:   snmpAddr,
		Retries:   1,
		Community: community,
	})

	if err != nil {
		return nil, err
	}

	cacheEntry, err := prepareOids()

	if err != nil {
		return nil, err
	}

	if err = snmp.Open(); err != nil {
		return nil, err
	}
	defer snmp.Close()

	pdu, err := snmp.GetBulkWalk(cacheEntry, 0, 10)
	if err != nil {
		return nil, err
	}

	if pdu.ErrorStatus() != snmpgo.NoError {
		fmt.Println(pdu.ErrorStatus(), pdu.ErrorIndex())
	}

	oid, err := snmpgo.NewOid(oidCDPCacheEntry)
	if err != nil {
		return nil, err
	}
	cacheEntrys := pdu.VarBinds().MatchBaseOids(oid)
	sysName := pdu.VarBinds().MatchBaseOids(snmpgo.MustNewOid(oidSysName))[0]

	for _, val := range cacheEntrys {
		ifindex := fmt.Sprintf("%d", val.Oid.Value[14])
		if _, ok := nei[ifindex]; !ok {
			oid, err = snmpgo.NewOid(oidIfDescr + "." + ifindex)
			if err != nil {
				return nil, err
			}
			ifnames := pdu.VarBinds().MatchBaseOids(oid)
			rifname := ifnames.MatchOid(oid).Variable.String()
			nei[ifindex] = cdpNeighbor{
				RIfName: compactIfName(rifname),
			}
		}

		cdp := nei[ifindex]
		cdp.RName = sysName.Variable.String()
		cdp.RIP = host

		switch val.Oid.Value[13] {
		case cacheAddress:
			cdp.LIP = val.Variable.String()
			var ip, ip1, ip2, ip3 int
			fmt.Sscanf(val.Variable.String(), "%x:%x:%x:%x", &ip, &ip1, &ip2, &ip3)
			cdp.LIP = fmt.Sprintf("%d.%d.%d.%d", ip, ip1, ip2, ip3)
		case cacheDeviceid:
			cdp.LName = val.Variable.String()
		case cacheDevicePort:
			cdp.LIfName = compactIfName(val.Variable.String())
		}

		nei[ifindex] = cdp
	}

	return nei, nil
}

func printTable(neigbors map[string]cdpNeighbor) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "IP", "Local IF", "--", "Remote IF", "Remote IP", "Remote Name"})
	table.SetBorder(false)
	var data [][]string
	var keys []string
	for k := range neigbors {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := neigbors[k]
		data = append(data, []string{v.RName, v.RIP, v.RIfName, "->", v.LIfName, v.LIP, v.LName})
	}
	table.SetAlignment(4)
	table.AppendBulk(data)
	table.Render()
}

func neighbors(c *cli.Context) error {
	community := c.String("community")
	host := c.String("host")
	if host == "" {
		if err := cli.ShowCommandHelp(c, "neigbors"); err != nil {
			return err
		}
		return cli.NewExitError("Host address required", 1)
	}
	res, err := getCDPNeighbors(host, community)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	printTable(res)

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "gocdp"
	app.Usage = "show CDP by snmp"
	app.Version = Version
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "hdhog",
			Email: "hdhog@hdhog.ru",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "neigbors",
			Aliases: []string{"n", "nei"},
			Usage:   "Show CDP neigbors over snmp",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "community, c",
					Value: "public",
					Usage: "community string",
				},
				cli.StringFlag{
					Name:  "host ,s",
					Usage: "host address",
				},
			},
			Action: neighbors,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
