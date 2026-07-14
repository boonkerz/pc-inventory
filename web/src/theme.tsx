import { createContext, useContext, useEffect, useState, ReactNode } from "react";
import { api } from "./api";
import { useAuth } from "./auth";

export type Theme = "light" | "dark";
const STORAGE_KEY = "roster-theme";

// resolveInitialTheme liest die lokal gecachte Wahl oder fällt auf die Systemeinstellung
// zurück. Wird vor dem Login und für den ersten Render genutzt.
export function resolveInitialTheme(): Theme {
  const saved = localStorage.getItem(STORAGE_KEY);
  if (saved === "light" || saved === "dark") return saved;
  return window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark";
}

export function applyTheme(theme: Theme) {
  document.documentElement.dataset.theme = theme;
}

interface ThemeState {
  theme: Theme;
  toggle: () => void;
}

const ThemeContext = createContext<ThemeState>(null!);

export function ThemeProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth();
  const [theme, setThemeState] = useState<Theme>(resolveInitialTheme);

  // Sobald der angemeldete Benutzer geladen ist, sein serverseitig gespeichertes
  // Theme übernehmen (folgt dem Benutzer geräteübergreifend).
  useEffect(() => {
    if (user?.theme === "light" || user?.theme === "dark") {
      setThemeState(user.theme);
    }
  }, [user]);

  // Theme auf das Dokument anwenden und lokal cachen (für Login-Screen / nächsten Start).
  useEffect(() => {
    applyTheme(theme);
    localStorage.setItem(STORAGE_KEY, theme);
  }, [theme]);

  const setTheme = (t: Theme) => {
    setThemeState(t);
    // Bei angemeldetem Benutzer die Präferenz serverseitig speichern.
    if (user) {
      api.put("/auth/me/theme", { theme: t }).catch(() => {});
    }
  };

  const toggle = () => setTheme(theme === "dark" ? "light" : "dark");

  return <ThemeContext.Provider value={{ theme, toggle }}>{children}</ThemeContext.Provider>;
}

export const useTheme = () => useContext(ThemeContext);

// ThemeToggle ist der Umschalt-Button; zeigt das Symbol des Zielthemes.
export function ThemeToggle() {
  const { theme, toggle } = useTheme();
  const toLight = theme === "dark";
  return (
    <button
      className="theme-toggle"
      onClick={toggle}
      title={toLight ? "Zu hellem Design wechseln" : "Zu dunklem Design wechseln"}
      aria-label="Design umschalten"
    >
      {toLight ? "☀" : "☾"}
    </button>
  );
}
