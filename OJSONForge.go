// ============================================================
//
//	OJSONForge — JSON Query / Transform / Validate / Pretty
//	Author  : khaled.s.haddad
//	Website : owlsec.org
//	Language: Go (stdlib only, CGO disabled, static build) — OWL suite
//
//	Pretty-print, minify, validate (with line/column error position),
//	query by dotted/bracket path, and profile the structure of any
//	JSON document — from a file, inline text, or standard input.
//
// ============================================================
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

// ---- OWL UI KIT (cyan "owl" house style) — stdlib only. Shared across the suite ----
const (
	R   = "\033[91m"
	GRN = "\033[92m"
	YLW = "\033[93m"
	CY  = "\033[96m"
	W   = "\033[97m"
	B   = "\033[1m"
	X   = "\033[0m"
	MG  = "\033[95m"
	DIM = "\033[2m"
	OR  = "\033[38;5;208m"
	GY  = "\033[90m"
)

var noColor = os.Getenv("NO_COLOR") != "" || !isatty()

func isatty() bool { fi, _ := os.Stdout.Stat(); return (fi.Mode() & os.ModeCharDevice) != 0 }
func col(s string, c ...string) string {
	if noColor {
		return s
	}
	return strings.Join(c, "") + s + X
}
func clear() {
	if !noColor {
		fmt.Print("\033[H\033[2J")
	}
}
func rule()            { fmt.Println(col(strings.Repeat("─", 66), CY)) }
func okMsg(s string)   { fmt.Println("  " + col("[✔] ", GRN+B) + col(s, W)) }
func errMsg(s string)  { fmt.Println("  " + col("[✘] ", R+B) + col(s, R)) }
func warnMsg(s string) { fmt.Println("  " + col("[!] ", OR+B) + col(s, YLW)) }
func info(s string)    { fmt.Println("  " + col("[*] ", CY) + col(s, DIM+W)) }

func header(title string) {
	fmt.Println()
	rule()
	fmt.Println("  " + col("⬡  "+title, B+CY))
	rule()
	fmt.Println()
}

var stdin = bufio.NewReader(os.Stdin)

func prompt(label string) string {
	fmt.Print("  " + col("[»] ", YLW) + col(label+": ", W))
	s, err := stdin.ReadString('\n')
	if err != nil && s == "" {
		fmt.Println()
		os.Exit(0)
	}
	return strings.TrimSpace(s)
}
func pause() {
	fmt.Print(col("\n  [↩] Press Enter to continue...", YLW))
	stdin.ReadString('\n')
}

func stripANSI(s string) string {
	out := []rune{}
	esc := false
	for _, r := range s {
		if r == '\033' {
			esc = true
			continue
		}
		if esc {
			if r == 'm' {
				esc = false
			}
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
func visw(s string) int { return len([]rune(stripANSI(s))) }
func pad(s string, n int) string {
	if visw(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-visw(s))
}
func table(headers []string, rows [][]string) {
	w := make([]int, len(headers))
	for i, h := range headers {
		w[i] = visw(h)
	}
	for _, r := range rows {
		for i, c := range r {
			if i < len(w) && visw(c) > w[i] {
				w[i] = visw(c)
			}
		}
	}
	var hb strings.Builder
	for i, h := range headers {
		hb.WriteString(col(pad(h, w[i]), B+CY))
		if i < len(headers)-1 {
			hb.WriteString(col(" │ ", DIM))
		}
	}
	fmt.Println("  " + hb.String())
	seg := make([]string, len(w))
	for i := range w {
		seg[i] = strings.Repeat("─", w[i])
	}
	fmt.Println("  " + col(strings.Join(seg, "─┼─"), DIM))
	for _, r := range rows {
		var rb strings.Builder
		for i := range headers {
			v := ""
			if i < len(r) {
				v = r[i]
			}
			rb.WriteString(col(pad(v, w[i]), W))
			if i < len(headers)-1 {
				rb.WriteString(col(" │ ", DIM))
			}
		}
		fmt.Println("  " + rb.String())
	}
}

const BANNER = `
 ██████╗      ██╗███████╗ ██████╗ ███╗   ██╗███████╗ ██████╗ ██████╗  ██████╗ ███████╗
██╔═══██╗     ██║██╔════╝██╔═══██╗████╗  ██║██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██╔════╝
██║   ██║     ██║███████╗██║   ██║██╔██╗ ██║█████╗  ██║   ██║██████╔╝██║  ███╗█████╗
██║   ██║██   ██║╚════██║██║   ██║██║╚██╗██║██╔══╝  ██║   ██║██╔══██╗██║   ██║██╔══╝
╚██████╔╝╚█████╔╝███████║╚██████╔╝██║ ╚████║██║     ╚██████╔╝██║  ██║╚██████╔╝███████╗
 ╚═════╝  ╚════╝ ╚══════╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝
`

const (
	APPNAME = "OJSONForge"
	SUITE   = "Owl Data Suite"
	SUBLINE = "Pretty · Minify · Validate · Query · Stats · Diff · Flatten · Schema · JSONL"
	VERSION = "1.1"
	AUTHOR  = "khaled.s.haddad"
	WEBSITE = "owlsec.org"
)

func banner() {
	fmt.Println(col(BANNER, CY))
	fmt.Println("  " + col(SUITE+"  ", B+MG) + col("v"+VERSION, YLW))
	fmt.Println("  " + col(SUBLINE, W))
	fmt.Println()
	fmt.Println("  " + col("Developer : ", DIM) + col(AUTHOR, CY) +
		col("   |   Website : ", DIM) + col(WEBSITE, CY))
	rule()
	fmt.Println()
}

// ---------------------------------------------------------------------------
// Core logic
// ---------------------------------------------------------------------------

// getInput returns raw JSON bytes from a file path or inline text.
func getInput(label string) ([]byte, bool) {
	in := strings.TrimSpace(prompt(label + " (file path OR inline JSON)"))
	if in == "" {
		return nil, false
	}
	if !strings.HasPrefix(in, "{") && !strings.HasPrefix(in, "[") {
		if b, err := os.ReadFile(in); err == nil {
			return b, true
		}
	}
	return []byte(in), true
}

// prettyBytes renders v with 2-space indent, no HTML escaping.
func prettyBytes(v interface{}) []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	enc.Encode(v)
	return bytes.TrimRight(buf.Bytes(), "\n")
}

func colorPrintJSON(s string) {
	for _, ln := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		trimmed := strings.TrimSpace(ln)
		indent := ln[:len(ln)-len(strings.TrimLeft(ln, " "))]
		if i := strings.Index(trimmed, "\":"); i >= 0 && strings.HasPrefix(trimmed, "\"") {
			key := trimmed[:i+1]
			rest := trimmed[i+1:]
			fmt.Printf("  %s%s%s\n", indent, col(key, CY), col(rest, W))
		} else {
			fmt.Printf("  %s%s\n", indent, col(trimmed, W))
		}
	}
}

func offsetToLineCol(b []byte, off int) (int, int) {
	line, colm := 1, 1
	for i := 0; i < off && i < len(b); i++ {
		if b[i] == '\n' {
			line++
			colm = 1
		} else {
			colm++
		}
	}
	return line, colm
}

// validate parses raw JSON and returns (type, depth, nodes) or a positioned error.
func validate(raw []byte) (string, int, int, error) {
	var v interface{}
	dec := json.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&v); err != nil {
		if se, ok := err.(*json.SyntaxError); ok {
			line, colm := offsetToLineCol(raw, int(se.Offset))
			return "", 0, 0, fmt.Errorf("line %d col %d (byte %d): %s", line, colm, se.Offset, se.Error())
		}
		return "", 0, 0, err
	}
	return typeName(v), depth(v), countNodes(v), nil
}

