package header;

use "rune:dom";

public Element Header(string title) {
    template {
        <header>
            <h1>{{ title }}</h1>
        </header>
    }
}
