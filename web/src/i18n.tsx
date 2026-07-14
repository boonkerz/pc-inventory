import { createContext, useContext, useState, ReactNode, useCallback } from "react";
import { EN } from "./i18n_en";

export type Lang = "de" | "en";

// Der deutsche Text ist der Schlüssel; EN[key] liefert die Übersetzung, sonst
// bleibt der deutsche Text stehen (graceful fallback).
type Vars = Record<string, string | number>;

function translate(lang: Lang, key: string, vars?: Vars): string {
  let s = lang === "en" ? (EN[key] ?? key) : key;
  if (vars) {
    for (const [k, v] of Object.entries(vars)) {
      s = s.replace(new RegExp(`\\{${k}\\}`, "g"), String(v));
    }
  }
  return s;
}

// Modul-globale Sprache – für reine Funktionen (z.B. relTime), die keinen Hook nutzen.
let currentLang: Lang = (typeof localStorage !== "undefined" && (localStorage.getItem("roster-lang") as Lang)) || "de";
export function gt(key: string, vars?: Vars): string {
  return translate(currentLang, key, vars);
}

interface I18nCtx {
  lang: Lang;
  setLang: (l: Lang) => void;
  t: (key: string, vars?: Vars) => string;
}

const Ctx = createContext<I18nCtx>({ lang: "de", setLang: () => {}, t: (k) => k });

export function I18nProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>(() => (localStorage.getItem("roster-lang") as Lang) || "de");
  const setLang = useCallback((l: Lang) => {
    localStorage.setItem("roster-lang", l);
    currentLang = l;
    setLangState(l);
    document.documentElement.lang = l;
  }, []);
  const t = useCallback((key: string, vars?: Vars) => translate(lang, key, vars), [lang]);
  return <Ctx.Provider value={{ lang, setLang, t }}>{children}</Ctx.Provider>;
}

export function useI18n() {
  return useContext(Ctx);
}

// LangSwitch ist der Sprachumschalter für die Kopfzeile.
export function LangSwitch() {
  const { lang, setLang } = useI18n();
  return (
    <button className="lang-switch" onClick={() => setLang(lang === "de" ? "en" : "de")}
      title={lang === "de" ? "Switch to English" : "Auf Deutsch umschalten"}>
      {lang === "de" ? "EN" : "DE"}
    </button>
  );
}
