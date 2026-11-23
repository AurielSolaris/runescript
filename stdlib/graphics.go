package stdlib

import (
	"fmt"
	"image/color"
	"math"
	"runtime"
	"strings"
	"sync"

	"github.com/AurielSolaris/runescript/internal/types"
)

type gfxPoint struct{ x, y float32 }
type gfxRect struct{ x, y, width, height int32 }

type graphicsBackend interface {
	open(width, height int32, title string) bool
	close()
	shouldClose() bool
	setFPS(int32)
	deltaTime() float32
	begin()
	end()
	clear(color.RGBA)
	rectangle(gfxRect, color.RGBA)
	circle(gfxPoint, float32, color.RGBA)
	line(gfxPoint, gfxPoint, color.RGBA)
	text(string, int32, int32, int32, color.RGBA)
	keyDown(string) bool
	mousePosition() gfxPoint
	mousePressed(int32) bool
	loadTexture(string) (any, bool)
	unloadTexture(any)
	drawTexture(any, int32, int32, color.RGBA)
	button(gfxRect, string) bool
	label(gfxRect, string)
	checkBox(gfxRect, string, bool) bool
}

// Native desktop backends install a factory at build time. The library owns
// process-global state, so one program at a time can own the native window.
var newGraphicsBackend func() graphicsBackend
var graphicsWindow sync.Mutex

type graphicsTexture struct{ handle any }

type graphicsSession struct {
	backend  graphicsBackend
	inFrame  bool
	textures map[*graphicsTexture]struct{}
}

func (e *Env) openGraphics(width, height int32, title string, factory func() graphicsBackend) {
	if e.graphics != nil {
		gfxFault("a window is already open")
	}
	if factory == nil {
		gfxFault("this build has no graphics backend; build with -tags runescript_gfx and CGO_ENABLED=1")
	}
	if !graphicsWindow.TryLock() {
		gfxFault("another RuneScript program owns the graphics window")
	}
	runtime.LockOSThread()
	opened := false
	defer func() {
		if !opened {
			runtime.UnlockOSThread()
			graphicsWindow.Unlock()
		}
	}()
	b := factory()
	if !b.open(width, height, title) {
		gfxFault("could not open a graphics window")
	}
	e.graphics = &graphicsSession{backend: b, textures: map[*graphicsTexture]struct{}{}}
	opened = true
}

// CloseGraphics releases a program's graphics resources at exit, including
// an exceptional exit inside a frame.
func (e *Env) CloseGraphics() {
	g := e.graphics
	if g == nil {
		return
	}
	e.graphics = nil
	defer graphicsWindow.Unlock()
	defer runtime.UnlockOSThread()
	defer g.backend.close()
	if g.inFrame {
		g.backend.end()
	}
	for tex := range g.textures {
		g.backend.unloadTexture(tex.handle)
	}
}

func gfxFault(format string, args ...any) {
	panic(Fault{Type: "Error", Message: fmt.Sprintf(format, args...)})
}

func (e *Env) window() *graphicsSession {
	if e.graphics == nil {
		gfxFault("open a window first with gfx.openWindow")
	}
	return e.graphics
}

func (e *Env) frame() graphicsBackend {
	g := e.window()
	if !g.inFrame {
		gfxFault("drawing requires an active frame")
	}
	return g.backend
}

func (g *graphicsSession) texture(value any) *graphicsTexture {
	tex := value.(*graphicsTexture)
	if _, ok := g.textures[tex]; !ok {
		gfxFault("the texture is no longer loaded in this window")
	}
	return tex
}

func (g *graphicsSession) betweenFrames() {
	if g.inFrame {
		gfxFault("load or unload textures between frames")
	}
}

func gfxInt(v any) int32 { return int32(v.(int64)) }

func gfxPositive(v any, name string) int32 {
	n := gfxInt(v)
	if n <= 0 {
		argFault("%s must be positive", name)
	}
	return n
}

func gfxBounds(a []any) gfxRect {
	return gfxRect{gfxInt(a[0]), gfxInt(a[1]), gfxPositive(a[2], "width"), gfxPositive(a[3], "height")}
}

func gfxColor(v any) color.RGBA {
	a := v.([]any)
	return color.RGBA{R: byte(gfxInt(a[0])), G: byte(gfxInt(a[1])), B: byte(gfxInt(a[2])), A: byte(gfxInt(a[3]))}
}

func gfxFloat(v any) float32 {
	f := v.(float64)
	if math.IsNaN(f) || math.IsInf(f, 0) {
		argFault("coordinates and sizes must be finite")
	}
	return float32(f)
}

func gfxVector(v any) gfxPoint {
	a := v.([]any)
	return gfxPoint{gfxFloat(a[0]), gfxFloat(a[1])}
}

func gfxString(v any) string {
	s := str(v)
	if strings.ContainsRune(s, 0) {
		argFault("graphics strings cannot contain a zero byte")
	}
	return s
}

