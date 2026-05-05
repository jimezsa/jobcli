---
name: pdf-cv-to-latex
description: Convert a CV/resume PDF into a styled LaTeX (.tex) source that visually matches the original. Use when the user supplies a PDF CV and asks for a LaTeX version, asks to recreate a CV's design in LaTeX, or wants to clone a resume's color/layout/photo treatment.
---

# Convert a PDF CV to LaTeX

Goal: produce a `.tex` file that compiles to a PDF visually close to the source — same content, same section order, same accent color, same header treatment (photo, gradient, contact icons), and the original language.

Work in this order. Every step exists to remove a guess.

## 1. Verify tooling

Required programs:

- Poppler CLI tools: `pdftotext`, `pdftoppm`, `pdfimages`
- Python 3 with Pillow
- `pdflatex` from TeX Live, MacTeX, or MiKTeX, with common LaTeX packages (`geometry`, `xcolor`, `hyperref`, `tikz`/`pgf`, `babel`, `lmodern`)

These commands must work before you begin:

```bash
pdftotext -v          # poppler — text extraction
pdftoppm -v           # poppler — page rendering
pdfimages -v          # poppler — embedded image extraction
pdflatex --version    # LaTeX compiler
kpsewhich tikz.sty    # TikZ/PGF package lookup
kpsewhich lmodern.sty # Latin Modern fonts lookup
python3 -c "from PIL import Image; print(Image.__version__)"  # color sampling
```

On Windows PowerShell, use `py -3` for the Python check:

```powershell
pdftotext -v
pdftoppm -v
pdfimages -v
pdflatex --version
kpsewhich tikz.sty
kpsewhich lmodern.sty
py -3 -c "from PIL import Image; print(Image.__version__)"
```

Install missing tools before continuing.

macOS with Homebrew:

```bash
brew install poppler
brew install --cask mactex-no-gui  # only if pdflatex is missing; large download
python3 -m pip install --user Pillow
```

Linux, Debian/Ubuntu:

```bash
sudo apt update
sudo apt install -y poppler-utils python3 python3-pil lmodern \
  texlive-latex-recommended texlive-latex-extra texlive-fonts-recommended \
  texlive-pictures
```

Linux, Fedora:

```bash
sudo dnf install -y poppler-utils python3-pillow texlive-lm \
  texlive-scheme-medium texlive-pgf
```

If a Linux distro package set still fails the `kpsewhich` checks, install the distro's full TeX Live package set or the upstream TeX Live distribution, then rerun the checks.

Windows 10/11, PowerShell:

```powershell
winget install -e --id oschwartz10612.Poppler
winget install -e --id MiKTeX.MiKTeX
winget install -e --id Python.Python.3.12
py -3 -m pip install --user Pillow
```

If `winget` is unavailable, install Python from python.org, MiKTeX from miktex.org, and a Poppler Windows release zip manually. Add the folder containing `pdftotext.exe`, `pdftoppm.exe`, and `pdfimages.exe` to `PATH`.

Restart the terminal after Windows installs so `PATH` updates are visible. If MiKTeX prompts to install missing LaTeX packages while compiling, accept; for non-interactive work, open MiKTeX Console and enable automatic package installation or preinstall the missing package named in the `.log`.

For CVs outside the default English setup, install the matching TeX Live/MiKTeX language package (`texlive-lang-spanish`, `texlive-lang-french`, etc.) before changing the `babel` option.

Do not proceed without these tools. Color and layout choices come from sampling, not from looking at the rendered page in chat — display rendering misleads.

## 2. Extract content

```bash
pdftotext -layout path/to/cv.pdf /tmp/cv.txt
```

`-layout` preserves columns and date alignment. Read the result and identify:

- Language (English/Spanish/French/…) — keep it. Add the matching `babel` option, such as `\usepackage[spanish]{babel}`.
- Section headings in the original order
- Each entry's three fields: institution/employer, role/program, date range
- Bullet lines under each entry
- Any free-text sections (Skills list, Languages, Hobbies)

## 3. Extract image assets if present

```bash
pdfimages -all path/to/cv.pdf assets/<name>/img
ls -la assets/<name>/         # the largest PNG/JPG is usually the portrait
```

If the CV has a portrait, keep the largest image, rename it to `photo.png` (or `.jpg`), and verify dimensions with PIL; near-square (e.g. 474×490) clips cleanly to a circle. If the CV has no portrait, skip photo handling entirely. Tiny images (≈100–3000 bytes) are usually decorative shapes/icons — keep only the ones that visibly matter.

## 4. Render and sample colors

