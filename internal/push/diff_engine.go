// diff_engine.go: bridge between the read side (intented ConfigData) and the
// renderer's Diff(). The engine is vendor-agnostic — its job is to pick the
// right renderer for an OS type, call Diff on it, and assemble the rendered
// patch (the human-readable text the operator sees in `nsl-graph push preview`).
package push

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/configparser/parsers"
)

// Engine holds a renderer lookup keyed by OS type. Two factories:
//
//   - NewEngine(map): explicit registry — used by tests and by callers that want
//     to pin a specific renderer set (e.g. feature-gated builds).
//   - NewDefaultEngine(): the global DefaultRendererRegistry from the
//     configparser package — what cmd/push/* uses.
type Engine struct {
	renderers map[string]configparser.ConfigRenderer
}

// NewEngine constructs an Engine from a renderer lookup.
func NewEngine(renderers map[string]configparser.ConfigRenderer) *Engine {
	return &Engine{renderers: renderers}
}

// NewDefaultEngine returns an Engine backed by the global DefaultRendererRegistry.
//
// Override semantics for the opnsense slot, in priority order:
//  1. WithOpnsenseFactory option supplied → the option's factory wins,
//     regardless of env vars. Tests and embedded callers use this.
//  2. OPNSENSE_URL/OPNSENSE_KEY/OPNSENSE_SECRET env vars all set →
//     build the typed renderer from those (lab path).
//  3. Otherwise keep the registry's lazy renderer.
func NewDefaultEngine(opts ...EngineOption) *Engine {
	r := configparser.DefaultRendererRegistry.Renderers()
	out := make(map[string]configparser.ConfigRenderer, len(r))
	for k, v := range r {
		out[k] = v
	}
	o := applyOptions(opts)
	if out["opnsense"] == nil {
		return &Engine{renderers: out}
	}
	if o.opnsenseOverridden {
		// Option wins unconditionally — caller takes responsibility for
		// wiring the renderer (e.g. httptest URL in unit tests).
		base, key, secret, _ := opnsenseEnv()
		out["opnsense"] = o.opnsenseFactory(base, key, secret)
		return &Engine{renderers: out}
	}
	if base, key, secret, ok := opnsenseEnv(); ok {
		out["opnsense"] = o.opnsenseFactory(base, key, secret)
	}
	return &Engine{renderers: out}
}

// Diff returns the per-line differences between the intended and observed configs
// by routing to the renderer registered for osType. Returns ErrUnsupported when
// no renderer is registered.
func (e *Engine) Diff(osType string, intended, observed *configparser.ConfigData) ([]configparser.ConfigChange, error) {
	r, err := e.pick(osType)
	if err != nil {
		return nil, err
	}
	return r.Diff(intended, observed), nil
}

// Preview assembles the renderer's per-change Patch lines into a single text
// blob, with a header. Empty when there are no changes.
//
// The format is intentionally close to a unified diff so an operator can read it
// on the terminal; it is NOT meant to be a NETCONF <edit-config> payload.
func (e *Engine) Preview(osType string, intended, observed *configparser.ConfigData) (string, error) {
	diffs, err := e.Diff(osType, intended, observed)
	if err != nil {
		return "", err
	}
	if len(diffs) == 0 {
		return "", nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# push preview for %s\n", osType)
	fmt.Fprintf(&b, "# %d change(s)\n", len(diffs))
	for _, d := range diffs {
		fmt.Fprintf(&b, "\n@@ %s @@\n", d.Path)
		for _, line := range d.Patch {
			fmt.Fprintf(&b, "+ %s\n", line)
		}
	}
	return b.String(), nil
}

// pick resolves the renderer for an OS type, with the same alias logic as
// GetRendererOrError in the configparser package.
func (e *Engine) pick(osType string) (configparser.ConfigRenderer, error) {
	r, ok := e.renderers[osType]
	if !ok {
		return nil, configparser.ErrUnsupported{OS: osType, Reason: "no renderer registered"}
	}
	return r, nil
}

// list returns the registered renderer OS types sorted. Useful for the
// `nsl-graph push devices --list` discoverability flag.
func (e *Engine) list() []string {
	out := make([]string, 0, len(e.renderers))
	for k := range e.renderers {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Renderers returns a copy of the renderers map keyed by OS type. Useful
// for tests that need to introspect the resolved renderer without going
// through Diff/Preview. Callers MUST NOT mutate the returned map.
func (e *Engine) Renderers() map[string]configparser.ConfigRenderer {
	out := make(map[string]configparser.ConfigRenderer, len(e.renderers))
	for k, v := range e.renderers {
		out[k] = v
	}
	return out
}

// ---------------------------------------------------------------------------
// Options for NewDefaultEngine
// ---------------------------------------------------------------------------

// EngineOption mutates the engine's construction-time options.
type EngineOption func(*engineOptions)

type engineOptions struct {
	opnsenseFactory    func(baseURL, key, secret string) configparser.ConfigRenderer
	opnsenseOverridden bool // true when WithOpnsenseFactory was supplied
}

func applyOptions(opts []EngineOption) *engineOptions {
	o := &engineOptions{}
	for _, opt := range opts {
		opt(o)
	}
	if o.opnsenseFactory == nil {
		o.opnsenseFactory = defaultOpnsenseFactory
	}
	return o
}

// WithOpnsenseFactory lets a test substitute the renderer constructor.
// ponytail: passing a non-nil factory flips the override flag so the
// engine invokes it without needing the OPNSENSE_URL env vars to also
// be set — otherwise a unit test would have to mock the env to wire a
// httptest URL into the renderer.
func WithOpnsenseFactory(f func(baseURL, key, secret string) configparser.ConfigRenderer) EngineOption {
	return func(o *engineOptions) {
		o.opnsenseFactory = f
		o.opnsenseOverridden = true
	}
}

// defaultOpnsenseFactory builds the real typed renderer. It is the only place
// in the push package that imports the vendored opnsense-api + parser; the
// rest of the engine stays vendor-agnostic.
func defaultOpnsenseFactory(baseURL, key, secret string) configparser.ConfigRenderer {
	return parsers.NewOpnsenseRendererForURL(baseURL, key, secret)
}

// opnsenseEnv reads OPNSENSE_URL/KEY/SECRET. All three must be non-empty
// for the typed branch to engage.
func opnsenseEnv() (string, string, string, bool) {
	base := strings.TrimSpace(os.Getenv("OPNSENSE_URL"))
	key := os.Getenv("OPNSENSE_KEY")
	secret := os.Getenv("OPNSENSE_SECRET")
	if base == "" || key == "" || secret == "" {
		return "", "", "", false
	}
	return base, key, secret, true
}
