package main;

use "rune:dom";

// A static page: no runes, just elements and attributes. Every element
// and attribute is checked against the HTML element list when the program
// compiles; try changing <ul> to <ull>, or href to hraf.
//
//     runescript dev examples/web/hello/app.rsc

public Element Hello() {
    template {
        <article class="card">
            <h1>Hello from RuneScript</h1>
            <p>This page was built by a <strong>RuneScript</strong> component.</p>
            <ul>
                <li>Elements are checked at compile time</li>
                <li>Text is escaped: {{ "<b>not bold</b>" }}</li>
                <li>Character references work: &copy; 2026 &amp; beyond</li>
            </ul>
            <a href="https://github.com/AurielSolaris/runescript" target="_blank" rel="noopener">The repository</a>
        </article>
    }
}

public void main() {
    dom.renderApp(Hello, "#app");
}
