# GoCDP

Show CDP over snmp

Example run:

```bash
gocdp nei -s 10.0.0.1
```
Command line help
```bash
NAME:
   gocdp - show CDP by snmp

USAGE:
   gocdp [global options] command [command options] [arguments...]
   
VERSION:
   1.0.3
   
AUTHOR(S):
   hdhog <hdhog@hdhog.ru> 
   
COMMANDS:
     neigbors, n, nei  Show CDP neigbors by snmp
     help, h           Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
   
```
```bash
NAME:
   gocdp neigbors - Show CDP neigbors by snmp

USAGE:
   gocdp neigbors [command options] [arguments...]

OPTIONS:
   --community value, -c value  community string (default: "public")
   --host value, -s value       host address

```