package checker

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AurielSolaris/runescript/internal/ast"
	"github.com/AurielSolaris/runescript/internal/diagnostics"
	"github.com/AurielSolaris/runescript/internal/loader"
	"github.com/AurielSolaris/runescript/internal/types"
	"github.com/AurielSolaris/runescript/stdlib"
)

func isModule(f *loader.File) bool {
	return filepath.Ext(f.Source.Name) == loader.ModuleExt
}

func (c *checker) checkMetadata(f *loader.File) {
	seen := map[string]bool{}
	for _, m := range f.AST.Metadata {
		switch {
		case !isModule(f):
			c.errorf(diagnostics.UseModuleMetadata, m.Span, "module metadata can only be declared in a .rsm file")
		case seen[m.Name]:
			c.errorf(diagnostics.UseModuleMetadata, m.Span, "module metadata '%s' is declared twice", m.Name)
		case m.Name == "version":
			version := m.Value
			if m.Args != nil {
				var ok bool
				version, ok = c.buildVersion(m)
				if !ok {
					break
				}
			}
			if !stdlib.ValidSemver(version) {
				c.errorf(diagnostics.UseModuleMetadata, m.Span, "%q is not a valid SemVer 2.0.0 version", version)
				break
			}
			c.versions[f] = version
		}
		seen[m.Name] = true
	}
	if isModule(f) && !seen["version"] {
		c.errorf(diagnostics.UseModuleMetadata, f.AST.Package.Span, "a .rsm file must declare a public module version").
			WithHelp(`add public semver version = 1.0.0; after the use declarations`)
	}
	if isModule(f) {
		for _, fn := range f.AST.Funcs {
			if fn.Name.Name == "version" {
				c.errorf(diagnostics.UseModuleMetadata, fn.Name.Span, "'version' is reserved for module metadata")
			}
		}
	}
}

func (c *checker) buildVersion(m *ast.ModuleMetadata) (string, bool) {
	args := m.Args
	if len(args) < 3 {
		c.errorf(diagnostics.UseModuleMetadata, m.Span, "semver needs major, minor, and patch integer literals")
		return "", false
	}
	var build []string
	if arr, ok := unparen(args[len(args)-1]).(*ast.ArrayLit); ok && len(args) > 3 {
		for _, item := range arr.Elems {
			id, ok := unparen(item).(*ast.StringLit)
			if !ok || strings.ContainsAny(id.Value, ".+") || !stdlib.ValidSemver("0.0.0+"+id.Value) {
				c.errorf(diagnostics.UseModuleMetadata, item.NodeSpan(), "each build identifier must be one string literal")
				return "", false
			}
			build = append(build, id.Value)
		}
		args = args[:len(args)-1]
	}
	parts := make([]string, len(args))
	for i, arg := range args {
		switch lit := unparen(arg).(type) {
		case *ast.IntLit:
			parts[i] = strconv.FormatUint(lit.Value, 10)
			continue
		case *ast.StringLit:
			if i >= 3 && !strings.ContainsAny(lit.Value, ".+") && stdlib.ValidSemver("0.0.0-"+lit.Value) {
				parts[i] = lit.Value
				continue
			}
		}
		c.errorf(diagnostics.UseModuleMetadata, arg.NodeSpan(), "expected an integer literal for major, minor, or patch, or an integer or string literal for one prerelease identifier")
		return "", false
	}
	version := strings.Join(parts[:3], ".")
	if len(parts) > 3 {
		version += "-" + strings.Join(parts[3:], ".")
	}
	if len(build) > 0 {
		version += "+" + strings.Join(build, ".")
	}
	return version, true
}

func (c *checker) versionPackage(e ast.Expr) *loader.Package {
	sel, ok := unparen(e).(*ast.SelectorExpr)
	if !ok || sel.Name.Name != "version" || !c.isPackageRef(sel) {
		return nil
	}
	b := c.quals[sel.X.(*ast.Ident).Name]
	if b.pkg != nil {
		for _, f := range b.pkg.Files {
			if isModule(f) {
				b.used = true
				return b.pkg
			}
		}
	}
	return nil
}

func (c *checker) moduleVersion(pkg *loader.Package, e, key ast.Expr) *types.Type {
	var name string
	if key != nil {
		switch k := unparen(key).(type) {
		case *ast.StringLit:
			name = k.Value
		case *ast.CharLit:
			name = string(k.Value)
		default:
			c.errorf(diagnostics.UseModuleVersion, key.NodeSpan(), "the module version key must be a file name literal")
			return types.InvalidT
		}
	}
	var version string
	found := false
	for _, f := range pkg.Files {
		if !isModule(f) || key != nil && strings.TrimSuffix(filepath.Base(f.Source.Name), loader.ModuleExt) != name {
			continue
		}
		v, valid := c.versions[f]
		if !valid {
			return types.InvalidT // the declaration has a diagnostic
		}
		if found && key != nil {
			c.errorf(diagnostics.UseModuleVersion, key.NodeSpan(), "package '%s' has several modules named %q", pkg.Name, name)
			return types.InvalidT
		}
		if found && v != version {
			c.errorf(diagnostics.UseModuleVersion, e.NodeSpan(), "package '%s' declares different versions", pkg.Name).
				WithHelp("select a file by name, e.g. geom.version['area']")
			return types.InvalidT
		}
		version, found = v, true
	}
	if !found {
		c.errorf(diagnostics.UseModuleVersion, key.NodeSpan(), "package '%s' has no module named %q", pkg.Name, name)
		return types.InvalidT
	}
	c.info.StringConsts[e] = version
	return types.StringT
}
