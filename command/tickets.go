package command

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/k3a/html2text" // Converts Incident History richtext to plaintext

	desktop "github.com/ForceCLI/force/desktop"
	. "github.com/ForceCLI/force/error"
	. "github.com/ForceCLI/force/lib"

	"html/template"
	"log"
	"net/http"
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
  
	force tickets recent
	
	force tickets describe <ids>

	force tickets history <id>

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
		case "recent":
			runTicketsRecent(args[1:])
		case "cli":
			ErrorAndExit("not yet implemented")
		case "describe":
			runGetTicketDescriptions(args[1:])
		case "history":
			runGetTicketHistory(args[1:])
		case "open":
			runTicketsOpen(args[1:])
		case "serve":
			runTicketServer(args[1:])
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

func getTicketsWithStatuses(statuses []string, maxNum int) (ForceQueryResult, error) {
	force, _ := ActiveForce()

	limitText := ""
	if maxNum != 0 {
		limitText = fmt.Sprintf("LIMIT %d", maxNum)
	}

	var statusText string
	if len(statuses) != 0 {
		statusTextParts := []string{}
		for _, status := range statuses {
			statusTextParts = append(statusTextParts, fmt.Sprintf("(BMCServiceDesk__Status_ID__c = '%s')", status))
		}
		statusText = " AND (" + strings.Join(statusTextParts, " OR ") + ")"
	} else {
		statusText = ""
	}

	soql := fmt.Sprintf("SELECT Id,COL_UCL_Summary__c,LastModifiedDate FROM BMCServiceDesk__Incident__c WHERE BMCServiceDesk__Queue__c = 'ISD.Research Computing Support' %s ORDER BY LastModifiedDate DESC %s", statusText, limitText)
	return force.Query(fmt.Sprintf("%s", soql))
}

func runTicketsNew(args []string) {
	if len(args) > 0 {
		ErrorAndExit("this command takes no arguments") // Maybe could, though: limit number in query?
	}
	maxNum := 15
	records, err := getTicketsWithStatuses([]string{"NEW"}, maxNum)

	if err != nil {
		ErrorAndExit(err.Error())
	} else {
		// Temporary: for showing guts
		fmt.Printf("%+v\n", records)
		DisplayForceRecords(records)
	}
}

func runTicketsWaiting(args []string) {
	if len(args) > 0 {
		ErrorAndExit("this command takes no arguments") // Maybe could, though: limit number in query?
	}
	maxNum := 15
	records, err := getTicketsWithStatuses([]string{"NEW", "CUSTOMER RESPONDED"}, maxNum)

	if err != nil {
		ErrorAndExit(err.Error())
	} else {
		// Temporary: for showing guts
		fmt.Printf("%+v\n", records)
		DisplayForceRecords(records)
	}
}

func runTicketsRecent(args []string) {
	if len(args) > 0 {
		ErrorAndExit("this command takes no arguments") // Maybe could, though: limit number in query?
	}
	maxNum := 50
	records, err := getTicketsWithStatuses([]string{}, maxNum)

	if err != nil {
		ErrorAndExit(err.Error())
	} else {
		DisplayForceRecords(records)
	}
}

func runTicketsCompletion(args []string) {
	if len(args) < 2 {
		ErrorAndExit("must specify object and id")
	}
	//force, _ := ActiveForce()
}

func runGetTicketDescriptions(args []string) {
	if len(args) < 1 {
		ErrorAndExit("must give a ticket id")
	}

	for index, ticketArg := range args {

		// In theory this isn't foolproof but it seems to work
		if ticketArg[0] != byte('a') {
			ticketArg = strings.TrimPrefix(ticketArg, "IN:")
			ticketArg = strings.TrimPrefix(ticketArg, "IN")
			args[index] = "Name:" + ticketArg
		}
	}

	tickets, err := GetTicketsByID(args)
	if err != nil {
		panic(err)
	}

	for _, ticket := range tickets {
		fmt.Println("Id: ", ticket["Id"])
		fmt.Println("User: ", ticket["UCL_userid_UPI__c"])
		fmt.Println("Ticket Subject: ", ticket["COL_UCL_Summary__c"])
		fmt.Println("-----------------------------------------------------")
		fmt.Println(ticket["BMCServiceDesk__incidentDescription__c"])
		fmt.Println("-----------------------------------------------------")
	}
}

func YamlDumpTicket(ticket ForceRecord) {
	yamlBytes, _ := yaml.Marshal(ticket)
	fmt.Println(string(yamlBytes))
}

