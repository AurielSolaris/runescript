package main;

use "rune:dom";

// click: #go

// A bound keyword attribute is checked when it is set.
public Element App() {
    Signal<string> kind = signal("text");
    template {
        <div>
            <input type={{ kind() }}>
            <button id="go" onclick={{ () => kind.set("textarea") }}>go</button>
        </div>
    }
}

public void main() {
    dom.renderApp(App, "#app");
}
