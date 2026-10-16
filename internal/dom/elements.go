package dom

// The element list. Sources: the HTML 4.01 element and attribute indexes
// (https://www.w3.org/TR/html401/index/), and the HTML living standard's
// element and attribute indexes and its list of obsolete features
// (https://html.spec.whatwg.org/multipage/indices.html,
// https://html.spec.whatwg.org/multipage/obsolete.html).
//
// Attribute lists use the notation described in dom.go. Global attributes
// and global event handlers apply to every element, including the HTML4
// elements HTML5 removed (HTML 4.01 itself was slightly narrower).

// Keyword lists shared by several attributes.
const (
	referrer   = "(|no-referrer|no-referrer-when-downgrade|same-origin|origin|strict-origin|origin-when-cross-origin|strict-origin-when-cross-origin|unsafe-url)"
	cors       = "(|anonymous|use-credentials)"
	priority   = "(high|low|auto)"
	lazy       = "(lazy|eager)"
	alignBlock = "(left|center|right|justify)"
	alignCell  = "(left|center|right|justify|char)"
	valign     = "(top|middle|bottom|baseline)"
	css        = "use CSS instead"
)

// Attribute groups shared by several elements.
const (
	hyperlink = "href:url; target; download; ping:tokens; rel:tokens; hreflang; type; referrerpolicy" + referrer
	formOwner = "form:idref; name"
	submitter = "formaction:url; formenctype(application/x-www-form-urlencoded|multipart/form-data|text/plain); " +
		"formmethod(get|post|dialog); formnovalidate:bool; formtarget"
	popoverTarget = "popovertarget:idref; popovertargetaction(toggle|show|hide)"
	media         = "src:url; crossorigin" + cors + "; preload(|none|metadata|auto); autoplay:bool; loop:bool; muted:bool; controls:bool"
	cellAlign     = "!align" + alignCell + "~" + css + "; !char~" + css + "; !charoff~" + css + "; !valign" + valign + "~" + css
	heading       = "!align" + alignBlock + "~use CSS text-align"
)

// globals are the global attributes (HTML5; HTML4 had a subset).
var globals = mustAttrs("accesskey:tokens; autocapitalize(off|none|on|sentences|words|characters); autofocus:bool; " +
	"class:tokens; contenteditable(|true|false|plaintext-only); dir(ltr|rtl|auto); draggable(true|false); " +
	"enterkeyhint(enter|done|go|next|previous|search|send); hidden(|hidden|until-found); id; inert:bool; " +
	"inputmode(none|text|decimal|numeric|tel|search|email|url); itemid:url; itemprop:tokens; itemref:idrefs; " +
	"itemscope:bool; itemtype:tokens; lang; nonce; popover(|auto|manual|hint); role:tokens; slot; " +
	"spellcheck(|true|false); style; tabindex:int; title; translate(|yes|no); writingsuggestions(|true|false)")

// events are the global event handlers.
var events = eventAttrs(`onabort onauxclick onbeforeinput onbeforetoggle onblur oncancel oncanplay
	oncanplaythrough onchange onclick onclose oncontextmenu oncopy oncuechange oncut ondblclick ondrag
	ondragend ondragenter ondragleave ondragover ondragstart ondrop ondurationchange onemptied onended
	onerror onfocus onfocusin onfocusout onformdata oninput oninvalid onkeydown onkeypress onkeyup onload
	onloadeddata onloadedmetadata onloadstart onmousedown onmouseenter onmouseleave onmousemove onmouseout
	onmouseover onmouseup onpaste onpause onplay onplaying onpointercancel onpointerdown onpointerenter
	onpointerleave onpointermove onpointerout onpointerover onpointerup onprogress onratechange onreset
	onresize onscroll onscrollend onseeked onseeking onselect onslotchange onstalled onsubmit onsuspend
	ontimeupdate ontoggle ontouchcancel ontouchend ontouchmove ontouchstart onvolumechange onwaiting onwheel`)

// windowEvents are body's own event handlers (the window's events).
const windowEvents = `onafterprint onbeforeprint onbeforeunload onhashchange onlanguagechange onmessage
	onmessageerror onoffline ononline onpagehide onpageshow onpopstate onrejectionhandled onstorage
	onunhandledrejection onunload`

