package main;

use "rune:dom";

// click: #bigger
// click: #bigger
// click: #smaller

// SVG works like HTML: inside <svg>, elements and attributes come from the
// SVG element list (names such as viewBox are case-sensitive).
//
//     runescript dev examples/web/dial/app.rsc

public Element Dial() {
    Signal<int> size = signal(20);

    private void change(int by) {
        int next = size() + by;
        if (next >= 10 && next <= 45) {
            size.set(next);
        }
    }

    template {
        <figure>
            <svg width="100" height="100" viewBox="0 0 100 100" role="img" aria-label="A dial">
                <circle cx="50" cy="50" r="48" fill="none" stroke="currentColor" stroke-width="2"/>
                <circle cx="50" cy="50" r={{ size() }} fill="teal"/>
            </svg>
            <figcaption>Radius {{ size() }}</figcaption>
            <button id="smaller" type="button" onclick={{ () => change(-5) }}>Smaller</button>
            <button id="bigger" type="button" onclick={{ () => change(5) }}>Bigger</button>
        </figure>
    }
}

protected int main() {
    dom.renderApp(Dial, "#app");
    return 0;
}
