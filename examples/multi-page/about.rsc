package about;
use "rune:dom";

public Element page() {
    template {
        <section style="padding: 24px; background-color: #edf4ff; color: #183454; border-radius: 12px">
            <h1>About</h1>
            <p>Both pages use the same cached runtime module.</p>
        </section>
    }
}

protected int main() {
    dom.renderApp(page, "#app");
    return 0;
}
