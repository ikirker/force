package command

import (
	"fmt"
	"strings"

	desktop "github.com/ForceCLI/force/desktop"
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

  force tickets open <ticket id or num>

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
			runTicketsCompletion(args[1:])
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

func runTicketsCompletion(args []string) {
	if len(args) < 2 {
		ErrorAndExit("must specify object and id")
	}
	//force, _ := ActiveForce()
}

func runTicketsOpen(args []string) {
	if len(args) > 3 {
		ErrorAndExit("only one id or number please")
	}
	ticketArg := args[2]

	// In theory this isn't foolproof but it seems to work
	if ticketArg[0] == "a" {
		argType := "Id"
	} else {
		argType := "Name"
	}

	// The URL field you want is BMCServiceDesk__Launch_console__c
	//  but it comes wrapped as an <a href=""></a>
	force, _ := ActiveForce()
	soql := fmt.Sprintf("SELECT %s,BMCServiceDesk__Launch_console__c FROM BMCServiceDesk__Incident__c WHERE %s = '%s' LIMIT 1", argType, argType, ticketArg[0])
	records, err := force.Query(fmt.Sprintf("%s", soql))

	if err != nil {
		ErrorAndExit(err.Error())
	} else {
		// Temporary: for showing guts
		fmt.Printf("%+v\n", records)
	}
	ticketAnchor := records["BMCServiceDesk__Launch_console__c"]
	ticketAnchorParts := strings.Split(ticketAnchor, "\"")
	if len(ticketAnchorParts) != 3 {
		ErrorAndExit("ticket URL was not in the format we expect\n")
	}
	ticketURL := ticketAnchorParts[1]

	err := desktop.Open(ticketURL)
	if err != nil {
		ErrorAndExit(err.Error())
	}
}
