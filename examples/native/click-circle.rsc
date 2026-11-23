package game;

use "rune.gfx";
use "rune.gui";

protected int main() {
    gfx.openWindow(640, 400, "Click the circle");
    gfx.setFPS(60);
    gfx.Color background = gfx.rgba(22, 30, 44, 255);
    gfx.Color target = gfx.rgba(244, 170, 70, 255);
    gfx.Color foreground = gfx.rgba(240, 245, 255, 255);
    gfx.Vector2 center = gfx.vector(320.0, 200.0);
    int score = 0;
    while (!gfx.shouldClose()) {
        gfx.Vector2 mouse = gfx.mousePosition();
        float32 dx = mouse.x - center.x;
        float32 dy = mouse.y - center.y;
        if (gfx.mousePressed(0) && dx * dx + dy * dy <= 625.0) {
            score++;
            center.x = (float32) (80 + (score * 137) % 480);
            center.y = (float32) (100 + (score * 61) % 200);
        }

        gfx.beginFrame();
        gfx.clear(background);
        gfx.circle(center, 25.0, target);
        gfx.text($"Score {score}", 20, 18, 24, foreground);
        if (gui.button(510, 14, 110, 34, "Reset score")) {
            score = 0;
            center = gfx.vector(320.0, 200.0);
        }
        gfx.endFrame();
    }
    gfx.closeWindow();
    return 0;
}