func navigate(v interface{}, path string) (interface{}, error) {
	path = strings.TrimPrefix(strings.TrimSpace(path), ".")
	if path == "" {
		return v, nil
	}
	path = strings.ReplaceAll(path, "[", ".")
	path = strings.ReplaceAll(path, "]", "")
	parts := strings.Split(path, ".")
	cur := v
	for _, p := range parts {
		if p == "" {
			continue
		}
		switch node := cur.(type) {
		case map[string]interface{}:
			nv, ok := node[p]
			if !ok {
				return nil, fmt.Errorf("key %q not found", p)
			}
			cur = nv
		case []interface{}:
			idx, err := strconv.Atoi(p)
			if err != nil {
				return nil, fmt.Errorf("expected array index, got %q", p)
			}
			if idx < 0 || idx >= len(node) {
				return nil, fmt.Errorf("index %d out of range (len %d)", idx, len(node))
			}
			cur = node[idx]
		default:
			return nil, fmt.Errorf("cannot descend into %s at %q", typeName(cur), p)
		}
	}
	return cur, nil
}

func typeName(v interface{}) string {
	switch v.(type) {
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "bool"
	case nil:
		return "null"
	}
	return "unknown"
}

func depth(v interface{}) int {
	switch node := v.(type) {
	case map[string]interface{}:
		m := 0
		for _, c := range node {
			if d := depth(c); d > m {
				m = d
			}
		}
		return m + 1
	case []interface{}:
		m := 0
		for _, c := range node {
			if d := depth(c); d > m {
				m = d
			}
		}
		return m + 1
	}
	return 1
}

func countNodes(v interface{}) int {
	switch node := v.(type) {
	case map[string]interface{}:
		n := 1
		for _, c := range node {
			n += countNodes(c)
		}
		return n
	case []interface{}:
		n := 1
		for _, c := range node {
			n += countNodes(c)
		}
		return n
	}
	return 1
}

func tallyTypes(v interface{}, tc map[string]int) {
	tc[typeName(v)]++
	switch node := v.(type) {
	case map[string]interface{}:
		for _, c := range node {
			tallyTypes(c, tc)
		}
	case []interface{}:
		for _, c := range node {
			tallyTypes(c, tc)
		}
	}
}

// ---------------------------------------------------------------------------
// v1.1 — diff, flatten/unflatten, schema inference, JSONL
// ---------------------------------------------------------------------------

// diffEntry is one change between two documents at a given path.
type diffEntry struct {
	Path string
	Kind string // added | removed | changed | type
	Old  string
	New  string
}

