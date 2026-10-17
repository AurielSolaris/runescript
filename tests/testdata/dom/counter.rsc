package main;

use "rune:dom";

// click: #inc
// click: #inc
// click: #reset

public Element Counter() {
    Signal<int> count = signal(0);
    private void increment() {
        count.set(count() + 1);
    }

    template {
        <div class="counter">
            <p>Clicked {{ count() }} times</p>
            <button id="inc" onclick={{ increment }}>+1</button>
            <button id="reset" onclick={{ () => count.set(0) }} disabled={{ count() == 0 }}>Reset</button>
            $if (count() > 1)
                <p class="note">That's a lot of clicks.</p>
            $endif
        </div>
    }
}

protected int main() {
    dom.renderApp(Counter, "#app");
    return 0;
}
