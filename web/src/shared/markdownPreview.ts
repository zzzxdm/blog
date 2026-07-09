import { ref, watch } from "vue";

export interface MarkdownPreviewThemeOption {
  label: string;
  value: string;
}

const previewThemeStorageKey = "site:markdown-preview-theme";
const codeThemeStorageKey = "site:markdown-code-theme";

export const markdownPreviewThemes: MarkdownPreviewThemeOption[] = [
  { label: "GitHub", value: "github" },
  { label: "VuePress", value: "vuepress" },
  { label: "简洁蓝", value: "smart-blue" },
  { label: "青瓷", value: "cyanosis" },
  { label: "可爱", value: "mk-cute" },
  { label: "默认", value: "default" }
];

export const markdownCodeThemes: MarkdownPreviewThemeOption[] = [
  { label: "GitHub", value: "github" },
  { label: "Atom", value: "atom" },
  { label: "Stack Overflow", value: "stackoverflow" },
  { label: "Qt Creator", value: "qtcreator" },
  { label: "Kimbie", value: "kimbie" },
  { label: "Paraiso", value: "paraiso" }
];

export const markdownPreviewOptions = {
  theme: "light",
  previewTheme: "github",
  codeTheme: "github",
  language: "zh-CN",
  noKatex: true,
  noMermaid: true,
  noImgZoomIn: false
} as const;

const selectedPreviewTheme = ref<string>(markdownPreviewOptions.previewTheme);
const selectedCodeTheme = ref<string>(markdownPreviewOptions.codeTheme);
let themeStateReady = false;

export function loadMarkdownPreviewTheme() {
  return storedOption(previewThemeStorageKey, markdownPreviewThemes, markdownPreviewOptions.previewTheme);
}

export function loadMarkdownCodeTheme() {
  return storedOption(codeThemeStorageKey, markdownCodeThemes, markdownPreviewOptions.codeTheme);
}

export function saveMarkdownPreviewTheme(value: string) {
  saveStoredOption(previewThemeStorageKey, value, markdownPreviewThemes);
}

export function saveMarkdownCodeTheme(value: string) {
  saveStoredOption(codeThemeStorageKey, value, markdownCodeThemes);
}

export function useMarkdownPreviewTheme() {
  if (!themeStateReady) {
    selectedPreviewTheme.value = loadMarkdownPreviewTheme();
    selectedCodeTheme.value = loadMarkdownCodeTheme();

    watch(selectedPreviewTheme, saveMarkdownPreviewTheme);
    watch(selectedCodeTheme, saveMarkdownCodeTheme);
    themeStateReady = true;
  }

  return {
    selectedPreviewTheme,
    selectedCodeTheme
  };
}

function storedOption(key: string, options: MarkdownPreviewThemeOption[], fallback: string) {
  if (typeof window === "undefined") {
    return fallback;
  }

  const stored = window.localStorage.getItem(key) || "";
  return options.some((option) => option.value === stored) ? stored : fallback;
}

function saveStoredOption(key: string, value: string, options: MarkdownPreviewThemeOption[]) {
  if (typeof window === "undefined" || !options.some((option) => option.value === value)) {
    return;
  }

  window.localStorage.setItem(key, value);
}
