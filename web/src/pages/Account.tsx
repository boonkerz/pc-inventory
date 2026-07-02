import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { api } from "../api";
import { useAuth } from "../auth";
import { useI18n } from "../i18n";

// Account ist der Selbstverwaltungs-Bereich: Kontoinfos + Passwort ändern.
export function Account() {
  const { user } = useAuth();
  const { t } = useI18n();
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);

  const isLocal = user?.auth_source === "local";

  const change = useMutation({
    mutationFn: () => api.post("/auth/me/password", { current_password: current, new_password: next }),
    onSuccess: () => {
      setMsg({ ok: true, text: t("Passwort geändert.") });
      setCurrent(""); setNext(""); setConfirm("");
    },
    onError: (e: Error) => setMsg({ ok: false, text: e.message }),
  });

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    setMsg(null);
    if (next.length < 8) { setMsg({ ok: false, text: t("Das neue Passwort muss mindestens 8 Zeichen haben.") }); return; }
    if (next !== confirm) { setMsg({ ok: false, text: t("Die Passwörter stimmen nicht überein.") }); return; }
    change.mutate();
  };

  return (
    <div className="account">
      <h1 style={{ marginTop: 0 }}>{t("Mein Konto")}</h1>

      <section className="card">
        <h2>{t("Kontodaten")}</h2>
        <dl className="kv">
          <dt>{t("Benutzername")}</dt><dd>{user?.username}</dd>
          <dt>{t("Rolle")}</dt><dd>{user?.role === "admin" ? t("Administrator") : user?.role === "technician" ? t("Techniker") : t("Viewer (nur lesen)")}</dd>
          <dt>{t("Anmeldequelle")}</dt><dd>{isLocal ? t("Lokal") : (user?.auth_source ?? "—").toUpperCase()}</dd>
          <dt>{t("Zwei-Faktor")}</dt><dd>{user?.totp_enabled ? t("aktiv") : t("nicht aktiv")}</dd>
        </dl>
      </section>

      <section className="card">
        <h2>{t("Passwort ändern")}</h2>
        {!isLocal ? (
          <p className="muted">{t("Dein Konto wird extern authentifiziert – das Passwort kann hier nicht geändert werden.")}</p>
        ) : (
          <form className="form-col" onSubmit={submit} style={{ maxWidth: 360 }}>
            <label className="field">
              <span>{t("Aktuelles Passwort")}</span>
              <input type="password" autoComplete="current-password" value={current} onChange={(e) => setCurrent(e.target.value)} required />
            </label>
            <label className="field">
              <span>{t("Neues Passwort")}</span>
              <input type="password" autoComplete="new-password" value={next} onChange={(e) => setNext(e.target.value)} required />
            </label>
            <label className="field">
              <span>{t("Neues Passwort wiederholen")}</span>
              <input type="password" autoComplete="new-password" value={confirm} onChange={(e) => setConfirm(e.target.value)} required />
            </label>
            {msg && <p className={msg.ok ? "form-ok" : "form-err"}>{msg.text}</p>}
            <div>
              <button className="btn primary" type="submit" disabled={change.isPending}>
                {change.isPending ? t("Speichert…") : t("Passwort ändern")}
              </button>
            </div>
          </form>
        )}
      </section>
    </div>
  );
}
