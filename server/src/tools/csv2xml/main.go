package main

import (
	"encoding/csv"
	"encoding/xml"
	"flag"
	"logger"
	"os"

	"strings"
)

var (
	csvFile    = flag.String("c", "in.csv", "csv input file")
	xmlFile    = flag.String("x", "out.xml", "xml output file")
	openIndent = flag.Bool("i", false, "show indent")
)

type Doc struct {
	Nodes []*Node `xml:"Node"`
}

type Node struct {
	Name string
	Rows []*Row `xml:"Row"`
}

type Row struct {
	Cols []*Col `xml:"Col"`
}

type Col struct {
	Key   string `xml:",attr"`
	Value string `xml:",attr"`
}

func main() {
	flag.Parse()
	logger.Info("process %s to %s..", *csvFile, *xmlFile)

	infile, err := os.Open(*csvFile)

	if err != nil {
		logger.Fatal("Error On open %s : %s", *csvFile, err.Error())
	}

	defer infile.Close()

	outfile, err := os.Create(*xmlFile)

	if err != nil {
		logger.Fatal("Error On open %s : %s", *csvFile, err.Error())
	}

	defer outfile.Close()

	r := csv.NewReader(infile)
	r.TrailingComma = true

	out, err := r.ReadAll()

	if err != nil {
		logger.Fatal("Error On read %s : %s", *csvFile, err.Error())
	}

	if len(out) <= 2 {
		logger.Fatal("Empty or wrong format %s ", *csvFile)
	}

	for col := 0; col < len(out[1]); col++ {
		switch strings.ToLower(out[1][col]) {
		case "string":

		case "int":

		case "boolean":

		default:
			logger.Fatal("wrong type %s for col %d ", out[1][col], col)
		}
	}

	if strings.ToLower(out[1][0]) != "string" {
		logger.Fatal("wrong type %s, col 1 must be string ", out[1][0])
	}

	doc := &Doc{
		Nodes: make([]*Node, 0),
	}

	var node *Node
	var node_row *Row

	for row := 2; row < len(out); row++ {
		for col := 0; col < len(out[row]); col++ {
			svalue := strings.TrimSpace(out[row][col])

			if col == 0 && svalue != "" {

				node = &Node{svalue, make([]*Row, 0)}
				doc.Nodes = append(doc.Nodes, node)
			}

			if col == 0 {

				node_row = &Row{make([]*Col, 0)}
				node.Rows = append(node.Rows, node_row)

			} else {
				if svalue != "" {
					node_row.Cols = append(node_row.Cols, &Col{out[0][col], svalue})
				}
			}
		}
	}

	if !*openIndent {
		output, err := xml.Marshal(doc)

		if err != nil {
			logger.Fatal("Marshl Error %s ", err.Error())
		}
		outfile.Write(output)

	} else {
		output, err := xml.MarshalIndent(doc, "  ", "    ")

		if err != nil {
			logger.Fatal("Marshl Error %s ", err.Error())
		}

		outfile.Write(output)
	}

	logger.Info("End..")
}
