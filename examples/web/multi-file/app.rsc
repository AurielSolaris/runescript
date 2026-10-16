package app;

use "rune:dom";
use "./header.rsc";

// App puts a header (from another file) above a small list. Child
// components are embedded with {{ Component() }}.

public Element App() {
    Signal<int> items = signal(1);

    template {
        <main>
            {{ header.Header("Shopping list") }}
            <p>{{ items() }} item(s) planned.</p>
            <button id="add" type="button" onclick={{ () => items.set(items() + 1) }}>Add one</button>
        </main>
    }
}
