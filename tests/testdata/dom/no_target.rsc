package main;

use "rune:dom";

public Element App() {
    template {
        <p>never shown</p>
    }
}

public void main() {
    dom.renderApp(App, "#missing");
}
