package game;

use "rune.gfx";

protected int main() {
    gfx.openWindow(640, 400, "Keep up");
    gfx.setFPS(60);
    gfx.Color background = gfx.rgba(20, 32, 40, 255);
    gfx.Color foreground = gfx.rgba(211, 247, 239, 255);
    gfx.Vector2 ball = gfx.vector(120.0, 60.0);
    float32 vx = 180.0;
    float32 vy = 170.0;
    float32 paddle = 270.0;
    int hits = 0;
    while (!gfx.shouldClose()) {
        float32 dt = gfx.deltaTime();
        if (gfx.keyDown("Left") && paddle > 8.0) {
            paddle -= 300.0 * dt;
        }
        if (gfx.keyDown("Right") && paddle < 532.0) {
            paddle += 300.0 * dt;
        }
        ball.x += vx * dt;
        ball.y += vy * dt;
        if ((ball.x < 13.0 && vx < 0.0) || (ball.x > 627.0 && vx > 0.0)) {
            vx = -vx;
        }
        if (ball.y < 13.0 && vy < 0.0) {
            vy = -vy;
        }
        if (vy > 0.0 && ball.y >= 343.0 && ball.y <= 365.0 &&
            ball.x >= paddle && ball.x <= paddle + 100.0) {
            vy = -vy;
            hits++;
        }
        if (ball.y > 412.0) {
            ball = gfx.vector(120.0, 60.0);
            vy = 170.0;
            hits = 0;
        }

        gfx.beginFrame();
        gfx.clear(background);
        gfx.text($"Hits {hits}   Left / Right to move", 20, 18, 22, foreground);
        gfx.rectangle((int) paddle, 355, 100, 12, foreground);
        gfx.circle(ball, 12.0, foreground);
        gfx.endFrame();
    }
    return 0;
}
