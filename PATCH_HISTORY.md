# Rogue Trooper Enhanced PC Patch - Technical History

## Stable base

**V64 is the frozen stable base.**

- Validated V64 executable: `RogueTrooper_V64_EncyclopediaImageScale068_TEST.exe`
- SHA-256: `cb2996b07333df68456afbe992242bba29609e7ffcde05ea5ee625e428103a5d`
- Retail source SHA-256: `f35d8ae6f52cbebf0d1427de5a573d2b8d4be1fe38154b8a7d44cd7ffb6b7525`

Future V65+ work must start from V64 and remain surgical unless a broader change is explicitly justified and tested.

## Final architecture to preserve

### Window / resolution

- Borderless only when game resolution equals desktop resolution.
- Otherwise retain the game's native display path.
- Do not return to permanently forced borderless.
- This resolved the earlier 4:3 image stretching seen on a 16:9 display.

### FOV

- Dynamic Hor+ FOV.
- No fixed FOV tied to one resolution.
- Experimental timer/tick tweak is not part of the final branch.
- VSync is not forcibly disabled.

### HUD / UI

- Prefer local scopes over global transforms.
- Minimap stays anchored bottom-left.
- Health/ammo stays anchored bottom-right.
- Petal Menu uses its validated local correction.
- Top-right Bio-Chip panel uses its validated local anchoring.
- Sniper scope uses the corrected fullscreen behavior.
- Fonts refresh after aspect updates.
- Floating markers use compensated projection and retain their reduced size.

### Reticles

- Reference method: target the actual sprite/render path, including UV/runtime identification where required.
- Important diagnostic example: `reticle_mortar_5_shot`.
- A local factor test proved the circle could be resized without disturbing the Petal Menu.
- Avoid false global parent containers.
- At exactly 0% HUD opacity, the reticle handler produces no visible reticle.
- The clean validated branch comes through V61.

### Alt+F4

- Earlier WndProc / WM_CLOSE / DestroyWindow paths were rejected.
- V63 solution: detect Alt+F4 from the engine's main loop and set the native quit flag.
- Native quit flag: `0x006BEF34`.
- This lets the engine perform its normal shutdown sequence.

### Encyclopedia

- The encyclopedia image has its own local scale.
- Original value: `0.775`.
- V64 value: `0.68`.
- Do not scale the whole menu to fix the image.

## Version chronology

- **Prototype / V1**: first 4K / FOV / borderless Steam branch; established direct EXE modification viability.
- **V2**: early true-borderless and high-resolution startup work.
- **V3**: desktop-based borderless and Alt-Tab work.
- **V4**: early wide FOV proof; FPS experiments later abandoned as policy.
- **V5 / V5A / V5B**: dynamic FOV; V5B retained dynamic FOV without the timer/tick experiment.
- **V6**: first global adaptive-UI pass; useful diagnostically but too broad for final use.
- **V7 / V7 FIXED**: sniper scope and Petal Menu stabilization.
- **V8**: dynamic top-right Bio-Chip placement.
- **V9 / V10**: dynamic fonts and PC menus; V10 became an early cumulative baseline.
- **V11-V20**: reticle audit; abandoned false global parents; local sprite/UV targeting established.
- **V21-V29**: transition from global HUD scaling to local/scoped HUD corrections; V28 stable scoped HUD, V29 additional scope/reticle refinements.
- **V30-V32**: top-left help panel, corner scopes and communication-panel diagnostics.
- **V33-V49**: long resolution/aspect/borderless audit; intrusive AspectFit/intermediate-surface routes rejected; V47 consolidated native resolution flow; V48/V49 established adaptive borderless with native fallback.
- **V50**: reticle-at-HUD-0 diagnostic.
- **V51**: font refresh after aspect update.
- **V52/V53**: screen-scaling and projection audit.
- **V54**: multi-aspect scoped HUD consolidation.
- **V55**: preserve native X/Y pair to avoid asymmetric corrections.
- **V56**: compensated floating-marker projection; stable reference base.
- **V57-V60**: Alt+F4 experiments rejected; V57's reticle-at-HUD-0 correction itself was useful.
- **V61**: clean rebuild from V56 plus only the validated HUD-0 reticle correction.
- **V62**: direct DestroyWindow Alt+F4 attempt rejected.
- **V63**: native main-loop Alt+F4 quit path validated.
- **V64**: V63 plus encyclopedia image scale `0.775 -> 0.68`; validated and frozen as current stable base.

## Explicitly rejected branches / ideas

- Global HUD scaling when a local scope exists.
- Reusing the old V11-V18 reticle branches wholesale.
- V16 specifically caused a Petal Menu regression.
- Alt+F4 experiments V57-V62 as implementations.
- Permanently forced borderless for every resolution.
- Forced VSync disable.
- Reintroducing the experimental timer/tick tweak without a fresh audit.
- Retouching already-correct systems merely to make a higher version number.

## Current validated state at V64

- Borderless / resolution: OK.
- Alt-Tab: OK.
- Alt+F4: OK.
- Dynamic FOV: OK.
- Multi-aspect HUD: OK.
- Petal Menu: OK.
- Top-right Bio-Chip panel: OK.
- Sniper scope: OK.
- Reticles: considered complete.
- HUD 0% reticle hiding: OK.
- Floating markers: OK.
- Fonts / menus: OK.
- Encyclopedia images: OK at scale `0.68`.
- Previous 4:3-on-16:9 stretching: resolved by conditional borderless/native fallback.
