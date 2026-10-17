package main;

use "rune:dom";
use "./testdata/dom/parts.rsm";

// click: .toggle
// click: #step
// click: #step
// click: .toggle
// click: .toggle

// Child components keep their own runes; $if removes a child's bindings
// with it, and a reopened child starts again.
public Element Step() {
    Signal<int> n = signal(1);
    template {
        <span>
            $if (n() == 1) one $elseif (n() == 2) two $else many $endif
            <button id="step" onclick={{ () => n.set(n() + 1) }}>next</button>
        </span>
    }
}

public Element App() {
    Signal<bool> open = signal(false);
    template {
        <main>
            {{ parts.Header("Steps & <stuff>") }}
            <button class="toggle" aria-expanded={{ (string) open() }} onclick={{ () => open.set(!open()) }}>toggle</button>
            $if (open())
                {{ Step() }}
            $else
                <em>closed</em>
            $endif
            <p>{{ "<script>alert(1)</script>" }} &amp; &lt;b&gt;</p>
        </main>
    }
}

protected int main() {
    dom.renderApp(App, "#app");
    return 0;
}
