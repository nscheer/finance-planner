// Command e2e-host serves the built frontend and the planner service over
// HTTP, so the end-to-end tests can drive the real application in a browser.
//
// The frontend talks to Go through @wailsio/runtime, and that transport is
// ordinary HTTP: a POST to /wails/runtime carrying {object, method, args},
// answered by the HTTP response itself. A method is addressed by a numeric
// id, which is the FNV-1a hash of "<import path>.<type>.<method>". This
// program reproduces exactly that, dispatching to a real planner.Service, so
// the shipped Svelte code and the generated bindings run unmodified.
//
// Only the desktop shell is missing: the three methods that open a native
// file dialog are answered from a queue the tests fill through /test/dialog.
//
// It is a test tool. It listens on localhost, it is never part of the
// application binary, and it must not be used to serve real data.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"finance-planner/planner"
)

// serviceFQN is the prefix Wails uses to compute the binding ids: the import
// path of the package, the type name, the method name.
const serviceFQN = "finance-planner/planner.Service."

// internalMethods are the methods Wails does not bind, so they have no id.
var internalMethods = map[string]bool{
	"ServiceName": true, "ServiceStartup": true, "ServiceShutdown": true, "ServeHTTP": true,
}

// dialogMethods need a running Wails application; they are answered from the
// queue instead (see host.dialog).
var dialogMethods = map[string]bool{
	"ExportCSV": true, "ExportData": true, "ChooseImportFile": true,
}

type host struct {
	mu      sync.Mutex
	service *planner.Service
	dataDir string
	// dialog is the answer for the next native file dialog: a path, or
	// empty for "the user pressed Cancel".
	dialog string
}

func main() {
	addr := flag.String("addr", "127.0.0.1:34115", "address to listen on")
	dist := flag.String("dist", "frontend/dist", "directory with the built frontend")
	data := flag.String("data", "", "directory for data.json (default: a temp directory)")
	flag.Parse()

	h := &host{}
	if err := h.reset(*data, ""); err != nil {
		log.Fatalf("start service: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /wails/runtime", h.handleRuntime)
	mux.HandleFunc("POST /test/reset", h.handleReset)
	mux.HandleFunc("POST /test/dialog", h.handleDialog)
	// The Wails application injects this file; serving it empty keeps the
	// browser console free of a 404 that means nothing here.
	mux.HandleFunc("GET /wails/custom.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
	})
	mux.Handle("/", http.FileServer(http.Dir(*dist)))

	log.Printf("e2e host on http://%s (data in %s)", *addr, h.dataDir)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}

// ---- service lifecycle ----------------------------------------------------