func scalarStr(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return "null"
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(t, 'g', -1, 64)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func joinPath(base, key string) string {
	if base == "" {
		return key
	}
	return base + "." + key
}

// diffJSON walks two decoded documents and reports structural differences.
func diffJSON(a, b interface{}, path string, out *[]diffEntry) {
	ta, tb := typeName(a), typeName(b)
	if ta != tb {
		*out = append(*out, diffEntry{path, "type", ta, tb})
		return
	}
	switch na := a.(type) {
	case map[string]interface{}:
		nb := b.(map[string]interface{})
		keys := map[string]bool{}
		for k := range na {
			keys[k] = true
		}
		for k := range nb {
			keys[k] = true
		}
		var ks []string
		for k := range keys {
			ks = append(ks, k)
		}
		sort.Strings(ks)
		for _, k := range ks {
			va, oka := na[k]
			vb, okb := nb[k]
			p := joinPath(path, k)
			switch {
			case oka && !okb:
				*out = append(*out, diffEntry{p, "removed", scalarStr(va), ""})
			case !oka && okb:
				*out = append(*out, diffEntry{p, "added", "", scalarStr(vb)})
			default:
				diffJSON(va, vb, p, out)
			}
		}
	case []interface{}:
		nb := b.([]interface{})
		max := len(na)
		if len(nb) > max {
			max = len(nb)
		}
		for i := 0; i < max; i++ {
			p := joinPath(path, strconv.Itoa(i))
			switch {
			case i >= len(na):
				*out = append(*out, diffEntry{p, "added", "", scalarStr(nb[i])})
			case i >= len(nb):
				*out = append(*out, diffEntry{p, "removed", scalarStr(na[i]), ""})
			default:
				diffJSON(na[i], nb[i], p, out)
			}
		}
	default:
		if scalarStr(a) != scalarStr(b) {
			*out = append(*out, diffEntry{path, "changed", scalarStr(a), scalarStr(b)})
		}
	}
}

// flatten turns a nested document into dotted-key → scalar pairs.
func flatten(v interface{}, path string, out map[string]interface{}) {
	switch node := v.(type) {
	case map[string]interface{}:
		if len(node) == 0 {
			out[path] = map[string]interface{}{}
			return
		}
		for k, c := range node {
			flatten(c, joinPath(path, k), out)
		}
	case []interface{}:
		if len(node) == 0 {
			out[path] = []interface{}{}
			return
		}
		for i, c := range node {
			flatten(c, joinPath(path, strconv.Itoa(i)), out)
		}
	default:
		out[path] = v
	}
}

// unflatten rebuilds a nested document from dotted-key pairs. Numeric segments
// become array indices.
func unflatten(flat map[string]interface{}) interface{} {
	var root interface{}
	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		root = setPath(root, strings.Split(k, "."), flat[k])
	}
	if root == nil {
		root = map[string]interface{}{}
	}
	return root
}

func isIndex(s string) (int, bool) {
	n, err := strconv.Atoi(s)
	return n, err == nil && n >= 0
}

func setPath(cur interface{}, segs []string, val interface{}) interface{} {
	if len(segs) == 0 {
		return val
	}
	seg := segs[0]
	if idx, ok := isIndex(seg); ok {
		arr, _ := cur.([]interface{})
		for len(arr) <= idx {
			arr = append(arr, nil)
		}
		arr[idx] = setPath(arr[idx], segs[1:], val)
		return arr
	}
	m, _ := cur.(map[string]interface{})
	if m == nil {
		m = map[string]interface{}{}
	}
	m[seg] = setPath(m[seg], segs[1:], val)
	return m
}

// inferSchema produces a compact type sketch of a document.
func inferSchema(v interface{}) interface{} {
	switch node := v.(type) {
	case map[string]interface{}:
		out := map[string]interface{}{}
		for k, c := range node {
			out[k] = inferSchema(c)
		}
		return out
	case []interface{}:
		if len(node) == 0 {
			return []interface{}{"any"}
		}
		// merge element schemas; if homogeneous, one entry, else list distinct.
		seen := map[string]interface{}{}
		var order []string
		for _, c := range node {
			sc := inferSchema(c)
			b, _ := json.Marshal(sc)
			if _, ok := seen[string(b)]; !ok {
				seen[string(b)] = sc
				order = append(order, string(b))
			}
		}
		if len(order) == 1 {
			return []interface{}{seen[order[0]]}
		}
		var arr []interface{}
		for _, k := range order {
			arr = append(arr, seen[k])
		}
		return arr
	default:
		return typeName(v)
	}
}

// parseJSONL decodes newline-delimited JSON; returns records and the first error.
func parseJSONL(raw []byte) ([]interface{}, int, error) {
	var recs []interface{}
	lineNo := 0
	for _, ln := range strings.Split(string(raw), "\n") {
		s := strings.TrimSpace(ln)
		lineNo++
		if s == "" {
			continue
		}
		var v interface{}
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			return recs, lineNo, fmt.Errorf("line %d: %v", lineNo, err)
		}
		recs = append(recs, v)
	}
	return recs, 0, nil
}

// ---------------------------------------------------------------------------
// Interactive handlers
// ---------------------------------------------------------------------------

func maybeWrite(b []byte) {
	if out := strings.TrimSpace(prompt("write to file? (path or blank)")); out != "" {
		if err := os.WriteFile(out, b, 0644); err != nil {
			errMsg(err.Error())
		} else {
			okMsg("Wrote " + out)
		}
	}
}

func doPretty() {
	raw, ok := getInput("source")
	if !ok {
		return
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		errMsg("invalid JSON: " + err.Error())
		pause()
		return
	}
	out := prettyBytes(v)
	clear()
	banner()
	header("Pretty-Printed JSON")
	colorPrintJSON(string(out))
	fmt.Println()
	maybeWrite(append(out, '\n'))
	pause()
}

