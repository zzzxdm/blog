export type ThemeMode = "light" | "dark";

export interface ThemeOption {
  label: string;
  value: string;
  accent: string;
  darkAccent: string;
  shell: string;
  darkShell: string;
  darkPalette: ThemePalette;
}

interface ThemePalette {
  bg: string;
  surface: string;
  surface2: string;
  ink: string;
  muted: string;
  line: string;
  code: string;
  shadow: string;
}

const lightPalette: ThemePalette = {
  bg: "#f7f5ef",
  surface: "#fffefa",
  surface2: "#f0eee6",
  ink: "#111714",
  muted: "#5f6f68",
  line: "#d9ded4",
  code: "#19231f",
  shadow: "0 18px 40px rgba(32, 39, 34, 0.08)"
};

export const themeOptions: ThemeOption[] = [
  {
    label: "松绿",
    value: "#295b4b",
    accent: "#163b32",
    darkAccent: "#9ec9ae",
    shell: "#17251f",
    darkShell: "#08120f",
    darkPalette: {
      bg: "#101614",
      surface: "#17211d",
      surface2: "#203028",
      ink: "#f7f3ea",
      muted: "#a9b8af",
      line: "#31433a",
      code: "#0b100e",
      shadow: "0 18px 44px rgba(0, 0, 0, 0.3)"
    }
  },
  {
    label: "赤陶",
    value: "#b95f2d",
    accent: "#7a3518",
    darkAccent: "#f0b089",
    shell: "#5a2d17",
    darkShell: "#34190d",
    darkPalette: {
      bg: "#170f0b",
      surface: "#211610",
      surface2: "#2d1d14",
      ink: "#fff4ed",
      muted: "#c9aa99",
      line: "#4b3023",
      code: "#0f0906",
      shadow: "0 18px 44px rgba(0, 0, 0, 0.34)"
    }
  },
  {
    label: "琥珀",
    value: "#e3b45d",
    accent: "#7a5412",
    darkAccent: "#f2d28a",
    shell: "#624719",
    darkShell: "#33250d",
    darkPalette: {
      bg: "#151107",
      surface: "#211a0e",
      surface2: "#2e2413",
      ink: "#fff6df",
      muted: "#d1bd91",
      line: "#4d3a17",
      code: "#0e0a04",
      shadow: "0 18px 44px rgba(0, 0, 0, 0.34)"
    }
  },
  {
    label: "灰色",
    value: "#64748b",
    accent: "#334155",
    darkAccent: "#cbd5e1",
    shell: "#273141",
    darkShell: "#151a22",
    darkPalette: {
      bg: "#0f1217",
      surface: "#171b23",
      surface2: "#212837",
      ink: "#f6f7f9",
      muted: "#b7c0cf",
      line: "#343d4c",
      code: "#090b0f",
      shadow: "0 18px 44px rgba(0, 0, 0, 0.32)"
    }
  }
];

const themePrimaryStorageKey = "site:theme-primary";
let currentPrimaryColor = "";

export function getInitialThemeMode(): ThemeMode {
  const stored = window.localStorage.getItem("site:theme");
  if (stored === "dark" || stored === "light") {
    return stored;
  }

  const prefersDark = typeof window.matchMedia === "function" && window.matchMedia("(prefers-color-scheme: dark)").matches;
  return prefersDark ? "dark" : "light";
}

export function applyThemeMode(mode: ThemeMode) {
  document.documentElement.dataset.theme = mode;
  window.localStorage.setItem("site:theme", mode);
  currentPrimaryColor ||= getStoredPrimaryColor();
  if (currentPrimaryColor) {
    setThemeVariables(currentPrimaryColor);
  }
}

export function applyPrimaryColor(color: string) {
  const primary = normalizeThemeColor(color);
  if (!primary) {
    return;
  }

  currentPrimaryColor = primary;
  window.localStorage.setItem(themePrimaryStorageKey, primary);
  setThemeVariables(primary);
}

export function normalizeThemeColor(color: string) {
  const value = color.trim().toLowerCase();
  return /^#[0-9a-f]{6}$/.test(value) ? value : "";
}

