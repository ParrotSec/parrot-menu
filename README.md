# Parrot menu

This package provides Parrot-specific menu definitions, desktop entries, and
launcher integration.

## Desktop launchers

Desktop entries live in two source directories:

- `desktop-files-common/` contains launchers that are suitable for every
  Parrot edition.
- `desktop-files/` contains Parrot Security launchers and is shipped by
  `parrot-menu-security`.

Each managed desktop entry must define:

```ini
X-Parrot-Package=foo
```

`foo` is the binary package that provides the primary command launched by that
desktop entry. The field supports a single binary package only.

`update-launchers` copies managed desktop entries to `/usr/share/applications`
when the related package is installed. If the package is not installed, it
creates a `[not installed]` launcher template that runs `parrot-exec --install
foo`.

If you want to override an existing desktop file instead of adding a new one,
reuse the same filename. When Parrot Menu detects the conflict, it uses
`dpkg-divert` to move the original desktop file away and place the Parrot one
there.

To hide an existing desktop file, override it and add `NoDisplay=true`.

## Edition split

The source package builds two binary packages:

- `parrot-menu` provides the common menu infrastructure, `parrot-exec`,
  `update-launchers`, common desktop launchers, desktop integration files, and
  the apt hook. It must be installed on every Parrot edition.
- `parrot-menu-security` provides the Parrot Security menu, security desktop
  directories, and launcher templates for security tools.

Parrot Home Edition should install only `parrot-menu` by default, so users do
not see installable security tool launchers unless they explicitly opt in.
Parrot Security Edition should install both packages through its metapackage.

An opt-in action on Home Edition should install `parrot-menu-security`; it
should not install all security tools automatically. An opt-out action should
purge `parrot-menu-security`, not just remove it, so the security menu
conffiles are removed from `/etc/xdg/menus/applications-merged`.
