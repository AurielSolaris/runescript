package main;

use "rune:dom";

// click: #details
// click: #details

// $if ... $else ... $endif shows one branch at a time. A branch that goes
// away takes its bindings with it.
//
//     runescript dev examples/web/toggle/app.rsc

public Element Details() {
    Signal<bool> open = signal(false);

    template {
        <div class="details">
            <button id="details" type="button" aria-expanded={{ (string) open() }} aria-controls="panel"
                    onclick={{ () => open.set(!open()) }}>
                $if (open()) Hide details $else Show details $endif
            </button>
            $if (open())
                <div id="panel" role="region">
                    <p>RuneScript checks every element and attribute against the HTML element list.</p>
                </div>
            $else
                <p><em>Details are hidden.</em></p>
            $endif
        </div>
    }
}

public void main() {
    dom.renderApp(Details, "#app");
}