function setThemeVariables(primary: string) {
  const darkMode = document.documentElement.dataset.theme === "dark";
  const option = themeOptions.find((item) => item.value === primary);
  const palette = darkMode ? option?.darkPalette ?? createDarkPalette(primary) : lightPalette;
  const accent = option ? (darkMode ? option.darkAccent : option.accent) : themeAccentColor(primary);

  setPaletteVariables(palette);
  document.documentElement.style.setProperty("--green", primary);
  document.documentElement.style.setProperty("--green-2", accent);
  document.documentElement.style.setProperty("--theme-shell", option ? (darkMode ? option.darkShell : option.shell) : darkenHex(primary, darkMode ? 0.68 : 0.48));
  setElementPlusVariables(primary, accent);
}

function setPaletteVariables(palette: ThemePalette) {
  document.documentElement.style.setProperty("--bg", palette.bg);
  document.documentElement.style.setProperty("--surface", palette.surface);
  document.documentElement.style.setProperty("--surface-2", palette.surface2);
  document.documentElement.style.setProperty("--ink", palette.ink);
  document.documentElement.style.setProperty("--muted", palette.muted);
  document.documentElement.style.setProperty("--line", palette.line);
  document.documentElement.style.setProperty("--code", palette.code);
  document.documentElement.style.setProperty("--shadow", palette.shadow);
}

function setElementPlusVariables(primary: string, accent: string) {
  const style = document.documentElement.style;
  style.setProperty("--el-color-primary", primary);
  style.setProperty("--el-color-primary-dark-2", accent);
  style.setProperty("--el-color-primary-light-3", lightenHex(primary, 0.3));
  style.setProperty("--el-color-primary-light-5", lightenHex(primary, 0.5));
  style.setProperty("--el-color-primary-light-7", lightenHex(primary, 0.7));
  style.setProperty("--el-color-primary-light-8", lightenHex(primary, 0.82));
  style.setProperty("--el-color-primary-light-9", lightenHex(primary, 0.9));
  style.setProperty("--el-fill-color-light", "var(--surface-2)");
  style.setProperty("--el-fill-color-blank", "var(--surface)");
  style.setProperty("--el-border-color", "var(--line)");
  style.setProperty("--el-text-color-primary", "var(--ink)");
  style.setProperty("--el-text-color-regular", "var(--ink)");
  style.setProperty("--el-text-color-secondary", "var(--muted)");
}

function getStoredPrimaryColor() {
  return normalizeThemeColor(window.localStorage.getItem(themePrimaryStorageKey) || "");
}

function themeAccentColor(color: string) {
  const darkMode = document.documentElement.dataset.theme === "dark";
  const option = themeOptions.find((item) => item.value === color);
  if (option) {
    return darkMode ? option.darkAccent : option.accent;
  }

  return darkMode ? lightenHex(color, 0.46) : darkenHex(color, 0.38);
}

function darkenHex(color: string, amount: number) {
  const channels = [1, 3, 5].map((start) => {
    const value = Number.parseInt(color.slice(start, start + 2), 16);
    return Math.max(0, Math.round(value * (1 - amount)));
  });

  return `#${channels.map((value) => value.toString(16).padStart(2, "0")).join("")}`;
}

function lightenHex(color: string, amount: number) {
  const channels = [1, 3, 5].map((start) => {
    const value = Number.parseInt(color.slice(start, start + 2), 16);
    return Math.min(255, Math.round(value + (255 - value) * amount));
  });

  return `#${channels.map((value) => value.toString(16).padStart(2, "0")).join("")}`;
}

function createDarkPalette(primary: string): ThemePalette {
  return {
    bg: mixHex(primary, "#050505", 0.12),
    surface: mixHex(primary, "#0b0b0b", 0.18),
    surface2: mixHex(primary, "#111111", 0.25),
    ink: "#f7f3ea",
    muted: mixHex(primary, "#d7d7d7", 0.28),
    line: mixHex(primary, "#252525", 0.42),
    code: mixHex(primary, "#030303", 0.08),
    shadow: "0 18px 44px rgba(0, 0, 0, 0.32)"
  };
}

function mixHex(left: string, right: string, leftWeight: number) {
  const rightWeight = 1 - leftWeight;
  const channels = [1, 3, 5].map((start) => {
    const leftValue = Number.parseInt(left.slice(start, start + 2), 16);
    const rightValue = Number.parseInt(right.slice(start, start + 2), 16);
    return Math.round(leftValue * leftWeight + rightValue * rightWeight);
  });

  return `#${channels.map((value) => value.toString(16).padStart(2, "0")).join("")}`;
}
