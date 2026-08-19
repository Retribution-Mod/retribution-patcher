<div align='center'>
   <h3>Retribution Patcher</h3>

   IPA patcher to rename, change the icon, remove device enforcement and enable file sharing for Discord, to ease sideloading of Retribution.<br />
   For the resulting IPA and installation instructions, see <a href="https://github.com/Retribution-Mod/retribution-tweak">retribution-tweak</a>. For sideloading, try <a href="https://sidestore.io">SideStore</a> or <a href="https://altstore.io">AltStore</a>.
</div>

---

## Usage

```bash
retribution-patcher \
  --name "Retribution" \
  --icon-zip /path/to/icons.zip \
  --output Retribution.ipa \
  /path/to/discord.ipa
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `--name` | `Retribution` | Display name of the patched app |
| `--icon-zip` | | Path to a local `.zip` of app icons |
| `--icon-url` | | URL to a `.zip` of app icons |
| `--output` | `Retribution.ipa` | Output file name |

## Icon zip format

The icon zip should contain the icon PNGs at the top level, for example:

```
RetributionIcon60x60@2x.png
RetributionIcon76x76@2x~ipad.png
```
