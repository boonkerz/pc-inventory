import { FormEvent, useEffect, useState } from "react";
import QRCode from "qrcode";
import { useMutation, useQuery } from "@tanstack/react-query";
import { api, ApiError } from "../api";
import { useI18n } from "../i18n";
import { useAuth } from "../auth";
import { ThemeToggle } from "../theme";

// TwoFactorSetup führt durch die TOTP-Einrichtung (Pflicht-Gate vor der App-Nutzung).
export function TwoFactorSetup() {
  const { t } = useI18n();
  const { logout, reload } = useAuth();
  const { data: setup } = useQuery({
    queryKey: ["2fa-setup"],
    queryFn: () => api.post<{ secret: string; otpauth_url: string }>("/auth/2fa/setup"),
    staleTime: Infinity, retry: false,
  });
  const [qr, setQr] = useState("");
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [codes, setCodes] = useState<string[] | null>(null);

  useEffect(() => {
    if (setup?.otpauth_url) QRCode.toDataURL(setup.otpauth_url, { margin: 1, width: 200 }).then(setQr);
  }, [setup]);

  const enable = useMutation({
    mutationFn: () => api.post<{ recovery_codes: string[] }>("/auth/2fa/enable", { code: code.trim() }),
    onSuccess: (d) => setCodes(d.recovery_codes),
    onError: (e) => setError(e instanceof ApiError ? e.message : t("Code ungültig")),
  });

  const submit = (e: FormEvent) => { e.preventDefault(); setError(""); enable.mutate(); };

  return (
    <div className="login-wrap">
      <div className="login-card" style={{ maxWidth: 440 }}>
        <ThemeToggle />
        <div className="brand login-brand"><span className="brand-mark">▣</span> PC-Inventar</div>

        {codes ? (
          <>
            <h2>{t("Wiederherstellungscodes")}</h2>
            <p className="muted small">{t("Bewahre diese Einmal-Codes sicher auf – sie sind dein Zugang, falls du dein Gerät verlierst. Sie werden nur jetzt angezeigt.")}</p>
            <pre className="help-code" style={{ columns: 2 }}>{codes.join("\n")}</pre>
            <button className="btn primary" onClick={() => reload()}>{t("Ich habe sie gespeichert – weiter")}</button>
          </>
        ) : (
          <>
            <h2>{t("Zwei-Faktor einrichten")}</h2>
            <p className="muted small">{t("Zwei-Faktor-Authentifizierung ist Pflicht. Scanne den QR-Code mit einer Authenticator-App (z.B. Google Authenticator, Aegis, 1Password) und gib den 6-stelligen Code ein.")}</p>
            {qr && <img src={qr} alt="QR-Code" style={{ alignSelf: "center", borderRadius: 6 }} />}
            {setup && <p className="muted small">{t("Manuell:")} <code>{setup.secret}</code></p>}
            <form onSubmit={submit}>
              <label>{t("Code")}<input value={code} onChange={(e) => setCode(e.target.value)} inputMode="numeric" placeholder="123456" autoFocus /></label>
              {error && <div className="form-error">{error}</div>}
              <button className="btn primary" type="submit" disabled={enable.isPending}>{t("Aktivieren")}</button>
            </form>
          </>
        )}
        <button type="button" className="btn ghost" onClick={() => logout()}>{t("Abmelden")}</button>
      </div>
    </div>
  );
}
