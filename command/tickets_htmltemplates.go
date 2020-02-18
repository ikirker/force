package command

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
			<div style="float: right">
				<a target="_blank" href="https://ucl--bmcservicedesk.eu28.visual.force.com/apex/DeepView?id={{.Id}}&showSidebarOnly=true&moduleName=Incident__c">
					<span style="text-shadow: 0px 0px 3px black; font-size: 300%"> 
						✉️ 
					</span>
				</a>
			</div>
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
