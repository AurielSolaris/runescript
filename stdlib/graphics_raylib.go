//go:build runescript_gfx && ((js && wasm) || (cgo && (windows || linux || darwin) && (amd64 || arm64)))

package stdlib

import (
	"image/color"
	"strings"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() { newGraphicsBackend = func() graphicsBackend { return raylibBackend{} } }

type raylibBackend struct{}

func (raylibBackend) open(width, height int32, title string) bool {
	rl.InitWindow(width, height, title)
	return rl.IsWindowReady()
}

func (raylibBackend) close()             { rl.CloseWindow() }
func (raylibBackend) shouldClose() bool  { return rl.WindowShouldClose() }
func (raylibBackend) setFPS(fps int32)   { rl.SetTargetFPS(fps) }
func (raylibBackend) deltaTime() float32 { return rl.GetFrameTime() }
func (raylibBackend) begin()             { rl.BeginDrawing() }
func (raylibBackend) end()               { rl.EndDrawing() }
func (raylibBackend) clear(c color.RGBA) { rl.ClearBackground(c) }

func (raylibBackend) rectangle(r gfxRect, c color.RGBA) {
	rl.DrawRectangle(r.x, r.y, r.width, r.height, c)
}

func (raylibBackend) circle(p gfxPoint, radius float32, c color.RGBA) {
	rl.DrawCircleV(rl.Vector2{X: p.x, Y: p.y}, radius, c)
}

func (raylibBackend) line(a, b gfxPoint, c color.RGBA) {
	rl.DrawLineV(rl.Vector2{X: a.x, Y: a.y}, rl.Vector2{X: b.x, Y: b.y}, c)
}

func (raylibBackend) text(s string, x, y, size int32, c color.RGBA) {
	rl.DrawText(s, x, y, size, c)
}

var raylibKeys = map[string]int32{
	"LEFT": rl.KeyLeft, "RIGHT": rl.KeyRight, "UP": rl.KeyUp, "DOWN": rl.KeyDown,
	"SPACE": rl.KeySpace, "ENTER": rl.KeyEnter, "ESCAPE": rl.KeyEscape, "TAB": rl.KeyTab,
}

func (raylibBackend) keyDown(name string) bool {
	name = strings.ToUpper(name)
	if len(name) == 1 {
		key := name[0]
		if key >= 'A' && key <= 'Z' || key >= '0' && key <= '9' {
			return rl.IsKeyDown(int32(key))
		}
	}
	if code, ok := raylibKeys[name]; ok {
		return rl.IsKeyDown(code)
	}
	argFault("unknown key %q", name)
	return false
}

func (raylibBackend) mousePosition() gfxPoint {
	p := rl.GetMousePosition()
	return gfxPoint{x: p.X, y: p.Y}
}

func (raylibBackend) mousePressed(button int32) bool {
	return rl.IsMouseButtonPressed(rl.MouseButton(button))
}

func (raylibBackend) loadTexture(path string) (any, bool) {
	t := rl.LoadTexture(path)
	return t, rl.IsTextureValid(t)
}

func (raylibBackend) unloadTexture(t any) { rl.UnloadTexture(t.(rl.Texture2D)) }

func (raylibBackend) drawTexture(t any, x, y int32, c color.RGBA) {
	rl.DrawTexture(t.(rl.Texture2D), x, y, c)
}

func raylibRect(r gfxRect) rl.Rectangle {
	return rl.Rectangle{X: float32(r.x), Y: float32(r.y), Width: float32(r.width), Height: float32(r.height)}
}

func (raylibBackend) button(r gfxRect, text string) bool {
	return gui.Button(raylibRect(r), text) != 0
}

func (raylibBackend) label(r gfxRect, text string) { gui.Label(raylibRect(r), text) }

func (raylibBackend) checkBox(r gfxRect, text string, checked bool) bool {
	gui.CheckBox(raylibRect(r), text, &checked)
	return checked
}
