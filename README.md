# Rogue Trooper - Enhanced PC Patch

![Rogue Trooper Enhanced PC Patch](media/Image%20Codex%2023%20sept.%202026%2C%2015_39_41.png)

A cumulative compatibility and presentation patch for the original PC release of **Rogue Trooper**.

**Current stable build: V64**

## Main features

- 4K and ultrawide support.
- Dynamic Hor+ field of view.
- Conditional borderless window behavior with native fallback.
- HUD and UI corrections across modern aspect ratios.
- Alt-Tab and native Alt+F4 fixes.
- Reticle and sniper scope corrections.
- Improved font and menu scaling.
- Corrected floating HUD markers and encyclopedia image layout.

The patcher is intentionally conservative: it only accepts the exact supported retail executable, creates a verified backup, reconstructs the validated V64 executable, verifies its SHA-256 before installation, and aborts without installing an unverified output if anything differs.

## What V64 fixes

- Dynamic Hor+ field of view for modern aspect ratios.
- Conditional borderless mode when the in-game resolution matches the desktop resolution.
- Native display path retained when the game resolution differs from the desktop, avoiding the old 4:3-on-16:9 stretch regression.
- Working Alt-Tab behavior.
- Native Alt+F4 shutdown path through the engine's own quit flag.
- Multi-aspect HUD/UI scaling with local, scoped corrections rather than one global multiplier.
- Correct placement for minimap, health/ammo, Petal Menu and Bio-Chip panel.
- Correct fullscreen sniper scope behavior.
- Dynamic font/menu refresh after aspect-ratio changes.
- Corrected floating/projected HUD markers while retaining their reduced size.
- Reticle handling corrected, including complete hiding when HUD opacity is exactly 0%.
- Encyclopedia image overlap fixed by changing only its local image scale from `0.775` to `0.68`.

V64 does **not** force VSync off and does **not** include the abandoned experimental timer/tick tweak.

## Installation

1. Download the latest release archive.
2. Extract `RogueTrooper_Patcher_V64.exe` into the game folder, next to `RogueTrooper.exe`.
3. Run the patcher.
4. Keep `RogueTrooper.exe.Backup`. It is the verified vanilla executable.

The patcher accepts only this retail executable:

- Size: `963584` bytes
- SHA-256: `f35d8ae6f52cbebf0d1427de5a573d2b8d4be1fe38154b8a7d44cd7ffb6b7525`

The installed V64 executable must be exactly:

- Size: `2959360` bytes
- SHA-256: `cb2996b07333df68456afbe992242bba29609e7ffcde05ea5ee625e428103a5d`

If either check fails, the patcher stops and does not intentionally install an unverified executable.

## Why the patched EXE is larger

The retail executable is packed: its `.text` section occupies about `0xE8000` bytes on disk while mapping a much larger code image. V64 is an unpacked/rebuilt PE image and includes a small `.rtfix` code section used by the cumulative fixes. The size difference is therefore expected.

## Source and reproducibility

The repository contains:

- `patcher/main.go`: standalone patcher source.
- `patcher/rogue_v64.rtdp1.zlib`: source-locked binary patch payload.
- `tools/make_patch.py`: deterministic payload generator.
- `build_windows.bat`: local Windows build helper.
- `.github/workflows/build.yml`: reproducible GitHub Actions build.
- `PATCH_HISTORY.md`: cumulative technical history and rejected branches.

The patch payload is tied to the exact retail SHA-256 and the patcher independently verifies the final V64 SHA-256 before installation.

## Design rule

Rogue Trooper proved unusually sensitive to broad engine changes. The project therefore prefers local, surgical corrections. Future versions should start from V64 and should not revive rejected global HUD transforms, forced-borderless paths, or abandoned Alt+F4 experiments without a new audit.
