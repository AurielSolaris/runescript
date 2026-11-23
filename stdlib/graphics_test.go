package stdlib

import (
	"image/color"
	"io"
	"reflect"
	"testing"
)

type fakeGraphics struct {
	graphicsBackend
	events     []string
	clearColor color.RGBA
}

func (b *fakeGraphics) open(_, _ int32, _ string) bool {
	b.events = append(b.events, "open")
	return true
}
func (b *fakeGraphics) close()                          { b.events = append(b.events, "close") }
func (b *fakeGraphics) begin()                          { b.events = append(b.events, "begin") }
func (b *fakeGraphics) end()                            { b.events = append(b.events, "end") }
func (b *fakeGraphics) clear(c color.RGBA)              { b.clearColor = c }
func (b *fakeGraphics) button(_ gfxRect, _ string) bool { return true }
func (b *fakeGraphics) loadTexture(_ string) (any, bool) {
	b.events = append(b.events, "load")
	return 17, true
}
func (b *fakeGraphics) unloadTexture(_ any) { b.events = append(b.events, "unload") }

func TestGraphicsBindingLifecycle(t *testing.T) {
	b := &fakeGraphics{}
	e := NewEnv(io.Discard, io.Discard, nil, nil)
	e.openGraphics(320, 240, "headless", func() graphicsBackend { return b })
	t.Cleanup(e.CloseGraphics)
	gfx := Lookup("rune.gfx").Funcs
	gui := Lookup("rune.gui").Funcs
	call := func(name string, args ...any) any { return gfx[name].EnvImpl(e, args) }
	fault := func(fn func()) {
		t.Helper()
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected graphics fault")
			} else if _, ok := r.(Fault); !ok {
				panic(r)
			}
		}()
		fn()
	}
	fault(func() { call("endFrame") })
	tex := call("loadTexture", "example.png")
	call("unloadTexture", tex)
	live := call("loadTexture", "other.png")
	call("beginFrame")
	fault(func() { call("beginFrame") })
	call("clear", []any{int64(10), int64(20), int64(30), int64(255)})
	if b.clearColor != (color.RGBA{R: 10, G: 20, B: 30, A: 255}) {
		t.Fatalf("wrong color: %v", b.clearColor)
	}
	if pressed := gui["button"].EnvImpl(e, []any{int64(8), int64(8), int64(48), int64(24), "Reset"}); pressed != true {
		t.Errorf("button returned %v", pressed)
	}
	fault(func() { call("drawTexture", tex, int64(0), int64(0), nil) })
	fault(func() { call("unloadTexture", live) })
	e.CloseGraphics()
	if want := []string{"open", "load", "unload", "load", "begin", "end", "unload", "close"}; !reflect.DeepEqual(b.events, want) {
		t.Fatalf("events %v, want %v", b.events, want)
	}
}
