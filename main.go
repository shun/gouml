package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// ########################################################
//
// global values
//
var rootNode Node
var reNode = regexp.MustCompile(`^([ \t]*)?:(.*);`)
var reIf = regexp.MustCompile(`^([ \t]*)?if ([^ ]+) {`)
var reElseIf = regexp.MustCompile(`^([ \t]*)?} elseif ([^ ]+) {`)
var reElse = regexp.MustCompile(`^([ \t]*)?} else {`)
var reCloseBrase = regexp.MustCompile(`^([ \t]*)?}`)

const (
	TypeEnd      = -1
	TypeStart    = 0
	TypeNode     = 1
	TypeDecision = 2
	TypeMerge    = 3
	TypeNote     = 4
)

var nodeList []*Node
var decList []*Node
var mrgList []*Node
var noteList []*Node
var dmynodeList []*Node

// ########################################################
//
// struct
//

type Node struct {
	nodetype int
	label    string
	parent   *Node
	children []*Node
	ex       int
}

// ########################################################
//
// functions
//

func parse(lines []string) {

	rootNode.label = ""
	rootNode.nodetype = TypeStart
	rootNode.parent = nil

	var ret []string
	var node *Node
	var mergeNode *Node
	var wp *Node

	wp = &rootNode
	for i := 0; i < len(lines); i++ {
		ret = reNode.FindStringSubmatch(lines[i])
		if len(ret) > 0 {
			node = new(Node)
			node.nodetype = TypeNode
			node.label = ret[2]
			node.parent = wp
			wp.children = append(wp.children, node)
			wp = node
			nodeList = append(nodeList, node)

			mergeNode = new(Node)
			mergeNode.nodetype = TypeMerge
			mergeNode.label = ""
			mergeNode.parent = nil

			continue
		}

		ret = reIf.FindStringSubmatch(lines[i])
		if len(ret) > 0 {
			node = new(Node)
			node.nodetype = TypeDecision
			node.label = ""
			node.parent = wp
			node.ex = 1
			wp.children = append(wp.children, node)
			wp = node
			decList = append(decList, node)

			node = new(Node)
			node.nodetype = TypeNote
			node.label = ret[2]
			noteList = append(noteList, node)
			mrgList = append(mrgList, mergeNode)

			continue
		}

		ret = reElseIf.FindStringSubmatch(lines[i])
		if len(ret) > 0 {
			wp.children = append(wp.children, mergeNode)
			wp = wp.parent

			node = new(Node)
			node.nodetype = TypeDecision
			node.label = ""
			node.parent = wp
			wp.children = append(wp.children, node)
			wp = node
			decList = append(decList, node)

			node = new(Node)
			node.nodetype = TypeNote
			node.label = ret[2]
			noteList = append(noteList, node)

			continue
		}

		ret = reElse.FindStringSubmatch(lines[i])
		if len(ret) > 0 {
			wp.children = append(wp.children, mergeNode)
			wp = wp.parent

			continue
		}

		ret = reCloseBrase.FindStringSubmatch(lines[i])
		if len(ret) > 0 {
			wp.children = append(wp.children, mergeNode)
			wp = mergeNode
			continue
		}
	}

	node = new(Node)
	node.nodetype = TypeEnd
	wp.children = append(wp.children, node)
	fmt.Println("")
}

func readfile(fin *os.File) []string {
	lines := []string{}
	sc := bufio.NewScanner(fin)

	reComment := regexp.MustCompile(`#.*$`)
	for sc.Scan() {
		line := reComment.ReplaceAllString(strings.TrimSpace(sc.Text()), "")
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}

	return lines
}

func printtree() {
	senum := `
TypeEnd = -1
TypeStart = 0
TypeNode = 1
TypeDecision = 2
TypeMerge = 3
TypeNote = 4`
	fmt.Println(senum)
	dump(&rootNode, 0)
}

func dumplists() {
	fmt.Println("nodeList :")
	for key, val := range nodeList {
		fmt.Printf("\t%d [%s]\n", key, val.label)
	}
	fmt.Println("decList :")
	for key, val := range decList {
		fmt.Printf("\t%d [%s]\n", key, val.label)
	}
	fmt.Println("noteList :")
	for key, val := range noteList {
		fmt.Printf("\t%d [%s]\n", key, val.label)
	}
	fmt.Println("dmynodeList :")
	for key, val := range dmynodeList {
		fmt.Printf("\t%d [%s]\n", key, val.label)
	}
}