// reset starts a fresh service. An empty dir means a new temp directory, so
// every test begins with an empty planner.
func (h *host) reset(dir, sampleLang string) error {
	if dir == "" {
		var err error
		dir, err = os.MkdirTemp("", "finance-planner-e2e-")
		if err != nil {
			return err
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	service, err := planner.NewService(filepath.Join(dir, "data.json"))
	if err != nil {
		return err
	}
	if sampleLang != "" {
		if _, err := service.LoadSampleData(sampleLang); err != nil {
			return err
		}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.service, h.dataDir, h.dialog = service, dir, ""
	return nil
}

func (h *host) handleReset(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Sample string `json:"sample"` // language code, empty for an empty planner
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if err := h.reset("", body.Sample); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"dataPath": filepath.Join(h.dataDir, "data.json")})
}

// handleDialog queues the answer of the next native file dialog.
func (h *host) handleDialog(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path   string `json:"path"`
		Cancel bool   `json:"cancel"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	h.mu.Lock()
	h.dialog = body.Path
	if body.Cancel {
		h.dialog = ""
	}
	h.mu.Unlock()
	writeJSON(w, map[string]bool{"ok": true})
}

// ---- the Wails call protocol ---------------------------------------------

// callEnvelope is what @wailsio/runtime posts for a bound method call.
type callEnvelope struct {
	Object int `json:"object"` // 0 = Call
	Args   struct {
		CallID   string            `json:"call-id"`
		MethodID uint32            `json:"methodID"`
		Args     []json.RawMessage `json:"args"`
	} `json:"args"`
}

func (h *host) handleRuntime(w http.ResponseWriter, r *http.Request) {
	var env callEnvelope
	if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
		writeCallError(w, "TypeError", fmt.Errorf("bad request: %w", err))
		return
	}
	if env.Object != 0 {
		writeCallError(w, "ReferenceError", fmt.Errorf("object %d is not served by the e2e host", env.Object))
		return
	}

	h.mu.Lock()
	service := h.service
	h.mu.Unlock()

	name, ok := methodNameByID(service, env.Args.MethodID)
	if !ok {
		writeCallError(w, "ReferenceError", fmt.Errorf("no bound method with id %d", env.Args.MethodID))
		return
	}
	if dialogMethods[name] {
		result, err := h.callDialogMethod(service, name)
		h.answer(w, result, err)
		return
	}

	result, err := call(service, name, env.Args.Args)
	h.answer(w, result, err)
}

// answer writes the result the way Wails does: the value as JSON, or a
// coded error the frontend can translate.
func (h *host) answer(w http.ResponseWriter, result any, err error) {
	if err != nil {
		writeCallError(w, "RuntimeError", err)
		return
	}
	if result == nil {
		writeJSON(w, struct{}{})
		return
	}
	writeJSON(w, result)
}

// callDialogMethod answers the three methods that would open a native file
// dialog. An empty queued path means the user cancelled, which the service
// reports as an empty result and no error (see planner.isDialogCancelled).
func (h *host) callDialogMethod(service *planner.Service, name string) (any, error) {
	h.mu.Lock()
	path := h.dialog
	h.dialog = ""
	h.mu.Unlock()

	switch name {
	case "ExportData":
		if path == "" {
			return "", nil
		}
		return path, service.ExportTo(path)
	case "ExportCSV":
		if path == "" {
			return "", nil
		}
		return path, service.ExportCSVTo(path)
	case "ChooseImportFile":
		if path == "" {
			return planner.ImportPreview{}, nil
		}
		return service.PreviewImport(path)
	}
	return nil, fmt.Errorf("unknown dialog method %s", name)
}

// call invokes a service method by name with JSON arguments.
func call(service *planner.Service, name string, rawArgs []json.RawMessage) (any, error) {
	method := reflect.ValueOf(service).MethodByName(name)
	methodType := method.Type()
	if methodType.NumIn() != len(rawArgs) {
		return nil, fmt.Errorf("%s takes %d arguments, got %d", name, methodType.NumIn(), len(rawArgs))
	}

	in := make([]reflect.Value, len(rawArgs))
	for i, raw := range rawArgs {
		value := reflect.New(methodType.In(i))
		if err := json.Unmarshal(raw, value.Interface()); err != nil {
			return nil, fmt.Errorf("argument %d of %s: %w", i+1, name, err)
		}
		in[i] = value.Elem()
	}

	// Which output is the error is decided by the method's signature, not by
	// the returned value: a nil error arrives as an interface with no dynamic
	// type, so a type assertion on it fails and would overwrite the result
	// with nil.
	var result any
	for i, out := range method.Call(in) {
		if methodType.Out(i).Implements(errorType) {
			if err, _ := out.Interface().(error); err != nil {
				return nil, err
			}
			continue
		}
		result = out.Interface()
	}
	return result, nil
}

var errorType = reflect.TypeFor[error]()

// methodNameByID reverses the binding id Wails computes from the fully
// qualified method name.
func methodNameByID(service *planner.Service, id uint32) (string, bool) {
	t := reflect.TypeOf(service)
	for i := range t.NumMethod() {
		name := t.Method(i).Name
		if internalMethods[name] {
			continue
		}
		if fnvHash(serviceFQN+name) == id {
			return name, true
		}
	}
	return "", false
}

func fnvHash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

// ---- responses ------------------------------------------------------------

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}

// writeCallError uses the shape @wailsio/runtime parses: the kind decides the
// exception class, and "cause" carries the coded error the frontend
// translates.
func writeCallError(w http.ResponseWriter, kind string, err error) {
	body := map[string]any{"kind": kind, "message": err.Error()}
	if raw := planner.MarshalError(err); raw != nil {
		body["cause"] = json.RawMessage(raw)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write error response: %v", err)
	}
}
