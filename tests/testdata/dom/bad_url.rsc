package main;

use "rune:dom";

// click: #go

// A bound URL that would run script is refused, and can be caught.
public Element App() {
    Signal<string> url = signal("https://example.com/");
    Signal<string> status = signal("ok");
    private void go() {
        try {
            url.set(" JavaScript:alert(1)");
        } catch (TypeError e) {
            status.set(e.message);
        }
    }
    template {
        <div><a href={{ url() }}>link</a> <span>{{ status() }}</span> <button id="go" onclick={{ go }}>go</button></div>
    }
}

public void main() {
    dom.renderApp(App, "#app");
}
