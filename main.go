package main

import (
	"bufio"
	"fmt"
	"log"
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
var reLoop = regexp.MustCompile(`^([ \t]*)?Loop ([^ ]+) {`)
var reElseIf = regexp.MustCompile(`^([ \t]*)?} elseif ([^ ]+) {`)
var reElse = regexp.MustCompile(`^([ \t]*)?} else {`)
var reCloseBrase = regexp.MustCompile(`^([ \t]*)?}`)

const (
	DecisionByIf   = 1
	DecisionByLoop = 1
	MergeByLoop    = 2
)

const (
	TypeEnd      = -1
	TypeStart    = 0
	TypeNode     = 1
	TypeDecision = 2
	TypeMerge    = 3
	TypeNote     = 4
	TypeLoop     = 5
)

var nodeList []*Node
var decList []*Node
var mrgList []*Node
var noteList []*Node
var dmynodeList []*Node
var lpList []*Node
var mrgCurIdx int

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
	if lines[0] == "type=act" {
		parseActivityDiagram(lines)
	} else if lines[0] == "type=seq" {
		parseSequenceDiagram(lines)
	} else {
		// not support yet
	}
}

func parseSequenceDiagram(lines []string) {
}

func parseActivityDiagram(lines []string) {

	rootNode.label = ""
	rootNode.nodetype = TypeStart
	rootNode.parent = nil

	var ret []string
	var node *Node
	var mergeNode *Node
	var wp *Node

	sp := 0

	wp = &rootNode
	for i := 1; i < len(lines); i++ {
		ret = reNode.FindStringSubmatch(lines[i])
		if len(ret) > 0 {
			//log.Println("### Node")
			node = new(Node)
			node.nodetype = TypeNode
			node.label = ret[2]
			node.parent = wp
			wp.children = append(wp.children, node)
			//log.Println("\t\tappend node")
			wp = node
			nodeList = append(nodeList, node)

			continue
		}

		ret = reLoop.FindStringSubmatch(lines[i])
		if len(ret) > 0 {
			//log.Println("### Loop")
			node = new(Node)
			node.nodetype = TypeLoop
			node.label = ""
			node.parent = wp
			node.ex = DecisionByLoop
			wp.children = append(wp.children, node)
			//log.Println("\t\tappend loop")
			wp = node
			lpList = append(lpList, node)

			mergeNode = new(Node)
			mergeNode.nodetype = TypeMerge
			mergeNode.label = ""
			mergeNode.parent = node
			mergeNode.ex = MergeByLoop
			//node.children = append(node.children, mergeNode)
			//log.Println("\t\tappend merge")
			mrgList = append(mrgList, mergeNode)
			mrgCurIdx++

			node = new(Node)
			node.nodetype = TypeNote
			node.label = ret[2]
			node.parent = wp
			noteList = append(noteList, node)

			continue
		}

		ret = reIf.FindStringSubmatch(lines[i])
		if len(ret) > 0 {
			//log.Println("### If")
			node = new(Node)
			node.nodetype = TypeDecision
			node.label = ""
			node.parent = wp
			node.ex = DecisionByIf
			wp.children = append(wp.children, node)
			//log.Println("\t\tappend if")
			wp = node
			decList = append(decList, node)

			mergeNode = new(Node)
			mergeNode.nodetype = TypeMerge
			mergeNode.label = ""
			mergeNode.parent = node
			//node.children = append(node.children, mergeNode)
			//log.Println("\t\tappend merge")
			mrgList = append(mrgList, mergeNode)
			mrgCurIdx++

			node = new(Node)
			node.nodetype = TypeNote
			node.label = ret[2]
			node.parent = wp
			noteList = append(noteList, node)

			continue
		}

		ret = reElseIf.FindStringSubmatch(lines[i])
		if len(ret) > 0 {
			//log.Println("### ElseIf")
			//wp = wp.parent
			wp.children = append(wp.children, mergeNode)
			//log.Println("\t\tappend merge")

			node = new(Node)
			node.nodetype = TypeDecision
			node.label = ""
			node.parent = wp.parent
			wp.parent.children = append(wp.parent.children, node)
			//log.Println("\t\tappend elseif")
			wp = node
			decList = append(decList, node)

			node = new(Node)
			node.nodetype = TypeNote
			node.label = ret[2]
			node.parent = wp
			noteList = append(noteList, node)

			continue
		}

		ret = reElse.FindStringSubmatch(lines[i])
		if len(ret) > 0 {
			//log.Println("### Else")
			wp.children = append(wp.children, mergeNode)
			//log.Println("\t\tappend else")
			wp = wp.parent

			continue
		}

		ret = reCloseBrase.FindStringSubmatch(lines[i])
		if len(ret) > 0 {
			//log.Println("### CloseBrase")
			wp.children = append(wp.children, mergeNode)
			//log.Println("\t\tappend merge")
			wp = mergeNode
			mrgCurIdx--
			//log.Printf("\t\tsp : %d, mrgCurIdx : %d\n", sp, mrgCurIdx)
			if mrgCurIdx < 0 {
				sp = len(mrgList)
				//log.Printf("\t\tsp : %d, mrgCurIdx : %d\n", sp, mrgCurIdx)
				mergeNode = nil
			} else {
				mergeNode = mrgList[sp+mrgCurIdx]
			}
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
	log.Println(senum)
	dump(&rootNode, 0)
}

func dumplists() {
	log.Println("nodeList :")
	for key, val := range nodeList {
		log.Printf("\t%d [%s]\n", key, val.label)
	}
	log.Println("decList :")
	for key, val := range decList {
		log.Printf("\t%d [%s]\n", key, val.label)
	}
	log.Println("noteList :")
	for key, val := range noteList {
		log.Printf("\t%d [%s]\n", key, val.label)
	}
	log.Println("dmynodeList :")
	for key, val := range dmynodeList {
		log.Printf("\t%d [%s]\n", key, val.label)
	}
}

var isMainFlow = true

func dump(node *Node, indent int) {
	idt := ""
	if node.nodetype == TypeEnd {
		isMainFlow = false
		return
	}

	for i := 0; i < indent; i++ {
		idt += "\t"
	}

	log.Printf("%s%d: [%s] %p\n", idt, node.nodetype, node.label, node)
	if (isMainFlow == false) && (node.nodetype == TypeMerge) {
		return
	}

	for _, value := range node.children {
		dump(value, indent+1)
	}
	log.Println("")
}

type EdgeInfo struct {
	count      []int
	output     string
	isMainFlow bool
}
type any interface{}

//func getIndex(lst []Any, p Any) int {
func getIndex(lst []*Node, p *Node) int {
	for i, v := range lst {
		if v == p {
			return i
		}
	}

	return -1
}

func getSymbol(node *Node, edgeInfo *EdgeInfo) string {
	sym := ""

	if node.nodetype == TypeStart {
		sym = "st"
	} else if node.nodetype == TypeNode {
		//sym = "nd" + strconv.Itoa(edgeInfo.count[TypeNode-1])
		i := getIndex(nodeList, node)
		sym = "nd" + strconv.Itoa(i+1)
	} else if node.nodetype == TypeDecision {
		//sym = "dec" + strconv.Itoa(edgeInfo.count[TypeDecision-1])
		i := getIndex(decList, node)
		sym = "dec" + strconv.Itoa(i+1)
	} else if node.nodetype == TypeMerge {
		//sym = "mrg" + strconv.Itoa(edgeInfo.count[TypeMerge-1])
		i := getIndex(mrgList, node)
		sym = "mrg" + strconv.Itoa(i+1)
	} else if node.nodetype == TypeEnd {
		sym = "ed"
	} else if node.nodetype == TypeLoop {
		//sym = "lp" + strconv.Itoa(edgeInfo.count[TypeLoop-1])
		i := getIndex(lpList, node)
		sym = "lp" + strconv.Itoa(i+1)
	}

	return sym
}

func makeEdgeString(node *Node, edgeInfo *EdgeInfo) {
	sym := getSymbol(node, edgeInfo)

	if node.nodetype == TypeEnd {
		edgeInfo.isMainFlow = false
		return
	}

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
		if node.ex == MergeByLoop {
			i := getIndex(lpList, node.parent)
			//edgeInfo.output += sym + " -> " + "lp" + strconv.Itoa(edgeInfo.count[TypeLoop-1]) + ";\n"
			edgeInfo.output += sym + " -> " + "lp" + strconv.Itoa(i+1) + ";\n"
			edgeInfo.count[TypeLoop-1]++
		}
		edgeInfo.output += sym + " -> " + getSymbol(child, edgeInfo) + tmp
		edgeInfo.output += ";\n"

		if (edgeInfo.isMainFlow == false) && (child.nodetype == TypeMerge) && (child.ex != MergeByLoop) {
			return
		}
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

	for idx := 0; idx < len(lpList); idx++ {
		output += fmt.Sprintf("lp%d", idx+1) + `[label="", shape = diamond, style = "filled", fillcolor = "lemonchiffon", height = 0.5, width = 0.5];` + "\n"
	}
	output += "\n"

	edgeInfo := new(EdgeInfo)
	edgeInfo.output = ""
	edgeInfo.count = []int{0, 0, 0, 0, 0}
	edgeInfo.isMainFlow = true
	makeEdgeString(&rootNode, edgeInfo)

	output += edgeInfo.output
	for idx := 0; idx < len(noteList); idx++ {
		//output += fmt.Sprintf("nt%d -> dec%d[arrowhead = none; style = \"dotted\"];\n", idx+1, idx+1)
		if noteList[idx].parent.nodetype == TypeLoop {
			i := getIndex(noteList, noteList[idx])
			j := getIndex(lpList, noteList[idx].parent)
			output += fmt.Sprintf("nt%d -> lp%d[arrowhead = none; style = \"dotted\"];\n", i+1, j+1)
		} else {
			i := getIndex(noteList, noteList[idx])
			j := getIndex(decList, noteList[idx].parent)
			output += fmt.Sprintf("nt%d -> dec%d[arrowhead = none; style = \"dotted\"];\n", i+1, j+1)
		}
	}

    output += makeRank(decList, TypeDecision)
    output += makeRank(lpList, TypeLoop)
//    for idx := 0; idx < len(decList); idx++ {
//        dec := decList[idx]
//
//        cnt := 0
//        tmp := "{rank = same; dec1; "
//        for i, v := range dec.children {
//            if v.nodetype == TypeLoop {
//                tmp += fmt.Sprintf("lp%d; ", getIndex(lpList, v)+1)
//                cnt++
//            } else {
//                tmp += fmt.Sprintf("dec%d; ", getIndex(decList, v)+1)
//                cnt++
//            }
//            tmp += "}\n"
//        }
//        if cnt > 0 {
//            output += tmp
//        }
//    }
//
//    for idx :=0; idx < len(lpList); idx++ {
//        lp := lpList[idx]
//
//        cnt := 0
//        tmp := "{rank = same; lp1; "
//        for i, v := range lp.children {
//            if v.nodetype == TypeLoop {
//                tmp += fmt.Sprintf("lp%d; ", getIndex(lpList, v)+1)
//                cnt++
//            } else {
//                tmp += fmt.Sprintf("dec%d; ", getIndex(decList, v)+1)
//                cnt++
//            }
//            tmp += "}\n"
//        }
//        if cnt > 0 {
//            output += tmp
//        }
//    }
	output += "\n}\n"

	fmt.Println(output)
}

func makeRank(lst []*Node, nodetype int) string {
    ret := ""
    for idx :=0; idx < len(lst); idx++ {
        node := lst[idx]

        cnt := 0
        tmp := ""
        if nodetype == TypeLoop {
            tmp = fmt.Sprintf("{rank = same; lp%d; ", getIndex(lst, node)+1)
        } else if nodetype == TypeDecision {
            tmp = fmt.Sprintf("{rank = same; dec%d; ", getIndex(lst, node)+1)
        } else {
            continue
        }
        for _, v := range node.children {
            if v.nodetype == TypeLoop {
                tmp += fmt.Sprintf("lp%d; ", getIndex(lpList, v)+1)
                cnt++
            } else if v.nodetype == TypeDecision {
                tmp += fmt.Sprintf("dec%d; ", getIndex(decList, v)+1)
                cnt++
            } else {
                // nop
            }
        }
        tmp += "}\n"
        if cnt > 0 {
           ret += tmp
        }
    }
    return ret 
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	mrgCurIdx = -1
	path := ""
	if len(os.Args) < 2 {
		path = "example/test.guml"
	} else {
		path = os.Args[1]
	}
	//fin, err := os.Open(os.Args[1])
	fin, err := os.Open(path)
	if err != nil {
		panic("file err")
	}
	lines := readfile(fin)

	parse(lines)
	execute()
	//printtree()
	//dumplists()

	fin.Close()
}