func ticketNameToId(name string) string {
	name = strings.TrimPrefix(name, "IN:")
	name = strings.TrimPrefix(name, "IN")
	name = "Name:" + name
	var ticketArr []ForceRecord
	ticketArr, err := GetTicketsByID([]string{name})
	if err != nil {
		panic(err)
	}
	return ticketArr[0]["Id"].(string)
}

func ensureIsID(nameOrID string) string {
	if nameOrID[0] != byte('a') {
		return ticketNameToId(nameOrID)
	}
	return nameOrID
}

func runGetTicketHistory(args []string) {
	if len(args) < 1 {
		ErrorAndExit("must give a ticket id")
	}
	if len(args) > 1 {
		ErrorAndExit("can only get history for one ticket at a time")
	}

	id := ensureIsID(args[0])
	records, err := GetHistories(id)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%+v\n", records)
	for _, record := range records.Records {
		fmt.Println("Id: ", record["Id"])
		fmt.Println("User: ", record["BMCServiceDesk__userId__c"])
		fmt.Println("Action: ", record["BMCServiceDesk__actionId__c"])
		fmt.Println("-----------------------------------------------------")
		fmt.Println(html2text.HTML2Text(record["BMCServiceDesk__RichTextNote__c"].(string)))
		fmt.Println("-----------------------------------------------------")
	}
}

func GetHistories(ticketID string) (ForceQueryResult, error) {
	force, _ := ActiveForce()

	limitText := ""
	soql := fmt.Sprintf("SELECT Id,BMCServiceDesk__RichTextNote__c,BMCServiceDesk__userId__c,BMCServiceDesk__actionId__c FROM BMCServiceDesk__IncidentHistory__c WHERE BMCServiceDesk__FKIncident__c = '"+ticketID+"' ORDER BY LastModifiedDate DESC %s", limitText)
	return force.Query(fmt.Sprintf("%s", soql))
}

func GetTicketsByID(ticketIDs []string) ([]ForceRecord, error) {
	tickets := []ForceRecord{}

	force, _ := ActiveForce()
	for _, arg := range ticketIDs {
		record, err := force.GetRecord("BMCServiceDesk__Incident__c", arg)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, record)
	}
	return tickets, nil
}

func runTicketsOpen(args []string) {
	if len(args) > 1 {
		ErrorAndExit("only one id or number please")
	}
	ticketArg := args[0]

	// In theory this isn't foolproof but it seems to work
	var argType string
	if ticketArg[0] == byte('a') {
		argType = "Id"
	} else {
		argType = "Name"
	}

	// The URL field you want is BMCServiceDesk__Launch_console__c
	//  but it comes wrapped as an <a href=""></a>
	force, _ := ActiveForce()
	soql := fmt.Sprintf("SELECT %s,BMCServiceDesk__Launch_console__c FROM BMCServiceDesk__Incident__c WHERE %s = '%s' LIMIT 1", argType, argType, ticketArg)
	records, err := force.Query(fmt.Sprintf("%s", soql))

	if err != nil {
		ErrorAndExit(err.Error())
	} else {
		// Temporary: for showing guts
		fmt.Printf("%+v\n", records)
	}
	ticketAnchor := records.Records[0]["BMCServiceDesk__Launch_console__c"].(string)
	ticketAnchorParts := strings.Split(ticketAnchor, "\"")
	// 2 for href="", 2 for target="", 1 for fence-post
	if len(ticketAnchorParts) != 5 {
		ErrorAndExit("ticket URL was not in the format we expect\n")
	}
	ticketURL := ticketAnchorParts[1]
	// Unfortunately I can't work out the proper prefix for these URLs right now
	fmt.Println(ticketURL)
	if false {
		err = desktop.Open(ticketURL)
	}
	if err != nil {
		ErrorAndExit(err.Error())
	}
}

/// Server code

func getURLReqKey(req *http.Request, key string) string {
	keys, err := req.URL.Query()[key]

	if !err || len(keys[0]) < 1 {
		log.Fatalln("Url Param ", key, " is missing")
		return ""
	}

	// Query()["key"] will return an array of items,
	// we only want the single item.
	val := keys[0]
	return val
}

// Deinterface
func dis(val interface{}) string {
	switch v := val.(type) {
	case int:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("%V", v)
	}
}

// Super-simple plaintext-to-HTML to avoid using <pre> or ignoring all whitespace
func p2h(s string) template.HTML {
	o := template.HTMLEscapeString(s)
	o = strings.ReplaceAll(o, "\n", "<br>\n")
	return (template.HTML)(o)
}