func init() {
	I, F, S, B, V := types.Int, types.Float32, types.StringT, types.BoolT, types.VoidT
	U := types.Lookup("uint8")
	C := types.NewStruct("gfx.Color")
	C.Fields = []*types.Field{{Name: "r", Type: U}, {Name: "g", Type: U}, {Name: "b", Type: U}, {Name: "a", Type: U}}
	P := types.NewStruct("gfx.Vector2")
	P.Fields = []*types.Field{{Name: "x", Type: F}, {Name: "y", Type: F}}
	T := types.NewOpaque("gfx.Texture")
	gfx := &Package{Path: "rune.gfx", Qualifier: "gfx", NativeOnly: true,
		Types: map[string]*types.Type{"Color": C, "Vector2": P, "Texture": T}}
	gfx.add(
		fn("rgba", C, func(a []any) any { return a }, U, U, U, U),
		fn("vector", P, func(a []any) any { return a }, F, F),
		envFn("openWindow", V, func(e *Env, a []any) any {
			e.openGraphics(gfxPositive(a[0], "width"), gfxPositive(a[1], "height"), gfxString(a[2]), newGraphicsBackend)
			return nil
		}, I, I, S),
		envFn("closeWindow", V, func(e *Env, _ []any) any {
			e.CloseGraphics()
			return nil
		}),
		envFn("shouldClose", B, func(e *Env, _ []any) any { return e.window().backend.shouldClose() }),
		envFn("setFPS", V, func(e *Env, a []any) any {
			e.window().backend.setFPS(gfxPositive(a[0], "frames per second"))
			return nil
		}, I),
		envFn("deltaTime", F, func(e *Env, _ []any) any { return float64(e.window().backend.deltaTime()) }),
		envFn("beginFrame", V, func(e *Env, _ []any) any {
			g := e.window()
			if g.inFrame {
				gfxFault("a frame is already active")
			}
			g.backend.begin()
			g.inFrame = true
			return nil
		}),
		envFn("endFrame", V, func(e *Env, _ []any) any {
			e.frame().end()
			e.graphics.inFrame = false
			return nil
		}),
		envFn("clear", V, func(e *Env, a []any) any {
			e.frame().clear(gfxColor(a[0]))
			return nil
		}, C),
		envFn("rectangle", V, func(e *Env, a []any) any {
			e.frame().rectangle(gfxBounds(a), gfxColor(a[4]))
			return nil
		}, I, I, I, I, C),
		envFn("circle", V, func(e *Env, a []any) any {
			radius := gfxFloat(a[1])
			if radius < 0 {
				argFault("radius cannot be negative")
			}
			e.frame().circle(gfxVector(a[0]), radius, gfxColor(a[2]))
			return nil
		}, P, F, C),
		envFn("line", V, func(e *Env, a []any) any {
			e.frame().line(gfxVector(a[0]), gfxVector(a[1]), gfxColor(a[2]))
			return nil
		}, P, P, C),
		envFn("text", V, func(e *Env, a []any) any {
			e.frame().text(gfxString(a[0]), gfxInt(a[1]), gfxInt(a[2]), gfxPositive(a[3], "font size"), gfxColor(a[4]))
			return nil
		}, S, I, I, I, C),
		envFn("keyDown", B, func(e *Env, a []any) any { return e.window().backend.keyDown(str(a[0])) }, S),
		envFn("mousePosition", P, func(e *Env, _ []any) any {
			p := e.window().backend.mousePosition()
			return []any{float64(p.x), float64(p.y)}
		}),
		envFn("mousePressed", B, func(e *Env, a []any) any {
			button := gfxInt(a[0])
			if button < 0 || button > 2 {
				argFault("mouse button must be 0 (left), 1 (right), or 2 (middle)")
			}
			return e.window().backend.mousePressed(button)
		}, I),
		envFn("loadTexture", T, func(e *Env, a []any) any {
			g := e.window()
			g.betweenFrames()
			path := gfxString(a[0])
			handle, ok := g.backend.loadTexture(path)
			if !ok {
				gfxFault("could not load texture %q", path)
			}
			tex := &graphicsTexture{handle: handle}
			g.textures[tex] = struct{}{}
			return tex
		}, S),
		envFn("unloadTexture", V, func(e *Env, a []any) any {
			g := e.window()
			g.betweenFrames()
			tex := g.texture(a[0])
			g.backend.unloadTexture(tex.handle)
			delete(g.textures, tex)
			return nil
		}, T),
		envFn("drawTexture", V, func(e *Env, a []any) any {
			b := e.frame()
			b.drawTexture(e.graphics.texture(a[0]).handle, gfxInt(a[1]), gfxInt(a[2]), gfxColor(a[3]))
			return nil
		}, T, I, I, C),
	)
	register(gfx)

	gui := &Package{Path: "rune.gui", Qualifier: "gui", NativeOnly: true}
	gui.add(
		envFn("button", B, func(e *Env, a []any) any {
			return e.frame().button(gfxBounds(a), gfxString(a[4]))
		}, I, I, I, I, S),
		envFn("label", V, func(e *Env, a []any) any {
			e.frame().label(gfxBounds(a), gfxString(a[4]))
			return nil
		}, I, I, I, I, S),
		envFn("checkBox", B, func(e *Env, a []any) any {
			return e.frame().checkBox(gfxBounds(a), gfxString(a[4]), a[5].(bool))
		}, I, I, I, I, S, B),
	)
	register(gui)
}
