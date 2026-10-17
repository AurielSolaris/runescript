package main;

use "rune:dom";

public Element App() {
    template {
        <p>never shown</p>
    }
}

protected int main() {
    dom.renderApp(App, "#missing");
    return 0;
}