func ticketServerHandler(w http.ResponseWriter, req *http.Request) {
	ticketName := getURLReqKey(req, "num")

	ticketName = strings.TrimPrefix(ticketName, "IN:")
	ticketName = strings.TrimPrefix(ticketName, "IN")

	ticketRecords, err := GetTicketsByID([]string{"Name:" + ticketName})
	if err != nil {
		templ := getErrorTemplate()
		err := templ.Execute(w, err.Error())
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if len(ticketRecords) == 0 {
		// Throw back empty
		templ := getErrorTemplate()
		err := templ.Execute(w, "Ticket not found")
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	tr := ticketRecords[0]

	var ticket templTicket
	ticket.Number = ticketName
	ticket.Summary = dis(tr["COL_UCL_Summary__c"])
	ticket.Headers = []templTicketHeaderField{
		templTicketHeaderField{Name: "Created On", Value: dis(tr["CreatedDate"])},
		templTicketHeaderField{Name: "Last Modified", Value: dis(tr["LastModifiedDate"])},
		templTicketHeaderField{Name: "Last Activity", Value: dis(tr["LastActivityDate"])},
		templTicketHeaderField{Name: "From", Value: dis(tr["BMCServiceDesk__clientEmail__c"])},
		templTicketHeaderField{Name: "User / UPI", Value: dis(tr["UCL_userid_UPI__c"])},
	}
	ticket.Description = p2h(dis(tr["BMCServiceDesk__incidentDescription__c"]))

	templ := getTicketTemplate()
	err = templ.Execute(w, ticket)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func runTicketServer(args []string) {

	http.HandleFunc("/ticket", ticketServerHandler)

	log.Fatal(http.ListenAndServe("localhost:9090", nil))
}

type templTicket struct {
	Number      string // Sorry
	Summary     string
	Headers     []templTicketHeaderField
	Description template.HTML
	Histories   []templTicketHistory
}

type templTicketHeaderField struct {
	Name  string
	Value string
}

type templTicketHistory struct {
	Headers     []templTicketHeaderField
	Description []string
}

var templateHeader = `<!DOCTYPE html>
<html>
	<head>
		<link
			href="https://unpkg.com/sanitize.css"
			rel="stylesheet"
			/>
		<style>
* {
	font-family: "Source Serif Pro";
}
body {
	background-color: #452b27; 
}
article {
	max-width: 80%;
	margin: 2em;
	padding: 2em;
	align: center;
	display: block;
	margin-left: auto;
	margin-right: auto;
	background-color: wheat;
}

h2 {
	font-style: italic;
}
h2::before {
	color: #442000;
	content: "“"
}
h2::after {
	color: #442000;
	content: "”"
}

table { border: 2px ridge #361d18; }
td { padding: 0.5em; }
td.table-field-key { font-weight: bold; }
tr:nth-child(even) {background-color: #e5cea3;}

.body-text {
}
</style>
</head>
	<body>
		<header>
		</header>
`

var templateFooter = `
	</body>
</html>
`

func getErrorTemplate() *template.Template {
	ts := templateHeader + `<article><h1>Error</h1><h2>{{.}}</h2></article>` + templateFooter
	t, err := template.New("ticket").Parse(ts)
	if err != nil {
		panic(err)
	}
	return t
}

func getTicketTemplate() *template.Template {
	ts := templateHeader + `
		<article>
			<h1>IN:{{.Number}}</h1>
			<h2>{{.Summary}}</h2>
			<section class="ticket-header">
				<table>
						<!-- <th><td></td><td></td></th> -->
					{{range .Headers}}
						<tr><td class="table-field-key">{{.Name}}</td><td>{{.Value}}</td></tr>
					{{end}}
				</table>
			</section>
			<section class="ticket-body">
				<p>
					{{.Description}}
				</p>
			</section>
		</article>
		{{range .Histories}}
			<hr>
			<article>
				<section class="ticket-followup-header">
					<table>
						<!-- <th><td></td><td></td></th> -->
						{{range .Headers}}
							<tr><td class="table-field-key">{{.Name}}</td><td>{{.Value}}</td></tr>
						{{end}}
					</table>
				</section>
				<section class="ticket-followup-body">
					{{.Description}}
				</section>
			</article>
		{{end}}
` + templateFooter

	t, err := template.New("ticket").Parse(ts)
	if err != nil {
		panic(err)
	}
	return t
}

var tsend = `
		{{range .Histories}}
			<article>
				<section class="ticket-followup-header">
					<table>
						<!-- <th><td></td><td></td></th> -->
						{{range .Headers}}
							<tr><td class="table-field-key">{{.Name}}</td><td>{{.Value}}</td></tr>
						{{end}}
					</table>
				</section>
				<section class="ticket-followup-body">
					{{.Description}}
				</section>
			</article>
		{{end}}
		`
