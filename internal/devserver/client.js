// The runescript dev client: `runescript dev` injects it into the page. It
// reloads the page after each successful rebuild, shows compile errors in
// an overlay instead, and logs warnings to the console.
(() => {
  "use strict";
  let server = null, version = -1, box = null;

  function hide() {
    if (box) box.remove();
    box = null;
  }

  // showError shows text (diagnostics, as plain text) in an overlay.
  function showError(title, text) {
    hide();
    box = document.createElement("div");
    box.setAttribute("role", "alertdialog");
    box.setAttribute("aria-modal", "true");
    box.setAttribute("aria-label", title);
    box.style.cssText = "position:fixed;inset:0;z-index:2147483647;overflow:auto;padding:24px;" +
      "background:rgba(20,22,26,.94);color:#f0f3f6;font:14px/1.5 ui-monospace,SFMono-Regular,Menlo,Consolas,monospace";
    const head = document.createElement("div");
    head.style.cssText = "display:flex;justify-content:space-between;align-items:center;gap:16px;margin-bottom:12px";
    const h = document.createElement("strong");
    h.textContent = "runescript dev: " + title;
    h.style.color = "#ff938a";
    const close = document.createElement("button");
    close.type = "button";
    close.textContent = "Close (Esc)";
    close.style.cssText = "font:inherit;padding:4px 12px;cursor:pointer";
    close.addEventListener("click", hide);
    head.append(h, close);
    const pre = document.createElement("pre");
    pre.textContent = text;
    pre.style.cssText = "margin:0;white-space:pre-wrap";
    const note = document.createElement("p");
    note.textContent = "Fix the problem and save: the page reloads by itself.";
    note.style.cssText = "color:#9198a1;margin-top:16px";
    box.append(head, pre, note);
    document.body.append(box);
    close.focus();
  }

  addEventListener("keydown", (e) => {
    if (e.key === "Escape") hide();
  });
  window.__runescript = { showError };

  const events = new EventSource("/__runescript/events");
  events.addEventListener("build", (m) => {
    const ev = JSON.parse(m.data);
    const changed = server !== null && (ev.server !== server || ev.version !== version);
    server = ev.server;
    version = ev.version;
    if (ev.ok && changed) {
      location.reload(); // the reloaded page logs the warnings
      return;
    }
    for (const w of ev.warnings) console.warn(w);
    if (!ev.ok) showError("Compile error", ev.errors.join("\n"));
  });
})();