func doMinify() {
	raw, ok := getInput("source")
	if !ok {
		return
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		errMsg("invalid JSON: " + err.Error())
		pause()
		return
	}
	clear()
	banner()
	header("Minify / Compact")
	saved := len(raw) - buf.Len()
	info(fmt.Sprintf("%d → %d bytes  (saved %d, %.1f%%)", len(raw), buf.Len(), saved, 100*float64(saved)/float64(max1(len(raw)))))
	fmt.Println()
	fmt.Println("  " + col(buf.String(), W))
	fmt.Println()
	maybeWrite(buf.Bytes())
	pause()
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

func doValidate() {
	raw, ok := getInput("source")
	if !ok {
		return
	}
	clear()
	banner()
	header("Validate")
	t, d, n, err := validate(raw)
	if err != nil {
		errMsg("INVALID JSON")
		info("Syntax error at " + err.Error())
		pause()
		return
	}
	okMsg("Valid JSON.")
	info("Root type: " + t)
	info(fmt.Sprintf("Max depth: %d", d))
	info(fmt.Sprintf("Total nodes: %d", n))
	pause()
}

func doQuery() {
	raw, ok := getInput("source")
	if !ok {
		return
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		errMsg("invalid JSON: " + err.Error())
		pause()
		return
	}
	path := prompt("path (e.g. users.0.name or .items[2].id)")
	res, err := navigate(v, path)
	clear()
	banner()
	header("Query by Path")
	if err != nil {
		errMsg(err.Error())
		pause()
		return
	}
	info("Path: " + path + "   →   " + typeName(res))
	fmt.Println()
	colorPrintJSON(string(prettyBytes(res)))
	pause()
}

func doStats() {
	raw, ok := getInput("source")
	if !ok {
		return
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		errMsg("invalid JSON: " + err.Error())
		pause()
		return
	}
	clear()
	banner()
	header("Structure Statistics")
	tc := map[string]int{}
	tallyTypes(v, tc)
	var keys []string
	for k := range tc {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rows := [][]string{}
	for _, k := range keys {
		rows = append(rows, []string{k, strconv.Itoa(tc[k])})
	}
	table([]string{"Type", "Count"}, rows)
	fmt.Println()
	info(fmt.Sprintf("Max depth: %d   Total nodes: %d", depth(v), countNodes(v)))
	if m, ok := v.(map[string]interface{}); ok {
		var ks []string
		for k := range m {
			ks = append(ks, k)
		}
		sort.Strings(ks)
		info("Top-level keys: " + strings.Join(ks, ", "))
	}
	if a, ok := v.([]interface{}); ok {
		info(fmt.Sprintf("Root is an array of %d element(s)", len(a)))
	}
	pause()
}

func parseInput(label string) (interface{}, []byte, bool) {
	raw, ok := getInput(label)
	if !ok {
		return nil, nil, false
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		errMsg("invalid JSON: " + err.Error())
		pause()
		return nil, nil, false
	}
	return v, raw, true
}

func doDiff() {
	a, _, ok := parseInput("document A source")
	if !ok {
		return
	}
	b, _, ok := parseInput("document B source")
	if !ok {
		return
	}
	clear()
	banner()
	header("Diff A → B")
	var diffs []diffEntry
	diffJSON(a, b, "", &diffs)
	if len(diffs) == 0 {
		okMsg("documents are structurally identical")
		pause()
		return
	}
	rows := [][]string{}
	for _, d := range diffs {
		v := d.New
		if d.Kind == "removed" {
			v = d.Old
		} else if d.Kind == "changed" || d.Kind == "type" {
			v = d.Old + " → " + d.New
		}
		rows = append(rows, []string{d.Kind, d.Path, v})
	}
	table([]string{"Change", "Path", "Value"}, rows)
	fmt.Println()
	info(fmt.Sprintf("%d difference(s)", len(diffs)))
	pause()
}

func doFlatten() {
	v, _, ok := parseInput("source")
	if !ok {
		return
	}
	clear()
	banner()
	header("Flatten → dotted keys")
	flat := map[string]interface{}{}
	flatten(v, "", flat)
	out := prettyBytes(flat)
	colorPrintJSON(string(out))
	fmt.Println()
	maybeWrite(append(out, '\n'))
	pause()
}

func doUnflatten() {
	v, _, ok := parseInput("flattened source (dotted-key object)")
	if !ok {
		return
	}
	m, isObj := v.(map[string]interface{})
	clear()
	banner()
	header("Unflatten → nested")
	if !isObj {
		errMsg("input must be a flat object of dotted keys")
		pause()
		return
	}
	out := prettyBytes(unflatten(m))
	colorPrintJSON(string(out))
	fmt.Println()
	maybeWrite(append(out, '\n'))
	pause()
}

func doSchema() {
	v, _, ok := parseInput("source")
	if !ok {
		return
	}
	clear()
	banner()
	header("Inferred Schema (type sketch)")
	out := prettyBytes(inferSchema(v))
	colorPrintJSON(string(out))
	fmt.Println()
	maybeWrite(append(out, '\n'))
	pause()
}

func doJSONL() {
	raw, ok := getInput("JSONL/ndjson source")
	if !ok {
		return
	}
	clear()
	banner()
	header("JSONL / ndjson")
	recs, badLine, err := parseJSONL(raw)
	if err != nil {
		errMsg("invalid at " + err.Error())
		pause()
		return
	}
	okMsg(fmt.Sprintf("%d valid record(s), 0 errors", len(recs)))
	_ = badLine
	tc := map[string]int{}
	for _, r := range recs {
		tc[typeName(r)]++
	}
	var keys []string
	for k := range tc {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rows := [][]string{}
	for _, k := range keys {
		rows = append(rows, []string{k, strconv.Itoa(tc[k])})
	}
	fmt.Println()
	table([]string{"Record type", "Count"}, rows)
	pause()
}

// ---------------------------------------------------------------------------
// Interactive menu
// ---------------------------------------------------------------------------

func menu(title string, items [][2]string) string {
	clear()
	banner()
	fmt.Println("  " + col("⬡  "+title, B+CY))
	fmt.Println()
	for i, it := range items {
		fmt.Println("  " + col(pad(fmt.Sprintf("[%d]", i+1), 4), CY+B) + "  " +
			col(pad(it[0], 24), W) + col("─ "+it[1], DIM))
	}
	fmt.Println()
	fmt.Println("  " + col(pad("[H]", 4), YLW+B) + "  " + col(pad("Help & Usage", 24), W) + col("─ How to use, every option, CLI flags & examples", DIM))
	fmt.Println("  " + col(pad("[0]", 4), R+B) + "  " + col(pad("Exit", 24), W) + col("─ Quit "+APPNAME, DIM))
	fmt.Println()
	rule()
	return prompt("Select option")
}

func interactive() {
	for {
		c := menu("Main Menu", [][2]string{
			{"Pretty-print", "indent + colourise"},
			{"Minify / compact", "strip whitespace"},
			{"Validate", "with line/column error position"},
			{"Query by path", "a.b.0.c or .items[2].id"},
			{"Structure stats", "type tally, depth, node count"},
			{"Diff two docs", "structural added/removed/changed"},
			{"Flatten", "nested → dotted keys"},
			{"Unflatten", "dotted keys → nested"},
			{"Infer schema", "type sketch of the document"},
			{"JSONL / ndjson", "validate + profile line-delimited JSON"},
		})
		switch strings.ToLower(c) {
		case "1":
			doPretty()
		case "2":
			doMinify()
		case "3":
			doValidate()
		case "4":
			doQuery()
		case "5":
			doStats()
		case "6":
			doDiff()
		case "7":
			doFlatten()
		case "8":
			doUnflatten()
		case "9":
			doSchema()
		case "10":
			doJSONL()
		case "h", "help", "?":
			tuiHelp()
		case "0", "00", "q", "exit", "":
			clear()
			okMsg("Goodbye — owlsec.org")
			return
		default:
			warnMsg("unknown selection")
			pause()
		}
	}
}

// ---------------------------------------------------------------------------
// CLI
// ---------------------------------------------------------------------------

func usage() {
	banner()
	sec := func(t string) { fmt.Println("  " + col(t, CY+B)) }
	sec("USAGE")
	fmt.Println("    " + APPNAME + " [mode] [flags] [FILE]   (no mode → interactive TUI)")
	fmt.Println("    ... | " + APPNAME + " [mode]              (read JSON from stdin)")
	fmt.Println()
	sec("MODES")
	fmt.Println("    --pretty            indent/pretty-print the input")
	fmt.Println("    --minify            compact the input (strip whitespace)")
	fmt.Println("    --validate          report validity + root type/depth/nodes")
	fmt.Println("    --query PATH        extract a value by dotted/bracket path")
	fmt.Println("    --stats             print a structure profile")
	fmt.Println("    --diff FILE2        structural diff of input vs FILE2 (exit 1 if differ)")
	fmt.Println("    --flatten           nested → dotted-key object")
	fmt.Println("    --unflatten         dotted-key object → nested")
	fmt.Println("    --schema            print an inferred type sketch")
	fmt.Println("    --jsonl             validate + profile line-delimited JSON")
	fmt.Println()
	sec("INPUT")
	fmt.Println("    FILE                path to a .json file (else stdin, else inline)")
	fmt.Println("    --in STR            inline JSON string")
	fmt.Println()
	sec("OUTPUT")
	fmt.Println("    --no-color          disable ANSI colour")
	fmt.Println()
	sec("GENERAL")
	fmt.Println("    --demo  --selftest  --version  -h/--help")
	fmt.Println()
	sec("EXAMPLES")
	fmt.Println("    " + APPNAME + " --pretty data.json")
	fmt.Println("    cat data.json | " + APPNAME + " --minify")
	fmt.Println("    " + APPNAME + " --query users.0.name data.json")
	fmt.Println("    " + APPNAME + " --validate --in '{\"a\":[1,2,3]}'")
	fmt.Println("    " + APPNAME + " --diff old.json new.json")
	fmt.Println("    " + APPNAME + " --flatten config.json")
	fmt.Println("    " + APPNAME + " --schema sample.json")
	fmt.Println("    cat events.ndjson | " + APPNAME + " --jsonl")
	fmt.Println()
	rule()
}

// cliInput resolves raw JSON from --in, a FILE arg, or stdin (in that order).
func cliInput(inFlag string, args []string) ([]byte, error) {
	if strings.TrimSpace(inFlag) != "" {
		return []byte(inFlag), nil
	}
	if len(args) > 0 {
		return os.ReadFile(args[0])
	}
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return io.ReadAll(os.Stdin)
	}
	return nil, fmt.Errorf("no input: give a FILE, pipe via stdin, or use --in")
}

func runCLI(mode, queryPath, inFlag string, args []string) int {
	raw, err := cliInput(inFlag, args)
	if err != nil {
		errMsg(err.Error())
		return 2
	}
	switch mode {
	case "pretty":
		var v interface{}
		if err := json.Unmarshal(raw, &v); err != nil {
			errMsg("invalid JSON: " + err.Error())
			return 1
		}
		fmt.Println(string(prettyBytes(v)))
	case "minify":
		var buf bytes.Buffer
		if err := json.Compact(&buf, raw); err != nil {
			errMsg("invalid JSON: " + err.Error())
			return 1
		}
		fmt.Println(buf.String())
	case "validate":
		t, d, n, err := validate(raw)
		if err != nil {
			errMsg("INVALID JSON at " + err.Error())
			return 1
		}
		fmt.Printf("valid  type=%s depth=%d nodes=%d\n", t, d, n)
	case "query":
		var v interface{}
		if err := json.Unmarshal(raw, &v); err != nil {
			errMsg("invalid JSON: " + err.Error())
			return 1
		}
		res, err := navigate(v, queryPath)
		if err != nil {
			errMsg(err.Error())
			return 1
		}
		fmt.Println(string(prettyBytes(res)))
	case "stats":
		var v interface{}
		if err := json.Unmarshal(raw, &v); err != nil {
			errMsg("invalid JSON: " + err.Error())
			return 1
		}
		tc := map[string]int{}
		tallyTypes(v, tc)
		var keys []string
		for k := range tc {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := []string{}
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%d", k, tc[k]))
		}
		fmt.Printf("type=%s depth=%d nodes=%d  [%s]\n", typeName(v), depth(v), countNodes(v), strings.Join(parts, " "))
	case "flatten":
		var v interface{}
		if err := json.Unmarshal(raw, &v); err != nil {
			errMsg("invalid JSON: " + err.Error())
			return 1
		}
		flat := map[string]interface{}{}
		flatten(v, "", flat)
		fmt.Println(string(prettyBytes(flat)))
	case "unflatten":
		var v map[string]interface{}
		if err := json.Unmarshal(raw, &v); err != nil {
			errMsg("invalid flat object: " + err.Error())
			return 1
		}
		fmt.Println(string(prettyBytes(unflatten(v))))
	case "schema":
		var v interface{}
		if err := json.Unmarshal(raw, &v); err != nil {
			errMsg("invalid JSON: " + err.Error())
			return 1
		}
		fmt.Println(string(prettyBytes(inferSchema(v))))
	case "jsonl":
		recs, _, err := parseJSONL(raw)
		if err != nil {
			errMsg("invalid JSONL at " + err.Error())
			return 1
		}
		tc := map[string]int{}
		for _, r := range recs {
			tc[typeName(r)]++
		}
		var keys []string
		for k := range tc {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := []string{}
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%d", k, tc[k]))
		}
		fmt.Printf("valid  records=%d  [%s]\n", len(recs), strings.Join(parts, " "))
	}
	return 0
}

// cliDiff compares the primary input (--in/FILE/stdin) against a second file.
func cliDiff(inFlag string, args []string, second string) int {
	rawA, err := cliInput(inFlag, args)
	if err != nil {
		errMsg(err.Error())
		return 2
	}
	rawB, err := os.ReadFile(second)
	if err != nil {
		errMsg(err.Error())
		return 2
	}
	var a, b interface{}
	if json.Unmarshal(rawA, &a) != nil || json.Unmarshal(rawB, &b) != nil {
		errMsg("both inputs must be valid JSON")
		return 1
	}
	var diffs []diffEntry
	diffJSON(a, b, "", &diffs)
	for _, d := range diffs {
		switch d.Kind {
		case "added":
			fmt.Printf("+ %s = %s\n", d.Path, d.New)
		case "removed":
			fmt.Printf("- %s = %s\n", d.Path, d.Old)
		default:
			fmt.Printf("~ %s: %s -> %s\n", d.Path, d.Old, d.New)
		}
	}
	if len(diffs) == 0 {
		fmt.Println("identical")
		return 0
	}
	return 1
}

func runDemo() {
	banner()
	header("Demo — a sample document through every mode")
	sample := `{"service":"owl","port":443,"tags":["net","json"],"opts":{"tls":true,"retries":3},"peers":[{"id":1},{"id":2}]}`
	info("Input:")
	fmt.Println("  " + col(sample, GY))
	fmt.Println()
	var v interface{}
	json.Unmarshal([]byte(sample), &v)
	okMsg("Pretty:")
	colorPrintJSON(string(prettyBytes(v)))
	fmt.Println()
	res, _ := navigate(v, "peers.1.id")
	okMsg("Query peers.1.id → " + fmt.Sprintf("%v", res))
	t, d, n, _ := validate([]byte(sample))
	okMsg(fmt.Sprintf("Validate → %s, depth %d, %d nodes", t, d, n))
	fmt.Println()
	info("Run --selftest to verify the query / depth / validation engines.")
}

func main() {
	var (
		fPretty = flag.Bool("pretty", false, "")
		fMin    = flag.Bool("minify", false, "")
		fVal    = flag.Bool("validate", false, "")
		fQuery  = flag.String("query", "", "")
		fStats  = flag.Bool("stats", false, "")
		fDiff   = flag.String("diff", "", "")
		fFlat   = flag.Bool("flatten", false, "")
		fUnflat = flag.Bool("unflatten", false, "")
		fSchema = flag.Bool("schema", false, "")
		fJSONL  = flag.Bool("jsonl", false, "")
		fIn     = flag.String("in", "", "")
		fNoCol  = flag.Bool("no-color", false, "")
		fDemo   = flag.Bool("demo", false, "")
		fSelf   = flag.Bool("selftest", false, "")
		fVer    = flag.Bool("version", false, "")
	)
	flag.Usage = usage
	flag.Parse()
	if *fNoCol {
		noColor = true
	}
	switch {
	case *fVer:
		fmt.Printf("%s %s  (OWL suite — %s / %s)\n", APPNAME, VERSION, AUTHOR, WEBSITE)
		return
	case *fSelf:
		os.Exit(runSelfTest())
	case *fDemo:
		runDemo()
		return
	case *fPretty:
		os.Exit(runCLI("pretty", "", *fIn, flag.Args()))
	case *fMin:
		os.Exit(runCLI("minify", "", *fIn, flag.Args()))
	case *fVal:
		os.Exit(runCLI("validate", "", *fIn, flag.Args()))
	case *fQuery != "":
		os.Exit(runCLI("query", *fQuery, *fIn, flag.Args()))
	case *fStats:
		os.Exit(runCLI("stats", "", *fIn, flag.Args()))
	case *fFlat:
		os.Exit(runCLI("flatten", "", *fIn, flag.Args()))
	case *fUnflat:
		os.Exit(runCLI("unflatten", "", *fIn, flag.Args()))
	case *fSchema:
		os.Exit(runCLI("schema", "", *fIn, flag.Args()))
	case *fJSONL:
		os.Exit(runCLI("jsonl", "", *fIn, flag.Args()))
	case *fDiff != "":
		os.Exit(cliDiff(*fIn, flag.Args(), *fDiff))
	default:
		interactive()
	}
}

// ---------------------------------------------------------------------------
// Self-test — checks the parsing / navigation / profiling engine.
// ---------------------------------------------------------------------------

func runSelfTest() int {
	banner()
	header("Self-Test — JSON engine")
	pass, fail := 0, 0
	check := func(name string, cond bool) {
		if cond {
			pass++
			fmt.Println("  " + col("[✔] ", GRN) + col(name, W))
		} else {
			fail++
			fmt.Println("  " + col("[✘] ", R) + col(name, W))
		}
	}

	doc := `{"a":1,"b":[10,20,30],"c":{"d":"x","e":[{"f":true}]},"g":null}`
	var v interface{}
	errU := json.Unmarshal([]byte(doc), &v)
	check("parse sample document", errU == nil)

	// typeName.
	check("typeName object", typeName(v) == "object")
	check("typeName array", typeName([]interface{}{}) == "array")
	check("typeName number", typeName(float64(3)) == "number")
	check("typeName string", typeName("s") == "string")
	check("typeName bool", typeName(true) == "bool")
	check("typeName null", typeName(nil) == "null")

	// navigate.
	r1, e1 := navigate(v, "b.1")
	check("navigate b.1 = 20", e1 == nil && r1 == float64(20))
	r2, e2 := navigate(v, "c.e[0].f")
	check("navigate c.e[0].f = true (bracket syntax)", e2 == nil && r2 == true)
	r3, e3 := navigate(v, ".a")
	check("navigate leading-dot .a = 1", e3 == nil && r3 == float64(1))
	_, e4 := navigate(v, "b.9")
	check("navigate out-of-range errors", e4 != nil)
	_, e5 := navigate(v, "nope")
	check("navigate missing key errors", e5 != nil)
	_, e6 := navigate(v, "a.x")
	check("navigate into scalar errors", e6 != nil)
	rroot, eroot := navigate(v, "")
	check("navigate empty path = root", eroot == nil && typeName(rroot) == "object")

	// depth / countNodes.
	check("depth of sample = 5", depth(v) == 5)
	check("depth of scalar = 1", depth(float64(1)) == 1)
	check("countNodes includes containers + leaves", countNodes([]interface{}{float64(1), float64(2)}) == 3)

	// validate + error positioning.
	_, _, _, ev := validate([]byte(`{"a":1}`))
	check("validate accepts good JSON", ev == nil)
	_, _, _, eb := validate([]byte("{\n  \"a\": ,\n}"))
	check("validate rejects bad JSON with position", eb != nil && strings.Contains(eb.Error(), "line 2"))

	// offsetToLineCol.
	l, c := offsetToLineCol([]byte("ab\ncd"), 4)
	check("offsetToLineCol maps byte 4 → line 2 col 2", l == 2 && c == 2)

	// tallyTypes.
	tc := map[string]int{}
	tallyTypes(v, tc)
	check("tallyTypes counts 3 numbers (1,10,20,30 incl a)", tc["number"] == 4)
	check("tallyTypes counts 1 null", tc["null"] == 1)
	check("tallyTypes counts 1 bool", tc["bool"] == 1)

	// prettyBytes + Compact round-trip.
	pb := prettyBytes(v)
	var buf bytes.Buffer
	json.Compact(&buf, pb)
	var back interface{}
	json.Unmarshal(buf.Bytes(), &back)
	check("pretty→compact round-trips to equal depth", depth(back) == depth(v))
	check("prettyBytes has no HTML escaping", !bytes.Contains(prettyBytes(map[string]interface{}{"u": "a<b>&c"}), []byte("\\u003c")))

	// v1.1: diff.
	var da, db interface{}
	json.Unmarshal([]byte(`{"a":1,"b":2,"c":[1,2]}`), &da)
	json.Unmarshal([]byte(`{"a":1,"b":3,"c":[1,2,9],"d":true}`), &db)
	var diffs []diffEntry
	diffJSON(da, db, "", &diffs)
	kinds := map[string]int{}
	for _, d := range diffs {
		kinds[d.Kind]++
	}
	check("diff detects a changed scalar (b)", kinds["changed"] >= 1)
	check("diff detects an added key (d)", kinds["added"] >= 1)
	check("diff detects an added array element (c.2)", func() bool {
		for _, d := range diffs {
			if d.Path == "c.2" && d.Kind == "added" {
				return true
			}
		}
		return false
	}())
	var same interface{}
	json.Unmarshal([]byte(`{"a":1}`), &same)
	var none []diffEntry
	diffJSON(same, same, "", &none)
	check("diff of identical docs is empty", len(none) == 0)

	// v1.1: flatten / unflatten round-trip.
	var nested interface{}
	json.Unmarshal([]byte(`{"a":{"b":[10,20]},"c":"x"}`), &nested)
	flat := map[string]interface{}{}
	flatten(nested, "", flat)
	check("flatten yields dotted keys", flat["a.b.0"] == float64(10) && flat["c"] == "x")
	rebuilt := unflatten(flat)
	rb1 := prettyBytes(rebuilt)
	rb2 := prettyBytes(nested)
	check("unflatten(flatten(x)) == x", string(rb1) == string(rb2))

	// v1.1: schema inference.
	var doc2 interface{}
	json.Unmarshal([]byte(`{"id":1,"tags":["a","b"],"ok":true}`), &doc2)
	sc := inferSchema(doc2).(map[string]interface{})
	check("schema maps number field", sc["id"] == "number")
	check("schema maps homogeneous array", fmt.Sprintf("%v", sc["tags"]) == "[string]")

	// v1.1: JSONL.
	recs, bad, errJL := parseJSONL([]byte("{\"a\":1}\n{\"b\":2}\n\n{\"c\":3}\n"))
	check("parseJSONL reads 3 records skipping blanks", errJL == nil && len(recs) == 3 && bad == 0)
	_, badLine, errBad := parseJSONL([]byte("{\"a\":1}\n{bad}\n"))
	check("parseJSONL reports the bad line number", errBad != nil && badLine == 2)

	fmt.Println()
	if fail == 0 {
		okMsg(fmt.Sprintf("all %d self-tests passed", pass))
		return 0
	}
	errMsg(fmt.Sprintf("%d passed, %d FAILED", pass, fail))
	return 1
}

// ---------------------------------------------------------------------------
// Help & Usage  (interactive menu [H])
// ---------------------------------------------------------------------------

type owlHelpItem struct{ key, text string }

func owlHelpWrap(s string, width int) []string {
	var out []string
	for _, para := range strings.Split(s, "\n") {
		words := strings.Fields(para)
		if len(words) == 0 {
			out = append(out, "")
			continue
		}
		cur := ""
		for _, w := range words {
			if cur != "" && len([]rune(cur))+1+len([]rune(w)) > width {
				out = append(out, cur)
				cur = w
			} else if cur == "" {
				cur = w
			} else {
				cur += " " + w
			}
		}
		out = append(out, cur)
	}
	return out
}

func owlHelpTitle(t string) {
	fmt.Println()
	fmt.Println("  " + col("▸ "+t, B+CY))
}

func owlHelpPara(lines []string) {
	for _, l := range lines {
		for _, w := range owlHelpWrap(l, 60) {
			fmt.Println("    " + col(w, W))
		}
	}
}

func owlHelpRows(items []owlHelpItem) {
	kw := 0
	for _, it := range items {
		if n := len([]rune(it.key)); n > kw {
			kw = n
		}
	}
	for _, it := range items {
		k := "[" + it.key + "]" + strings.Repeat(" ", kw-len([]rune(it.key)))
		for i, w := range owlHelpWrap(it.text, 56-kw) {
			if i == 0 {
				fmt.Println("    " + col(k, YLW+B) + "  " + col(w, W))
			} else {
				fmt.Println("    " + strings.Repeat(" ", kw+2) + "  " + col(w, W))
			}
		}
	}
}

func owlHelpRule() { fmt.Println(col(strings.Repeat("─", 66), CY)) }

func tuiHelp() {
	clear()
	banner()
	owlHelpRule()
	fmt.Println("  " + col("⬡  Help & Usage — "+APPNAME, B+CY))
	owlHelpRule()
	owlHelpTitle("WHAT IT DOES")
	owlHelpPara(owlHelpAbout)
	owlHelpTitle("HOW TO USE THE MENU")
	owlHelpPara(owlHelpFlow)
	owlHelpTitle("MENU OPTIONS")
	owlHelpRows(owlHelpMenu)
	owlHelpTitle("TIPS")
	owlHelpRows(owlHelpTips)
	fmt.Println()
	owlHelpRule()
	fmt.Println("  " + col("Next page: command-line flags & copy-paste examples", DIM))
	pause()
	clear()
	usage()
	fmt.Println()
	fmt.Println("  " + col("Tip: the same reference is printed by  "+APPNAME+" --help", DIM))
	pause()
}

var owlHelpAbout = []string{
	"OJSONForge is a JSON workbench. It pretty-prints with colourised keys, minifies, validates with an exact line/column error position, extracts any value by a dotted or bracket path, and profiles a document's structure (type tally, maximum depth and node count).",
	"Input can be a file path, an inline JSON string, or standard input on the command line — so it drops straight into shell pipelines.",
}

var owlHelpFlow = []string{
	"Type the number of an option and press Enter. Each option asks for a source: paste JSON inline, or type a file path. [H] opens this help, [0] or q quits.",
	"Pretty and Minify offer to write the result to a file. Query uses a path like users.0.name — both dot (.0.) and bracket ([0]) index syntax work.",
}

var owlHelpMenu = []owlHelpItem{
	{"1", "Pretty-print: parse and re-emit with 2-space indentation and colourised keys; optionally write to a file. HTML characters are left unescaped."},
	{"2", "Minify: compact to the smallest valid form and report the bytes saved."},
	{"3", "Validate: confirm the document parses; on failure, report the exact line, column and byte offset of the syntax error."},
	{"4", "Query by path: descend into objects by key and arrays by index (users.0.name or .items[2].id) and print the value found."},
	{"5", "Structure statistics: count values by type, measure maximum nesting depth and total node count, and list the top-level keys."},
	{"6", "Diff two docs: structural comparison of two documents reporting added / removed / changed / type-changed values with their path."},
	{"7", "Flatten: collapse a nested document into a single object of dotted keys (a.b.0 = …) — handy for grepping or spreadsheets."},
	{"8", "Unflatten: rebuild a nested document from a flat dotted-key object; numeric segments become array indices."},
	{"9", "Infer schema: print a compact type sketch of the document (objects keep keys, arrays show element type)."},
	{"10", "JSONL / ndjson: validate every line of a line-delimited JSON stream and profile the record types."},
}

var owlHelpTips = []owlHelpItem{
	{"*", "On the CLI every mode reads from a FILE argument, a pipe (cat x.json | OJSONForge --minify), or --in '…'."},
	{"*", "Object keys are emitted in sorted order by the Go encoder, so pretty-printing also canonicalises key order for diffing."},
	{"*", "Query path indexing is zero-based: the first array element is .0."},
	{"*", "--validate exits non-zero on invalid JSON, so it works as a pre-commit or pipeline gate."},
	{"*", "--diff old.json new.json exits 1 when the documents differ — drop it straight into CI to catch config drift."},
	{"*", "Flatten then unflatten is loss-less and re-sorts keys, giving you a stable canonical form for diffing."},
	{"*", "--schema is a quick way to document an unfamiliar API response or spot an unexpected type."},
}
