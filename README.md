# Parrot menu

Parrot menu provides the Parrot-specific desktop menu integration used by
Parrot GNU/Linux editions.

It ships Freedesktop menu definitions, desktop directory metadata, managed
desktop entries, icon assets, and two helper binaries:

- `update-launchers`, which keeps managed launchers in
  `/usr/share/applications` aligned with installed packages.
- `parrot-exec`, which is the execution wrapper used by Parrot desktop entries.

## Package split

The source package builds two binary packages:

- `parrot-menu` provides the common menu infrastructure, `parrot-exec`,
  `update-launchers`, the complete launcher catalog, menu definitions, desktop
  integration files, and the apt hook. It is suitable for every Parrot edition.
- `parrot-menu-security` provides a policy preset that exposes install-on-demand
  launchers for security tools that are not installed.

Parrot Home Edition should install only `parrot-menu` by default. Users can opt
in to installable security launchers by installing `parrot-menu-security`; this
must not install the security toolset automatically. Removing the package hides
only launchers for tools that are not installed.

Parrot Security Edition should install both packages through its metapackage.

## Repository layout

| Path | Purpose |
| --- | --- |
| `apt.conf.d/` | apt hook that refreshes managed launchers after dpkg runs. |
| `config/` | Package-provided launcher policy presets. |
| `dconf/` | GNOME application folder defaults. |
| `debian/` | Debian packaging metadata. |
| `desktop-directories/` | Freedesktop `.directory` files for menu categories. |
| `desktop-files-common/` | Common launchers shipped by `parrot-menu`. |
| `desktop-files/` | Security launcher templates shipped by `parrot-menu`. |
| `launcher-updater/` | Go source for `update-launchers`. |
| `menu-icons/` | Source and generated menu icons. |
| `menus/` | Freedesktop menu XML files. |
| `parrot-exec/` | Go source for the desktop entry execution wrapper. |

## Managed desktop launchers

Desktop entries managed by `update-launchers` must be stored in either
`desktop-files-common/` or `desktop-files/`, and their filenames must start with
`parrot-` or `serv-`.

Desktop entries tied to an installable package must define:

```ini
X-Parrot-Package=foo
```

`foo` is the binary package that provides the primary command launched by that
desktop entry. The field supports a single binary package only.

Launchers that are always available must omit `X-Parrot-Package` and define:

```ini
X-Parrot-Managed=true
```

During launcher refreshes, `update-launchers` reads
`/var/lib/dpkg/status` directly:

- if `foo` is installed, the real desktop file is always copied to
  `/usr/share/applications`;
- if `foo` is not installed and installable launchers are enabled, a
  `[not installed]` launcher points to `parrot-exec --install foo`;
- if `foo` is not installed and installable launchers are disabled, its managed
  launcher is removed from `/usr/share/applications`.

Package presets are loaded in filename order from
`/usr/share/parrot-menu/config.d/*.conf`. An administrator can override them in
`/etc/parrot-menu/launcher.conf`:

```ini
ShowInstallableLaunchers=false
```

Run `sudo /usr/share/parrot-menu/update-launchers` after changing the local
configuration. `parrot-menu-security` supplies a preset that sets the option to
`true`; `parrot-menu` supplies the default preset set to `false`.

Common launchers should use standard Freedesktop categories. Security launchers
can use Parrot-specific categories defined by the menu files.

To hide a launcher, ship a managed desktop file with `NoDisplay=true`.

## Icon generation

`menu-icons/hicolor/256x256/apps` is the source directory for app icons.

Regenerate all derived icon sizes from the 256x256 sources:

```sh
make icons
```

Import one or more PNG files and generate only their derived sizes:

```sh
make icons IMAGES="/path/to/icon.png /path/to/another-icon.png"
```

For compatibility, `IMAGE=/path/to/icon.png` also works for a single icon.

Imported icons are copied to the 256x256 source directory, resized to 256x256
when needed, then emitted as 16, 22, 24, 32, and 48 pixel variants, including
their `@2` HiDPI versions.

Preview the generated files without modifying the tree:

```sh
python3 generate_icons.py --dry-run
python3 generate_icons.py --dry-run /path/to/icon.png
```

`generate_icons.py` requires Pillow (`python3-pil` on Debian).

## Building

Build the helper binaries:

```sh
make binary
```

This creates:

- `build/update-launchers`
- `build/parrot-exec`
- `build/parrot-ls`, a compatibility symlink to `parrot-exec`

Build the Debian package with the standard packaging workflow:

```sh
dpkg-buildpackage -us -uc
```

## Testing

Run Go checks:

```sh
(cd launcher-updater && go test ./...)
(cd parrot-exec && go test ./...)
```
