# Template 1 — Single-column CV with gradient header + circular photo

A reusable layout for CVs with a full-bleed gradient at the top, circular profile photo with white ring on the left, name + contact stack on the right, then standard section/entry/bullet body.

Replace placeholders marked `<…>`. Sampled colors come from step 4 of `SKILL.md`.

```latex
\documentclass[11pt,a4paper]{article}

\usepackage[a4paper,margin=1.35cm]{geometry}
\usepackage[T1]{fontenc}
\usepackage[utf8]{inputenc}
\usepackage[english]{babel}              % match source language
\usepackage{graphicx}
\usepackage{lmodern}
\usepackage{xcolor}
\usepackage{tabularx}
\usepackage{hyperref}
\usepackage{tikz}
\usetikzlibrary{shadings}

\pagestyle{empty}
\setlength{\parindent}{0pt}
\setlength{\parskip}{0pt}
\emergencystretch=2em

% --- Sampled colors (replace with values from step 4 of SKILL.md) ---
\definecolor{Accent}{HTML}{<sampled-accent>}        % e.g. 28696F
\definecolor{HeaderDark}{HTML}{<sampled-dark>}      % e.g. 151723
\definecolor{HeaderBgTop}{HTML}{<sampled-bg-top>}   % e.g. 28696F
\definecolor{HeaderBgBottom}{HTML}{FFFFFF}

\hypersetup{colorlinks=true, urlcolor=Accent, linkcolor=Accent}

% --- Icons (use white versions when the header is dark — see SKILL.md step 5) ---
\newcommand{\contacticon}[2][1.05em]{%
  \raisebox{-0.22ex}{\includegraphics[height=#1]{#2}}%
}
\newcommand{\emailicon}{\contacticon{assets/icons-white/email.png}}
\newcommand{\websiteicon}{\contacticon{assets/icons-white/website.png}}
\newcommand{\linkedinicon}{\contacticon{assets/icons-white/InBug-Black.png}}
\newcommand{\pinicon}{\textcolor{white}{\textbullet}}
\newcommand{\phoneicon}{\textcolor{white}{\textbullet}}

% --- Candidate fields ---
\newcommand{\candidateName}{<NAME IN CAPS>}
\newcommand{\candidateAddress}{<address>}
\newcommand{\candidatePhone}{<phone>}
\newcommand{\candidateEmail}{<email@example.com>}
\newcommand{\candidateLinkedIn}{<https://linkedin.com/in/...>}
\newcommand{\candidateWebsite}{<https://example.com>}

% --- Section / entry / bullet macros ---
\newcommand{\cvsection}[1]{%
  \vspace{0.75em}%
  {\large\bfseries\color{HeaderDark} #1}\par
  \vspace{0.2em}%
  {\color{HeaderDark}\hrule height 0.8pt}%
  \vspace{0.35em}%
}

\newcommand{\cventry}[3]{%
  \vspace{0.25em}%
  \begin{tabularx}{\textwidth}{@{}Xr@{}}
    \textbf{#1} & \textbf{#3} \\
    \multicolumn{2}{@{}l@{}}{\textit{#2}} \\
  \end{tabularx}%
}

\newenvironment{cvitems}{%
  \begin{list}{\textbullet}{%
    \setlength{\leftmargin}{1.2em}%
    \setlength{\labelwidth}{0.6em}%
    \setlength{\labelsep}{0.6em}%
    \setlength{\itemsep}{0.2em}%
    \setlength{\topsep}{0.25em}%
    \setlength{\parsep}{0pt}%
    \setlength{\partopsep}{0pt}%
    \setlength{\itemindent}{0pt}%
    \setlength{\listparindent}{0pt}%
  }
}{%
  \end{list}
}

\begin{document}

% Full-bleed gradient — no bottom border, the fade IS the edge
\begin{tikzpicture}[remember picture,overlay]
  \shade[top color=HeaderBgTop, bottom color=HeaderBgBottom]
    (current page.north west) rectangle
    ([yshift=-11cm]current page.north east);
\end{tikzpicture}

\vspace*{0.6em}
\noindent
\begin{minipage}[c]{0.22\textwidth}
  % Circular crop via \clip — works even when the source photo isn't perfectly square
  \begin{tikzpicture}
    \begin{scope}
      \clip (0,0) circle (1.7cm);
      \node at (0,0) {\includegraphics[width=3.6cm]{assets/<name>/photo.png}};
    \end{scope}
    \draw[white, line width=1.5pt] (0,0) circle (1.7cm);
  \end{tikzpicture}
\end{minipage}\hfill
\begin{minipage}[c]{0.74\textwidth}
  \raggedright
  {\fontsize{26}{30}\selectfont\bfseries\color{white} \candidateName}\par
  \vspace{0.6em}
  {\small\color{white}
    \pinicon\hspace{0.4em}\candidateAddress\par
    \vspace{0.25em}
    \phoneicon\hspace{0.4em}\candidatePhone\par
    \vspace{0.25em}
    \href{mailto:\candidateEmail}{\emailicon\hspace{0.4em}\candidateEmail}\par
    \vspace{0.25em}
    \href{\candidateLinkedIn}{\linkedinicon\hspace{0.4em}<linkedin label>}\par
    \vspace{0.25em}
    \href{\candidateWebsite}{\websiteicon\hspace{0.4em}<website label>}\par
  }
\end{minipage}

\vspace{1.2em}

% --- Body sections (repeat \cvsection / \cventry / cvitems for each block) ---

\cvsection{<First section, e.g. Education>}

\cventry{<Institution>}{<Program / Role>}{<Dates>}
\begin{cvitems}
  \item <bullet>
\end{cvitems}

\cvsection{<Next section, e.g. Work Experience>}

\cventry{<Employer>}{<Role>}{<Dates>}
\begin{cvitems}
  \item <bullet>
\end{cvitems}

\cvsection{Skills}
\begin{cvitems}
  \item <skill>
\end{cvitems}

\cvsection{<Languages section, e.g. Languages>}
\begin{cvitems}
  \item <language list>
\end{cvitems}

\cvsection{<Hobbies section, e.g. Hobbies>}
<free text>

\end{document}
```

## Knobs to tune

- **Gradient height** — change `-11cm` to where the source's fade visually reaches white.
- **Photo radius** — `circle (1.7cm)` + `\includegraphics[width=3.6cm]` keeps the photo slightly larger than the clip so the edge stays clean. Scale both proportionally.
- **Name size** — `\fontsize{26}{30}` for ~A4 + 1.35 cm margins. Drop to 22/26 for a longer name.
- **Body density** — if the page count grows past the source, tighten `\itemsep` to `0.15em` and `\topsep` to `0.2em` before touching the base font.
- **Language** — swap `[english]{babel}` for `spanish`, `french`, etc. Keep escaped diacritics (`Jos\'e`, `Bogot\'a`, `Fran\c{c}ais`) for portability.
