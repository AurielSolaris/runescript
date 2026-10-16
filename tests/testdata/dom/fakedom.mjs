// A small fake DOM for testing rune:dom programs on Node.js: just what the
// runtime uses, plus serialization and clicks. It is not a browser; the
// dev-server check in Edge covers a real one.

const VOID = new Set(["area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "source", "track", "wbr"]);

class Node {
  constructor() {
    this.parentNode = null;
    this.childNodes = [];
  }
  get firstChild() {
    return this.childNodes[0] || null;
  }
  appendChild(n) {
    return this.insertBefore(n, null);
  }
  insertBefore(n, ref) {
    const nodes = n instanceof DocumentFragment ? n.childNodes.slice() : [n];
    for (const x of nodes) {
      if (x.parentNode) x.parentNode.removeChild(x);
      const i = ref === null ? this.childNodes.length : this.childNodes.indexOf(ref);
      if (i < 0) throw new Error("insertBefore: reference is not a child");
      this.childNodes.splice(i, 0, x);
      x.parentNode = this;
    }
    return n;
  }
  removeChild(n) {
    const i = this.childNodes.indexOf(n);
    if (i < 0) throw new Error("removeChild: not a child");
    this.childNodes.splice(i, 1);
    n.parentNode = null;
    return n;
  }
  remove() {
    if (this.parentNode) this.parentNode.removeChild(this);
  }
  get textContent() {
    return this.childNodes.map((c) => c.textContent).join("");
  }
}

class Text extends Node {
  constructor(data) {
    super();
    this.data = String(data);
  }
  get textContent() {
    return this.data;
  }
}

class Comment extends Node {
  constructor(data) {
    super();
    this.data = data;
  }
  get textContent() {
    return "";
  }
}

class DocumentFragment extends Node {}

class Element extends Node {
  constructor(tag, ns) {
    super();
    this.tagName = tag;
    this.namespaceURI = ns;
    this.attrs = new Map();
    this.listeners = {};
  }
  setAttribute(name, value) {
    this.attrs.set(name, String(value));
  }
  getAttribute(name) {
    return this.attrs.has(name) ? this.attrs.get(name) : null;
  }
  removeAttribute(name) {
    this.attrs.delete(name);
  }
  addEventListener(type, fn) {
    (this.listeners[type] = this.listeners[type] || []).push(fn);
  }
  dispatch(type) {
    for (const fn of this.listeners[type] || []) fn({ type, target: this });
  }
  click() {
    this.dispatch("click");
  }
}

function matches(e, sel) {
  if (sel.startsWith("#")) return e.getAttribute("id") === sel.slice(1);
  if (sel.startsWith(".")) return (e.getAttribute("class") || "").split(/\s+/).includes(sel.slice(1));
  return e.tagName === sel;
}

function find(n, sel) {
  for (const c of n.childNodes) {
    if (c instanceof Element && matches(c, sel)) return c;
    const r = find(c, sel);
    if (r) return r;
  }
  return null;
}

const escapeText = (s) => s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
const escapeAttr = (s) => s.replace(/&/g, "&amp;").replace(/"/g, "&quot;");

// serialize renders a node as HTML, leaving out the runtime's anchors.
export function serialize(n) {
  if (n instanceof Text) return escapeText(n.data);
  if (n instanceof Comment) return "";
  if (!(n instanceof Element)) return n.childNodes.map(serialize).join("");
  const attrs = Array.from(n.attrs, ([k, v]) => (v === "" ? " " + k : ` ${k}="${escapeAttr(v)}"`)).join("");
  const ns = n.namespaceURI === "http://www.w3.org/2000/svg" && n.tagName === "svg" ? "" : "";
  if (VOID.has(n.tagName) && n.namespaceURI === null) return `<${n.tagName}${ns}${attrs}>`;
  return `<${n.tagName}${attrs}>${n.childNodes.map(serialize).join("")}</${n.tagName}>`;
}

export function install() {
  const body = new Element("body", null);
  const document = {
    body,
    createElement: (tag) => new Element(tag, null),
    createElementNS: (ns, tag) => new Element(tag, ns),
    createTextNode: (s) => new Text(s),
    createComment: (s) => new Comment(s),
    createDocumentFragment: () => new DocumentFragment(),
    querySelector: (sel) => find(body, sel),
  };
  globalThis.document = document;
  return document;
}
