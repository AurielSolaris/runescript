package main;

use "rune:dom";
use "./app.rsc";

// click: #add

// The entry file only mounts the app: no router, one component, one
// element. A component of another file is reached through that file's
// package qualifier, here app.App.
//
// Run it from this folder, so that "./app.rsc" resolves here:
//
//     cd examples/web/multi-file
//     runescript dev main.rsc

public void main() {
    dom.renderApp(app.App, "#app");
}