// def describes one element.
type def struct {
	name         string
	html4, html5 bool
	hint         string // for obsolete elements
	attrs        string
	events       string
}

func both(name, attrs string) def  { return def{name: name, html4: true, html5: true, attrs: attrs} }
func html5(name, attrs string) def { return def{name: name, html5: true, attrs: attrs} }

// removed is an HTML4 element that HTML5 made obsolete.
func removed(name, hint, attrs string) def {
	return def{name: name, html4: true, hint: hint, attrs: attrs}
}

// voids are the elements without content or end tag.
var voids = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true, "hr": true, "img": true,
	"input": true, "link": true, "meta": true, "source": true, "track": true, "wbr": true,
	// HTML4 only
	"basefont": true, "frame": true, "isindex": true, "param": true,
}

var defs = []def{
	// ---- in HTML4 and HTML5 ----
	both("a", hyperlink+"; !charset~"+"specify the encoding with HTTP"+"; !coords~use area for image maps; "+
		"!name~use the id attribute; !rev~use rel with the opposite term; !shape~use area for image maps"),
	both("abbr", ""),
	both("address", ""),
	both("area", "alt; coords; shape(circle|default|poly|rect); "+hyperlink+"; !nohref:bool~omit the href attribute"),
	both("b", ""),
	both("base", "href:url; target"),
	both("bdo", ""),
	both("blockquote", "cite:url"),
	{name: "body", html4: true, html5: true, events: windowEvents,
		attrs: "!alink~" + css + "; !background~" + css + "; !bgcolor~" + css + "; !link~" + css + "; !text~" + css + "; !vlink~" + css},
	both("br", "!clear(left|all|right|none)~use CSS clear"),
	both("button", "disabled:bool; "+formOwner+"; "+submitter+"; "+popoverTarget+"; type(submit|reset|button); value"),
	both("caption", "!align(top|bottom|left|right)~use CSS caption-side or text-align"),
	both("cite", ""),
	both("code", ""),
	both("col", "span:uint; "+cellAlign+"; !width~"+css),
	both("colgroup", "span:uint; "+cellAlign+"; !width~"+css),
	both("dd", ""),
	both("del", "cite:url; datetime"),
	both("dfn", ""),
	both("div", "!align"+alignBlock+"~"+css),
	both("dl", "!compact:bool~"+css),
	both("dt", ""),
	both("em", ""),
	both("fieldset", "disabled:bool; "+formOwner),
	both("form", "accept-charset:tokens; action:url; autocomplete(on|off); "+
		"enctype(application/x-www-form-urlencoded|multipart/form-data|text/plain); method(get|post|dialog); "+
		"name; novalidate:bool; target; rel:tokens; !accept~use the accept attribute of input"),
	both("h1", heading), both("h2", heading), both("h3", heading),
	both("h4", heading), both("h5", heading), both("h6", heading),
	both("head", "!profile:url~use meta name values"),
	both("hr", "!align(left|center|right)~"+css+"; !noshade:bool~"+css+"; !size~"+css+"; !width~"+css),
	both("html", "!version~omit it"),
	both("i", ""),
	both("iframe", "src:url; srcdoc; name; sandbox:tokens; allow; allowfullscreen:bool; width:uint; height:uint; "+
		"referrerpolicy"+referrer+"; loading"+lazy+"; !align~"+css+"; !frameborder(1|0)~use CSS border; "+
		"!longdesc:url~link to the description; !marginheight~"+css+"; !marginwidth~"+css+"; !scrolling(yes|no|auto)~"+css),
	both("img", "alt; src:url; srcset; sizes; crossorigin"+cors+"; usemap; ismap:bool; width:uint; height:uint; "+
		"referrerpolicy"+referrer+"; decoding(sync|async|auto); loading"+lazy+"; fetchpriority"+priority+"; "+
		"!align~"+css+"; !border~"+css+"; !hspace~"+css+"; !vspace~"+css+"; !longdesc:url~link to the description; "+
		"!name~use the id attribute"),
	both("input", "accept:tokens; alt; autocomplete:tokens; checked:bool; dirname; disabled:bool; "+formOwner+"; "+
		submitter+"; height:uint; list:idref; max; maxlength:uint; min; minlength:uint; multiple:bool; pattern; "+
		"placeholder; "+popoverTarget+"; readonly:bool; required:bool; size:uint; src:url; step; "+
		"type(button|checkbox|color|date|datetime-local|email|file|hidden|image|month|number|password|radio|"+
		"range|reset|search|submit|tel|text|time|url|week); value; width:uint; "+
		"!align~"+css+"; !usemap~use img for image maps"),
	both("ins", "cite:url; datetime"),
	both("kbd", ""),
	both("label", "for:idref"),
	both("legend", "!align(top|bottom|left|right)~"+css),
	both("li", "value:int; !type~use CSS list-style-type"),
	both("link", "href:url; crossorigin"+cors+"; rel:tokens; media; integrity; hreflang; type; "+
		"referrerpolicy"+referrer+"; sizes:tokens; imagesrcset; imagesizes; as; blocking:tokens; color; "+
		"disabled:bool; fetchpriority"+priority+"; !charset~specify the encoding with HTTP; "+
		"!rev~use rel with the opposite term; !target~omit it"),
	both("map", "name"),
	both("menu", "!compact:bool~"+css),
	both("meta", "name; content; charset; media; "+
		"http-equiv(content-language|content-type|default-style|refresh|set-cookie|x-ua-compatible|content-security-policy); "+
		"!scheme~put the scheme in the value"),
	both("noscript", ""),
	both("object", "data:url; type; "+formOwner+"; width:uint; height:uint; !align~"+css+"; "+
		"!archive~use data and type; !border~"+css+"; !classid~use data and type; !code~use data and type; "+
		"!codebase~use data and type; !codetype~use data and type; !declare:bool~repeat the object; "+
		"!hspace~"+css+"; !standby~optimize the resource to load quickly; !vspace~"+css),
	both("ol", "reversed:bool; start:int; type(1|a|A|i|I); !compact:bool~"+css),
	both("optgroup", "disabled:bool; label"),
	both("option", "disabled:bool; label; selected:bool; value"),
	both("p", "!align"+alignBlock+"~"+css),
	both("pre", "!width~"+css),
	both("q", "cite:url"),
	both("s", ""),
	both("samp", ""),
	both("script", "src:url; type; nomodule:bool; async:bool; defer:bool; crossorigin"+cors+"; integrity; "+
		"referrerpolicy"+referrer+"; blocking:tokens; fetchpriority"+priority+"; "+
		"!charset~omit it: scripts are UTF-8; !language~use type, or omit it for JavaScript"),
	both("select", "autocomplete:tokens; disabled:bool; "+formOwner+"; multiple:bool; required:bool; size:uint"),
	both("small", ""),
	both("span", ""),
	both("strong", ""),
	both("style", "media; blocking:tokens"),
	both("sub", ""),
	both("sup", ""),
	both("table", "!align(left|center|right)~"+css+"; !bgcolor~"+css+"; !border~"+css+"; !cellpadding~"+css+"; "+
		"!cellspacing~"+css+"; !frame(void|above|below|hsides|lhs|rhs|vsides|box|border)~"+css+"; "+
		"!rules(none|groups|rows|cols|all)~"+css+"; !summary~use caption or text around the table; !width~"+css),
	both("tbody", cellAlign),
	both("td", "colspan:uint; rowspan:uint; headers:idrefs; "+cellAlign+"; !abbr~put the abbreviation in the cell, or use th; "+
		"!axis~use scope on th; !bgcolor~"+css+"; !height~"+css+"; !nowrap:bool~"+css+"; "+
		"!scope(row|col|rowgroup|colgroup)~use th; !width~"+css),
	both("textarea", "autocomplete:tokens; cols:uint; dirname; disabled:bool; "+formOwner+"; maxlength:uint; "+
		"minlength:uint; placeholder; readonly:bool; required:bool; rows:uint; wrap(soft|hard)"),
	both("tfoot", cellAlign),
	both("th", "colspan:uint; rowspan:uint; headers:idrefs; scope(row|col|rowgroup|colgroup); abbr; "+cellAlign+"; "+
		"!axis~use scope; !bgcolor~"+css+"; !height~"+css+"; !nowrap:bool~"+css+"; !width~"+css),
	both("thead", cellAlign),
	both("title", ""),
	both("tr", cellAlign+"; !bgcolor~"+css),
	both("u", ""),
	both("ul", "!compact:bool~"+css+"; !type~use CSS list-style-type"),
	both("var", ""),

	// ---- new in HTML5 ----
	html5("article", ""),
	html5("aside", ""),
	html5("audio", media),
	html5("bdi", ""),
	html5("canvas", "width:uint; height:uint"),
	html5("data", "value"),
	html5("datalist", ""),
	html5("details", "open:bool; name"),
	html5("dialog", "open:bool; closedby(any|closerequest|none)"),
	html5("embed", "src:url; type; width:uint; height:uint"),
	html5("figcaption", ""),
	html5("figure", ""),
	html5("footer", ""),
	html5("header", ""),
	html5("hgroup", ""),
	html5("main", ""),
	html5("mark", ""),
	html5("meter", "value:float; min:float; max:float; low:float; high:float; optimum:float"),
	html5("nav", ""),
	html5("output", "for:idrefs; "+formOwner),
	html5("picture", ""),
	html5("progress", "value:float; max:float"),
	html5("rp", ""),
	html5("rt", ""),
	html5("ruby", ""),
	html5("search", ""),
	html5("section", ""),
	html5("slot", "name"),
	html5("source", "type; media; src:url; srcset; sizes; width:uint; height:uint"),
	html5("summary", ""),
	html5("template", "shadowrootmode(open|closed); shadowrootdelegatesfocus:bool; shadowrootclonable:bool; shadowrootserializable:bool"),
	html5("time", "datetime"),
	html5("track", "default:bool; kind(subtitles|captions|descriptions|chapters|metadata); label; src:url; srclang"),
	html5("video", media+"; poster:url; playsinline:bool; width:uint; height:uint"),
	html5("wbr", ""),

	// ---- HTML4 elements that HTML5 made obsolete ----
	removed("acronym", "use abbr", ""),
	removed("applet", "use object or embed", "code; codebase:url; archive; object; alt; name; width:uint; height:uint; align; hspace:uint; vspace:uint"),
	removed("basefont", css, "size; color; face"),
	removed("big", "use CSS font-size", ""),
	removed("center", "use CSS text-align, or margin: auto", ""),
	removed("dir", "use ul", "compact:bool"),
	removed("font", css, "size; color; face"),
	removed("frame", "use iframe, or CSS for the layout", "src:url; name; longdesc:url; frameborder(1|0); "+
		"marginwidth:uint; marginheight:uint; noresize:bool; scrolling(yes|no|auto)"),
	{name: "frameset", html4: true, hint: "use iframe, or CSS for the layout", attrs: "rows; cols", events: "onload onunload"},
	removed("isindex", "use a form with an input", "prompt"),
	removed("noframes", "omit it: without frameset there is nothing to replace", ""),
	removed("param", "use the data attribute of object", "name; value; valuetype(data|ref|object); type"),
	removed("strike", "use s or del", ""),
	removed("tt", "use code, kbd, samp, or CSS", ""),
}

var elements = buildElements()

func buildElements() map[string]*Element {
	m := map[string]*Element{}
	for _, d := range defs {
		if m[d.name] != nil {
			panic("dom: element " + d.name + " listed twice")
		}
		e := &Element{Name: d.name, Namespace: HTML, HTML4: d.html4, HTML5: d.html5, Void: voids[d.name],
			Obsolete: !d.html5, Hint: d.hint, Attrs: mustAttrs(d.attrs), Events: eventAttrs(d.events)}
		if e.Obsolete != (e.Hint != "") {
			panic("dom: element " + d.name + ": an obsolete element needs a hint, and only it")
		}
		for name := range e.Attrs {
			if globals[name] != nil || events[name] != nil {
				panic("dom: element " + d.name + " repeats global attribute " + name)
			}
		}
		m[d.name] = e
	}
	return m
}
