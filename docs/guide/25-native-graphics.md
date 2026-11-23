# 25. Graphics on desktop and in the browser

`rune.gfx` provides a desktop window or browser canvas, drawing, textures,
and input. `rune.gui` adds immediate-mode widgets using raygui. The same
RuneScript program can run through the desktop CLI or WebAssembly.

## Desktop

Build with Go 1.25+, cgo, and a desktop C toolchain. On Windows, use a
MinGW-w64 compiler on PATH. See the upstream platform dependencies for
Linux and macOS. [raylib-go](https://github.com/gen2brain/raylib-go)

```sh
CGO_ENABLED=1 go build -tags runescript_gfx -o runescript ./cmd/runescript
./runescript run examples/native/keep-up.rsc
```

In PowerShell:

```powershell
$env:CGO_ENABLED = '1'
go build -tags runescript_gfx -o runescript.exe ./cmd/runescript
.\runescript.exe run examples/native/click-circle.rsc
```

## Browser

From the RuneScript checkout, build the game into a new directory:

```sh
go run ./cmd/runescript build --target wasm -o game-web examples/native/click-circle.rsc
python -m http.server --bind 127.0.0.1 --directory game-web 8000
```

Open `http://localhost:8000`. The output includes the pinned raylib WASM
module and a Go WASM runtime that runs the game. Serve the whole directory
over HTTP with the `.wasm` files served as `application/wasm`.

For a project outside the RuneScript checkout, set `RUNESCRIPT_SOURCE`
to the RuneScript source directory and run `runescript build --target wasm`
from the project's root. Imports and assets use paths relative to that
working directory. Repeat `--asset path/to/image.png` for each texture or
other file the program reads at run time. The bundle includes the entry
and its imported source files automatically. The browser uses in-memory
files, so writes last for the current page session.

This target runs RuneScript's native graphics API through raylib's WebAssembly
backend. Use the JavaScript target for `rune:dom` and JavaScript `extern`s.

The examples run on both targets. [Click the circle](../../examples/native/click-circle.rsc)
uses the mouse and a reset button. [Keep up](../../examples/native/keep-up.rsc)
uses the arrow keys to move a paddle.

## Window and frames

```text
package game;
use "rune.gfx";

protected int main() {
    gfx.openWindow(640, 400, "Hello");
    gfx.setFPS(60);
    gfx.Color background = gfx.rgba(18, 24, 36, 255);
    gfx.Color ink = gfx.rgba(240, 246, 255, 255);
    while (!gfx.shouldClose()) {
        gfx.beginFrame();
        gfx.clear(background);
        gfx.circle(gfx.vector(320.0, 200.0), 24.0, ink);
        gfx.endFrame();
    }
    gfx.closeWindow();
    return 0;
}
```

One program at a time can own the graphics window. Closing it releases
its textures. The runtime also closes it when the program exits or throws
an uncaught error. Drawing requires an active frame.

| Function | Behavior |
| --- | --- |
| `rgba(r, g, b, a)` | Returns a `Color` struct with four `uint8` fields. |
| `vector(x, y)` | Returns a `Vector2` struct with two `float32` fields. |
| `rectangle(x, y, width, height, color)` | Draws a filled rectangle. The four geometry arguments are `int`. |
| `circle(center, radius, color)` | Draws a filled circle. The center is a `Vector2` and radius is `float32`. |
| `line(start, end, color)` | Draws between two `Vector2` values. |
| `text(text, x, y, fontSize, color)` | Draws text with raylib's default font. |
| `deltaTime()` | Returns seconds elapsed during the previous frame as a `float32`. |
| `keyDown(key)` | Tests a held key. Names are case-insensitive. Supported keys are letters, digits, arrows, Space, Enter, Escape, and Tab. |
| `mousePosition()` | Returns the pointer position as a `Vector2` in window coordinates. |
| `mousePressed(button)` | Tests the press for this frame. Use 0 for left, 1 for right, and 2 for middle. |
| `loadTexture(path)` | Loads an image between frames and returns a `Texture` handle. Paths resolve from the working directory. |
| `drawTexture(texture, x, y, tint)` | Draws a loaded texture at integer coordinates. White with full alpha preserves the image colors. |
| `unloadTexture(texture)` | Releases a texture between frames. Copies of the handle become invalid together. |

Import `rune.gui` and call its widgets inside the frame. Each takes
`x, y, width, height, text` with four `int` coordinates and a `string`.
`gui.button` returns true on activation. `gui.label` renders a label.
`gui.checkBox` takes the current `bool` as a sixth argument and returns
the updated value. The program stores widget state between frames.
