import { FormEvent, useState } from "react";
import { useAuth } from "../auth";
import { ApiError } from "../api";
import { ThemeToggle } from "../theme";
import { LangSwitch, useI18n } from "../i18n";

export function Login() {
  const { login, completeTotp } = useAuth();
  const { t } = useI18n();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [pending, setPending] = useState<string | null>(null); // gesetzt = 2. Faktor nötig
  const [code, setCode] = useState("");

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      const res = await login(username, password);
      if ("totp" in res) setPending(res.pending);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("Anmeldung fehlgeschlagen"));
    } finally {
      setBusy(false);
    }
  };

  const submitCode = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      await completeTotp(pending!, code.trim());
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("Code ungültig"));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="login-wrap">
      {pending ? (
        <form className="login-card" onSubmit={submitCode}>
          <div className="login-top"><LangSwitch /><ThemeToggle /></div>
          <div className="brand login-brand"><span className="brand-mark">▣</span> PC-Inventar</div>
          <p className="muted small">{t("Bestätigungscode aus deiner Authenticator-App (oder ein Backup-Code).")}</p>
          <label>
            {t("Code")}
            <input value={code} onChange={(e) => setCode(e.target.value)} autoFocus inputMode="numeric"
              autoComplete="one-time-code" placeholder="123456" />
          </label>
          {error && <div className="form-error">{error}</div>}
          <button className="btn primary" disabled={busy} type="submit">{busy ? t("Prüfe…") : t("Bestätigen")}</button>
          <button type="button" className="btn ghost" onClick={() => { setPending(null); setCode(""); setError(""); }}>{t("Zurück")}</button>
        </form>
      ) : (
        <form className="login-card" onSubmit={submit}>
          <div className="login-top"><LangSwitch /><ThemeToggle /></div>
          <div className="brand login-brand"><span className="brand-mark">▣</span> PC-Inventar</div>
          <label>
            {t("Benutzername")}
            <input value={username} onChange={(e) => setUsername(e.target.value)} autoComplete="username" autoFocus />
          </label>
          <label>
            {t("Passwort")}
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} autoComplete="current-password" />
          </label>
          {error && <div className="form-error">{error}</div>}
          <button className="btn primary" disabled={busy} type="submit">{busy ? t("Anmelden…") : t("Anmelden")}</button>
        </form>
      )}
    </div>
  );
}
