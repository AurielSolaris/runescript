package main;

use "rune:dom";

// click: #grow
// click: #grow

public Element Dot() {
    Signal<float> radius = signal(4.5);
    Signal<string> color = signal("teal");
    private void grow() {
        radius.set(radius() * 2);
        color.set("orange");
    }

    template {
        <figure>
            <svg width="100" height="100" viewBox="0 0 100 100">
                <circle cx="50" cy="50" r={{ radius() }} fill={{ color() }} stroke-linecap="round"/>
                <foreignObject x="0" y="0" width="100" height="20"><p>r = {{ radius() }}</p></foreignObject>
            </svg>
            <button id="grow" type="button" onclick={{ grow }}>Grow</button>
        </figure>
    }
}

protected int main() {
    dom.renderApp(Dot, "#app");
    return 0;
}
