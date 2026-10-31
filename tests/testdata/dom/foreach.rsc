package main;

use "rune:dom";

// click: #more
// click: #more
// click: #clear

public Array<string> fruits(int n) {
    Array<string> all = ["apple", "banana", "cherry", "date"];
    Array<string> out = [];
    for (int i = 0; i < n && i < all.length; i++) {
        out.push(all[i]);
    }
    return out;
}

public Element List() {
    Signal<int> count = signal(1);
    Array<Array<int>> grid = [[1, 2], [3]];

    template {
        <div>
            <ul>
                $foreach (string f in fruits(count()))
                    <li class={{ f }}>
                        {{ f }}
                        $if (f == "banana")
                            (yellow)
                        $endif
                    </li>
                $endforeach
            </ul>
            $if (count() > 0)
                <p>
                    $foreach (Array<int> row in grid)
                        $foreach (int x in row)<b>{{ x }}</b>$endforeach;
                    $endforeach
                </p>
            $endif
            <button id="more" onclick={{ () => count.set(count() + 1) }}>More</button>
            <button id="clear" onclick={{ () => count.set(0) }}>Clear</button>
        </div>
    }
}

protected int main() {
    dom.renderApp(List, "#app");
    return 0;
}
