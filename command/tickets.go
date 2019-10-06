package command

import (
	"fmt"
	"strings"

	. "github.com/ForceCLI/force/error"
	. "github.com/ForceCLI/force/lib"
)

var cmdTickets = &Command{
	Run:   runTickets,
	Usage: "tickets <command> [<args>]",
	Short: "View, modify, or summarize tickets",
	Long: `
View, modify, or summarize tickets

Usage:

  force tickets

  force tickets new

  force tickets waiting

  force tickets cli

  force tickets _completion 

  force tickets _completion <partial>

`,
	MaxExpectedArgs: -1,
}

func runTickets(cmd *Command, args []string) {
	if len(args) == 0 {
		cmd.PrintUsage()
	} else {
		switch args[0] {
		case "new":
			runTicketsNew(args[1:])
		case "waiting":
			runTicketsWaiting(args[1:])
		case "cli":
			ErrorAndExit("not yet implemented")
		case "_completion":
			runCompletion(args[1:])
			if len(args) == 3 {
				createBulkInsertJob(args[2], args[1], "CSV", "Parallel")
			} else if len(args) == 4 {
				if strings.EqualFold(args[3], "parallel") || strings.EqualFold(args[3], "serial") {
					createBulkInsertJob(args[2], args[1], "CSV", args[3])
				} else {
					createBulkInsertJob(args[2], args[1], args[3], "Parallel")
				}
			} else if len(args) == 5 {
				if strings.EqualFold(args[3], "parallel") || strings.EqualFold(args[3], "serial") {
					createBulkInsertJob(args[2], args[1], args[4], args[3])
				} else if strings.EqualFold(args[4], "parallel") || strings.EqualFold(args[4], "serial") {
					createBulkInsertJob(args[2], args[1], args[3], args[4])
				}
			}
		default:
			ErrorAndExit("no such command: %s", args[0])
		}
	}
}

func runTicketsNew(args []string) {
	if len(args) > 2 {
		ErrorAndExit("this command takes no arguments") // Maybe could, though: limit number in query?
	}
	force, _ := ActiveForce()
	soql := "SELECT Id,COL_UCL_Summary__c FROM BMCServiceDesk__Incident__c WHERE BMCServiceDesk__Queue__c = 'ISD.Research Computing Support' AND BMCServiceDesk__Status_ID__c = 'NEW' LIMIT 1"
	records, err := force.Query(fmt.Sprintf("%s", soql))

	if err != nil {
		ErrorAndExit(err.Error())
	} else {
		// Temporary: for showing guts
		fmt.Printf("%+v\n", records)
		DisplayForceRecords(records)
	}
}

func runTicketsWaiting(args []string) {
	if len(args) < 1 {
		ErrorAndExit("must specify object")
	}
	//force, _ := ActiveForce()
}

func runCompletion(args []string) {
	if len(args) < 2 {
		ErrorAndExit("must specify object and id")
	}
	//force, _ := ActiveForce()
}