```bash
pdftoppm -r 150 -f 1 -l 1 path/to/cv.pdf /tmp/cv_orig -png
```

Then sample with Python — never eyeball hex from the rendered preview, the display is not color-accurate:

```python
from PIL import Image
img = Image.open('/tmp/cv_orig-1.png')
w, h = img.size
# Walk a vertical column down the page to map a gradient or colored header
for y in range(0, 600, 30):
    px = img.getpixel((w//2, y))
    print(f'y={y}: #{px[0]:02X}{px[1]:02X}{px[2]:02X}')
```

To find text inside a colored band, scan a row for max contrast — the lightest pixel is usually the text on a dark background, the darkest is text on a light background:

```python
for y in [110, 120, 130, 140]:
    pixels = [img.getpixel((x, y))[:3] for x in range(int(w*0.55), int(w*0.95))]
    lightest = max(pixels, key=sum)
    darkest  = min(pixels, key=sum)
    print(f'y={y}: dark={darkest} light={lightest}')
```

Record:
- `Accent` — the brand color (top of gradient or icon color)
- `HeaderBgTop` / `HeaderBgBottom` — gradient endpoints, only if the source has a gradient
- `HeaderDark` — name and section-rule color, if the source uses a distinct header/rule color
- Text-on-header color, only if text sits on a colored/photo background

A non-zero "spread" in a row means text is present at that y; a zero spread usually means you're inside a flat color/gradient area with no glyphs.

## 5. Recolor icon assets if needed

Black PNG icons stay black on a dark header/gradient/photo area — invisible. If the icon background is dark, generate white versions:

```python
from PIL import Image
img = Image.open('assets/icons/email.png').convert('RGBA')
px = img.load()
for y in range(img.height):
    for x in range(img.width):
        r, g, b, a = px[x, y]
        px[x, y] = (255, 255, 255, a)   # keep alpha, force white
img.save('assets/icons-white/email.png')
```

`\textcolor{...}{\includegraphics{...}}` does **not** recolor a raster icon. Either swap the PNG or use a vector icon font (`fontawesome5` if installed).

## 6. Build the `.tex`

Match the source layout first. Use [`template_1.md`](./template_1.md) only when the CV actually uses a single-column layout with a large gradient/colored header, portrait, contact stack, and standard section/entry/bullet body. For a plain CV, two-column CV, no-photo CV, or no-gradient CV, reuse only the relevant macros and omit the visual modules that are not present in the source.

Optional design modules:

- **Full-bleed gradient/header, if present**: TikZ `[remember picture,overlay]` + `current page.north west/east` paints behind the text. Do **not** add a bottom border line when the source relies on the gradient itself as the visual edge.
- **Circular photo, if present**: `\clip (0,0) circle (R)` inside a TikZ scope, then draw the ring afterwards. The `path picture=` trick also works but `\clip` is more predictable when the source isn't perfectly square.
- **Escape non-ASCII characters in source** (`Jos\'e`, `Bogot\'a`, `Fran\c{c}ais`) — even with utf8 input, escaping survives encoding issues across systems.
- **Gradient height, if present**: ≈ 11 cm on A4 reproduces the typical "fade out behind the photo + contact block" look. Adjust to match where the source's gradient effectively reaches white.
- **Section rule color**: match the sampled source color. Use `HeaderDark` only when the source rules are dark; do not force teal/blue/accent rules unless the source uses them.

## 7. Compile and verify

```bash
pdflatex -interaction=nonstopmode CV_<Name>.tex   # twice if hyperref warns about labels
pdftoppm -r 130 -f 1 -l 1 CV_<Name>.pdf /tmp/check -png
```

Read `/tmp/check-1.png` and compare to `/tmp/cv_orig-1.png`. Specific things to check:

- If a photo exists, it is cropped to the same shape as the source; circular photos are round, not oval
- If a colored/gradient header exists, name/contact text sits on the correct part of the band
- Contact text is legible against the background at the y where it lands
- Section heading and rule colors match the sampled source colors
- Page count is the same as the source — if it overflows, tighten `\itemsep` or drop a redundant bullet, do not shrink the font as a first move

If the page count grows, that's usually a sign the source uses tighter `\setlength{\itemsep}`/`\topsep` than the defaults, or has merged short bullets.

## 8. Don't

- Don't invent contact info, dates, or bullets — if the source is ambiguous, ask.
- Don't translate the CV unless asked. Match source language, quotation style, and date conventions.
- Don't add a profile/tagline that isn't in the source.
- Don't reuse another CV's accent color "because it looks similar" — sample the source.
- Don't use `\amend` and don't commit unless the user asks.
