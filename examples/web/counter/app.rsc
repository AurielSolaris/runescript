package main;

use "rune:dom";

// click: #increment
// click: #increment
// click: #increment
// click: #reset

// A rune (a signal) holds state. Reading it inside the template, as
// {{ count() }}, ties that piece of the page to it: setting the rune
// updates exactly the text, attribute, or $if that read it.
//
//     runescript dev examples/web/counter/app.rsc

public Element Counter() {
    Signal<int> count = signal(0);

    // Nested functions may use the component's runes.
    private void increment() {
        count.set(count() + 1);
    }

    template {
        <section class="counter">
            <h2>Clicked {{ count() }} times</h2>
            <button id="increment" type="button" onclick={{ increment }}>+1</button>
            <button id="reset" type="button" disabled={{ count() == 0 }} onclick={{ () => count.set(0) }}>Reset</button>
            $if (count() >= 3)
                <p class="hint">Three or more: the reset button is waiting.</p>
            $endif
        </section>
    }
}

public void main() {
    dom.renderApp(Counter, "#app");
}