func dump(node *Node, indent int) {
	for i := 0; i < indent; i++ {
		fmt.Print("\t")
	}

	fmt.Printf("%d: [%s]\n", node.nodetype, node.label)

	for _, value := range node.children {
		dump(value, indent+1)
	}
}

type EdgeInfo struct {
	count  []int
	output string
}

func getSymbol(node *Node, edgeInfo *EdgeInfo) string {
	sym := ""

	if node.nodetype == TypeStart {
		sym = "st"
	} else if node.nodetype == TypeNode {
		sym = "nd" + strconv.Itoa(edgeInfo.count[TypeNode-1])
	} else if node.nodetype == TypeDecision {
		sym = "dec" + strconv.Itoa(edgeInfo.count[TypeDecision-1])
	} else if node.nodetype == TypeMerge {
		sym = "mrg" + strconv.Itoa(edgeInfo.count[TypeMerge-1])
	} else if node.nodetype == TypeEnd {
		sym = "ed"
	}

	return sym
}

func makeEdgeString(node *Node, edgeInfo *EdgeInfo) {
	sym := getSymbol(node, edgeInfo)

	for idx, child := range node.children {
		if child.nodetype == TypeNode {
			edgeInfo.count[TypeNode-1]++
		} else if child.nodetype == TypeDecision {
			edgeInfo.count[TypeDecision-1]++
		} else {

		}
		tmp := ""
		if node.nodetype == TypeDecision {
			if idx == 0 {
				tmp += `[xlabel = "yes"]`
				if node.ex == 1 {
					edgeInfo.count[TypeMerge-1]++
				}
			} else {
				tmp += `[xlabel = "no"]`
			}
		}
		edgeInfo.output += sym + " -> " + getSymbol(child, edgeInfo) + tmp
		edgeInfo.output += ";\n"
		makeEdgeString(child, edgeInfo)
	}

}

func execute() {
	output := "digraph G {\n"

	output += `graph [rankdir = TD;splines=curved];` + "\n"
	output += `node [ color = "black", fixedsize = true ];` + "\n"
	output += "\n"

	output += `st [label="", shape = circle, style = "filled", fillcolor = "black", height = 0.2, width = 0.2];` + "\n"
	output += `ed [label="", shape = doublecircle, style = "filled", fillcolor = "black", height = 0.2, width = 0.2];` + "\n"
	output += "\n"

	for idx := 0; idx < len(nodeList); idx++ {
		output += fmt.Sprintf("nd%d", idx+1) + ` [label="` + nodeList[idx].label + `", shape = box, style = "filled", fillcolor = "lemonchiffon", fixedsize = false];` + "\n"
	}
	output += "\n"

	for idx := 0; idx < len(decList); idx++ {
		output += fmt.Sprintf("dec%d", idx+1) + `[label="", shape = diamond, style = "filled", fillcolor = "lemonchiffon", height = 0.5, width = 0.5];` + "\n"
	}
	output += "\n"

	for idx := 0; idx < len(mrgList); idx++ {
		output += fmt.Sprintf("mrg%d", idx+1) + `[label="", shape = diamond, style = "filled", fillcolor = "lemonchiffon", height = 0.5, width = 0.5];` + "\n"
	}
	output += "\n"

	for idx := 0; idx < len(noteList); idx++ {
		output += fmt.Sprintf("nt%d", idx+1) + `[label="` + noteList[idx].label + `", shape = note, style = "filled", fillcolor = "lemonchiffon", fixedsize = false];` + "\n"
	}
	output += "\n"

	for idx := 0; idx < len(dmynodeList); idx++ {
		output += fmt.Sprintf("dmynd%d", idx+1) + `[ label="", shape = box, style = "filled", fillcolor = "lemonchiffon", fixedsize = false];` + "\n"
	}
	output += "\n"

	edgeInfo := new(EdgeInfo)
	edgeInfo.output = ""
	edgeInfo.count = []int{0, 0, 0}
	makeEdgeString(&rootNode, edgeInfo)

	output += edgeInfo.output
	for idx := 0; idx < len(noteList); idx++ {
		output += fmt.Sprintf("nt%d -> dec%d[arrowhead = none; style = \"dotted\"];\n", idx+1, idx+1)
	}

	output += "}\n"

	fmt.Println(output)
}

func main() {
	fin, err := os.Open(os.Args[1])
	if err != nil {
		panic("file err")
	}
	lines := readfile(fin)

	parse(lines)
	execute()
	//printtree()
	//	dumplists()

	fin.Close()
}
