package dom

// The SVG element list. Sources: the element and attribute indexes of SVG
// 1.1 (second edition, https://www.w3.org/TR/SVG11/) and SVG 2
// (https://www.w3.org/TR/SVG2/), and the Filter Effects module for the
// filter primitives. SVG 1.1 elements that SVG 2 removed (the SVG fonts,
// tref, altGlyph, animateColor, ...) are obsolete and warn, as HTML4's do.
//
// Not included: HTML elements SVG 2 allows inside SVG (audio, canvas,
// iframe, video), the unknown element, the draft-only mesh and hatch
// elements, and discard (SVG Animations 2, not widely implemented).

// cssWide are the CSS-wide keywords every presentation attribute accepts.
const cssWide = "|inherit|initial|unset|revert|revert-layer"

// Attribute groups shared by several SVG elements.
const (
	svgLink    = "href:url; !xlink:href:url~use href"
	units      = "(userSpaceOnUse|objectBoundingBox)"
	box        = "x:length; y:length; width:length; height:length"
	primitive  = box + "; result"
	transfer   = "type(identity|table|discrete|linear|gamma); tableValues; slope:float; intercept:float; amplitude:float; exponent:float; offset:float"
	edgeMode   = "edgeMode(duplicate|wrap|none)"
	channel    = "(R|G|B|A)"
	textPos    = "x; y; dx; dy; rotate; textLength:length; lengthAdjust(spacing|spacingAndGlyphs)"
	animTiming = "begin; dur; end; min; max; restart(always|whenNotActive|never); repeatCount; repeatDur; fill(freeze|remove)"
	animValues = "calcMode(discrete|linear|paced|spline); values; keyTimes; keySplines; from; to; by; additive(replace|sum); accumulate(none|sum)"
	animTarget = "attributeName; !attributeType(CSS|XML|auto)~omit it"
	animation  = svgLink + "; " + animTiming + "; " + animValues
	svgFonts   = "use CSS @font-face with WOFF fonts"
)

// svgGlobals are the attributes of every SVG element: the core and
// conditional-processing attributes, and those SVG 2 shares with HTML.
var svgGlobals = mustAttrsNS("id; class:tokens; style; lang; tabindex:int; autofocus:bool; nonce; role:tokens; "+
	"requiredExtensions:tokens; systemLanguage; !requiredFeatures~omit it; "+
	"!xml:space(default|preserve)~use the CSS white-space property; !xml:lang~use lang; !xml:base~omit it; "+
	"!externalResourcesRequired(true|false)~omit it", SVG)

// presentation are the presentation attributes, which SVG 2 allows on
// every SVG element. Their values are CSS values; keyword-only ones are
// checked, including the CSS-wide keywords.
var presentation = mustAttrsNS("alignment-baseline; baseline-shift; clip-path; clip-rule(nonzero|evenodd"+cssWide+"); "+
	"color; color-interpolation(auto|sRGB|linearRGB"+cssWide+"); color-interpolation-filters(auto|sRGB|linearRGB"+cssWide+"); "+
	"color-rendering; cursor; direction(ltr|rtl"+cssWide+"); display; dominant-baseline; fill; fill-opacity; "+
	"fill-rule(nonzero|evenodd"+cssWide+"); filter; flood-color; flood-opacity; font-family; font-size; "+
	"font-size-adjust; font-stretch; font-style; font-variant; font-weight; image-rendering; letter-spacing; "+
	"lighting-color; marker-end; marker-mid; marker-start; mask; mask-type(luminance|alpha"+cssWide+"); opacity; "+
	"overflow(visible|hidden|scroll|auto|clip"+cssWide+"); paint-order; pointer-events; "+
	"shape-rendering(auto|optimizeSpeed|crispEdges|geometricPrecision"+cssWide+"); stop-color; stop-opacity; stroke; "+
	"stroke-dasharray; stroke-dashoffset; stroke-linecap(butt|round|square"+cssWide+"); "+
	"stroke-linejoin(miter|miter-clip|round|bevel|arcs"+cssWide+"); stroke-miterlimit; stroke-opacity; stroke-width; "+
	"text-anchor(start|middle|end"+cssWide+"); text-decoration; text-overflow; "+
	"text-rendering(auto|optimizeSpeed|optimizeLegibility|geometricPrecision"+cssWide+"); transform; transform-origin; "+
	"unicode-bidi; vector-effect(none|non-scaling-stroke|non-scaling-size|non-rotation|fixed-position"+cssWide+"); "+
	"visibility(visible|hidden|collapse"+cssWide+"); white-space; word-spacing; writing-mode; "+
	"!clip~use clip-path; !color-profile~omit it: SVG 2 removed color profiles; "+
	"!enable-background~omit it: SVG 2 removed BackgroundImage; "+
	"!glyph-orientation-horizontal~use the CSS text-orientation property; "+
	"!glyph-orientation-vertical~use the CSS text-orientation property; !kerning~use the CSS font-kerning property", SVG)

