package home;
use "rune:dom";

public Element page() {
    template {
        <section style="padding: 24px; background-color: #edf4ff; color: #183454; border-radius: 12px">
            <h1>Home</h1>
            <p>Each page starts its own RuneScript program.</p>
        </section>
    }
}

protected int main() {
    dom.renderApp(page, "#app");
    return 0;
}