// svgDef describes one SVG element.
type svgDef struct {
	name        string
	svg11, svg2 bool
	hint        string // for obsolete elements
	attrs       string
}

func svgBoth(name, attrs string) svgDef {
	return svgDef{name: name, svg11: true, svg2: true, attrs: attrs}
}
func svgNew(name, attrs string) svgDef { return svgDef{name: name, svg2: true, attrs: attrs} }

// svgRemoved is an SVG 1.1 element that SVG 2 removed.
func svgRemoved(name, hint, attrs string) svgDef {
	return svgDef{name: name, svg11: true, hint: hint, attrs: attrs}
}

var svgDefs = []svgDef{
	// ---- in SVG 1.1 and SVG 2 ----
	svgBoth("a", svgLink+"; target; download; hreflang; ping:tokens; referrerpolicy"+referrer+"; rel:tokens; type"),
	svgBoth("animate", animation+"; "+animTarget),
	svgBoth("animateMotion", animation+"; path; keyPoints; rotate; !origin~omit it"),
	svgBoth("animateTransform", animation+"; "+animTarget+"; type(translate|scale|rotate|skewX|skewY)"),
	svgBoth("circle", "cx:length; cy:length; r:length; pathLength:float"),
	svgBoth("clipPath", "clipPathUnits"+units),
	svgBoth("defs", ""),
	svgBoth("desc", ""),
	svgBoth("ellipse", "cx:length; cy:length; rx:length; ry:length; pathLength:float"),
	svgBoth("feBlend", primitive+"; in; in2; mode(normal|multiply|screen|overlay|darken|lighten|color-dodge|"+
		"color-burn|hard-light|soft-light|difference|exclusion|hue|saturation|color|luminosity)"),
	svgBoth("feColorMatrix", primitive+"; in; type(matrix|saturate|hueRotate|luminanceToAlpha); values"),
	svgBoth("feComponentTransfer", primitive+"; in"),
	svgBoth("feComposite", primitive+"; in; in2; operator(over|in|out|atop|xor|lighter|arithmetic); k1:float; k2:float; k3:float; k4:float"),
	svgBoth("feConvolveMatrix", primitive+"; in; order; kernelMatrix; divisor:float; bias:float; targetX:int; targetY:int; "+
		edgeMode+"; kernelUnitLength; preserveAlpha(true|false)"),
	svgBoth("feDiffuseLighting", primitive+"; in; surfaceScale:float; diffuseConstant:float; kernelUnitLength"),
	svgBoth("feDisplacementMap", primitive+"; in; in2; scale:float; xChannelSelector"+channel+"; yChannelSelector"+channel),
	svgBoth("feDistantLight", "azimuth:float; elevation:float"),
	svgBoth("feFlood", primitive),
	svgBoth("feFuncA", transfer), svgBoth("feFuncB", transfer), svgBoth("feFuncG", transfer), svgBoth("feFuncR", transfer),
	svgBoth("feGaussianBlur", primitive+"; in; stdDeviation; "+edgeMode),
	svgBoth("feImage", primitive+"; "+svgLink+"; preserveAspectRatio; crossorigin"+cors),
	svgBoth("feMerge", primitive),
	svgBoth("feMergeNode", "in"),
	svgBoth("feMorphology", primitive+"; in; operator(erode|dilate); radius"),
	svgBoth("feOffset", primitive+"; in; dx:float; dy:float"),
	svgBoth("fePointLight", "x:float; y:float; z:float"),
	svgBoth("feSpecularLighting", primitive+"; in; surfaceScale:float; specularConstant:float; specularExponent:float; kernelUnitLength"),
	svgBoth("feSpotLight", "x:float; y:float; z:float; pointsAtX:float; pointsAtY:float; pointsAtZ:float; "+
		"specularExponent:float; limitingConeAngle:float"),
	svgBoth("feTile", primitive+"; in"),
	svgBoth("feTurbulence", primitive+"; baseFrequency; numOctaves:uint; seed:float; stitchTiles(stitch|noStitch); type(fractalNoise|turbulence)"),
	svgBoth("filter", box+"; filterUnits"+units+"; primitiveUnits"+units+"; !filterRes~omit it; !xlink:href:url~omit it"),
	svgBoth("foreignObject", box),
	svgBoth("g", ""),
	svgBoth("image", box+"; "+svgLink+"; preserveAspectRatio; crossorigin"+cors),
	svgBoth("line", "x1:length; y1:length; x2:length; y2:length; pathLength:float"),
	svgBoth("linearGradient", svgLink+"; x1:length; y1:length; x2:length; y2:length; gradientUnits"+units+"; "+
		"gradientTransform; spreadMethod(pad|reflect|repeat)"),
	svgBoth("marker", "viewBox; preserveAspectRatio; refX:length; refY:length; markerUnits(strokeWidth|userSpaceOnUse); "+
		"markerWidth:length; markerHeight:length; orient"),
	svgBoth("mask", box+"; maskUnits"+units+"; maskContentUnits"+units),
	svgBoth("metadata", ""),
	svgBoth("mpath", svgLink),
	svgBoth("path", "d; pathLength:float"),
	svgBoth("pattern", svgLink+"; "+box+"; patternUnits"+units+"; patternContentUnits"+units+"; patternTransform; viewBox; preserveAspectRatio"),
	svgBoth("polygon", "points; pathLength:float"),
	svgBoth("polyline", "points; pathLength:float"),
	svgBoth("radialGradient", svgLink+"; cx:length; cy:length; r:length; fx:length; fy:length; fr:length; "+
		"gradientUnits"+units+"; gradientTransform; spreadMethod(pad|reflect|repeat)"),
	svgBoth("rect", box+"; rx:length; ry:length; pathLength:float"),
	svgBoth("script", svgLink+"; type; crossorigin"+cors),
	svgBoth("set", svgLink+"; "+animTiming+"; "+animTarget+"; to"),
	svgBoth("stop", "offset"),
	svgBoth("style", "type; media; title"),
	svgBoth("svg", box+"; viewBox; preserveAspectRatio; !version~omit it; !baseProfile~omit it; "+
		"!contentScriptType~omit it; !contentStyleType~omit it; !zoomAndPan(disable|magnify)~omit it"),
	svgBoth("switch", ""),
	svgBoth("symbol", box+"; viewBox; preserveAspectRatio; refX:length; refY:length"),
	svgBoth("text", textPos),
	svgBoth("textPath", svgLink+"; path; startOffset:length; method(align|stretch); spacing(auto|exact); side(left|right); "+
		"textLength:length; lengthAdjust(spacing|spacingAndGlyphs)"),
	svgBoth("title", ""),
	svgBoth("tspan", textPos),
	svgBoth("use", svgLink+"; "+box),
	svgBoth("view", "viewBox; preserveAspectRatio; !viewTarget~omit it; !zoomAndPan(disable|magnify)~omit it"),

	// ---- new in SVG 2 (Filter Effects) ----
	svgNew("feDropShadow", primitive+"; in; dx:float; dy:float; stdDeviation"),

	// ---- SVG 1.1 elements that SVG 2 removed ----
	svgRemoved("altGlyph", "use tspan with CSS fonts", "x; y; dx; dy; rotate; glyphRef; format; xlink:href:url"),
	svgRemoved("altGlyphDef", "use tspan with CSS fonts", ""),
	svgRemoved("altGlyphItem", "use tspan with CSS fonts", ""),
	svgRemoved("animateColor", "use animate", animation+"; attributeName"),
	svgRemoved("color-profile", "omit it: SVG 2 removed color profiles",
		"local; name; rendering-intent(auto|perceptual|relative-colorimetric|saturation|absolute-colorimetric); xlink:href:url"),
	svgRemoved("cursor", "use the CSS cursor property", "x:length; y:length; xlink:href:url"),
	svgRemoved("font", svgFonts, "horiz-origin-x; horiz-origin-y; horiz-adv-x; vert-origin-x; vert-origin-y; vert-adv-y"),
	svgRemoved("font-face", svgFonts, "unicode-range; units-per-em; ascent; descent"),
	svgRemoved("font-face-format", svgFonts, "string"),
	svgRemoved("font-face-name", svgFonts, "name"),
	svgRemoved("font-face-src", svgFonts, ""),
	svgRemoved("font-face-uri", svgFonts, "xlink:href:url"),
	svgRemoved("glyph", svgFonts, "d; unicode; glyph-name; orientation(h|v); arabic-form(initial|medial|terminal|isolated); "+
		"horiz-adv-x; vert-origin-x; vert-origin-y; vert-adv-y"),
	svgRemoved("glyphRef", "use tspan with CSS fonts", "x; y; dx; dy; glyphRef; format; xlink:href:url"),
	svgRemoved("hkern", "use the CSS font-kerning property", "u1; g1; u2; g2; k:float"),
	svgRemoved("missing-glyph", svgFonts, "d; horiz-adv-x; vert-origin-x; vert-origin-y; vert-adv-y"),
	svgRemoved("tref", "put the text in a tspan", "xlink:href:url"),
	svgRemoved("vkern", "use the CSS font-kerning property", "u1; g1; u2; g2; k:float"),
}

var svgElements = buildSVG()

func buildSVG() map[string]*Element {
	m := map[string]*Element{}
	for _, d := range svgDefs {
		if m[d.name] != nil {
			panic("dom: SVG element " + d.name + " listed twice")
		}
		e := &Element{Name: d.name, Namespace: SVG, SVG11: d.svg11, SVG2: d.svg2, Obsolete: !d.svg2, Hint: d.hint,
			Attrs: mustAttrsNS(d.attrs, SVG), Events: map[string]*Attr{}}
		if e.Obsolete != (e.Hint != "") {
			panic("dom: SVG element " + d.name + ": an obsolete element needs a hint, and only it")
		}
		for name := range e.Attrs {
			if svgGlobals[name] != nil || events[name] != nil {
				panic("dom: SVG element " + d.name + " repeats global attribute " + name)
			}
		}
		m[d.name] = e
	}
	return m
}
