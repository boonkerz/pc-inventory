import { useEffect, useRef, useState } from "react";
import { useI18n } from "../i18n";
import { Terminal as XTerm } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";

// DeviceTerminal öffnet ein interaktives Live-Terminal zum Agent über eine WebSocket.
// Rohe Terminal-I/O läuft als Binär-Frames, Steuerung (resize/exit) als Text-Frames.
// fill = das Terminal füllt die verfügbare Höhe (Popout-Fenster); autoStart =
// sofort verbinden (ohne Klick auf „Verbinden"), für das Popout-Fenster.
export function DeviceTerminal({ id, os, fill, autoStart, initialShell, initialRunas }: {
  id: string; os: string; fill?: boolean; autoStart?: boolean;
  initialShell?: string; initialRunas?: "system" | "user";
}) {
  const { t } = useI18n();
  const isWindows = /win/i.test(os);
  const [shell, setShell] = useState(initialShell || (isWindows ? "cmd" : "shell"));
  const [runas, setRunas] = useState<"system" | "user">(initialRunas || "system");
  const [status, setStatus] = useState<string>("");
  const [session, setSession] = useState(autoStart ? 1 : 0); // hochzählen = (neu) verbinden
  const hostRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

  // Popout: aktuelles Terminal in eigenem Fenster öffnen (gleiche Origin →
  // Cookie-Auth + WebSocket funktionieren). shell/runas werden mitgegeben.
  const popout = () => {
    const q = new URLSearchParams({ os, shell, runas });
    window.open(`/devices/${id}/terminal?${q.toString()}`, `term-${id}`,
      "width=900,height=560,menubar=no,toolbar=no,location=no,status=no");
  };

  useEffect(() => {
    if (session === 0 || !hostRef.current) return;
    const term = new XTerm({
      fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
      fontSize: 13,
      cursorBlink: true,
      theme: { background: "#0b0e14", foreground: "#d7dce5" },
    });
    const fit = new FitAddon();
    term.loadAddon(fit);
    term.open(hostRef.current);

    const proto = location.protocol === "https:" ? "wss" : "ws";
    const url = `${proto}://${location.host}/api/v1/devices/${id}/terminal?shell=${encodeURIComponent(shell)}&runas=${runas}`;
    const ws = new WebSocket(url);
    wsRef.current = ws;
    ws.binaryType = "arraybuffer";
    setStatus(t("verbinde…"));

    const sendResize = () => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: "resize", cols: term.cols, rows: term.rows }));
      }
    };
    // Neu vermessen + Server informieren. Mehrfach nötig, weil die korrekte
    // Zellenhöhe erst nach dem Laden der Monospace-Schrift feststeht – sonst wird
    // die letzte Zeile abgeschnitten.
    const refit = () => { try { fit.fit(); } catch { /* leeres Layout */ } sendResize(); };
    requestAnimationFrame(refit);
    if (document.fonts?.ready) document.fonts.ready.then(refit);
    ws.onopen = () => { setStatus(t("verbunden")); refit(); term.focus(); };
    ws.onmessage = (ev) => {
      if (typeof ev.data === "string") {
        try {
          const m = JSON.parse(ev.data);
          if (m.type === "exit") {
            term.write(`\r\n\x1b[90m[Sitzung beendet – Exit-Code ${m.code}]\x1b[0m\r\n`);
            setStatus(t("beendet ({code})", { code: m.code }));
          }
        } catch { /* ignorieren */ }
      } else {
        term.write(new Uint8Array(ev.data));
      }
    };
    ws.onclose = () => setStatus((s) => (s.startsWith(t("beendet")) ? s : t("getrennt")));
    ws.onerror = () => setStatus(t("Verbindungsfehler"));

    const dataSub = term.onData((d) => {
      if (ws.readyState === WebSocket.OPEN) ws.send(new TextEncoder().encode(d));
    });
    const ro = new ResizeObserver(refit);
    ro.observe(hostRef.current);

    return () => {
      ro.disconnect();
      dataSub.dispose();
      ws.close();
      term.dispose();
      if (wsRef.current === ws) wsRef.current = null;
    };
    // shell/runas werden beim Verbinden gelesen; ein Wechsel erfordert „Neu verbinden".
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [session, id]);

  const connected = status === t("verbunden") || status === t("verbinde…");
  const disconnect = () => wsRef.current?.close();

  return (
    <section className={fill ? "term-fill" : "card"}>
      <div className="inline-form" style={{ marginBottom: 10 }}>
        <select value={shell} onChange={(e) => setShell(e.target.value)} disabled={connected}>
          {isWindows
            ? <>
              <option value="cmd">{t("CMD")}</option>
              <option value="powershell">{t("PowerShell")}</option>
            </>
            : <option value="shell">{t("Shell")}</option>}
        </select>
        <select value={runas} onChange={(e) => setRunas(e.target.value as "system" | "user")} disabled={connected}>
          <option value="system">{isWindows ? t("als SYSTEM (Dienst)") : t("als root (Dienst)")}</option>
          <option value="user">{t("als angemeldeter Benutzer")}</option>
        </select>
        {connected ? (
          <button className="btn" onClick={disconnect}>{t("Trennen")}</button>
        ) : (
          <button className="btn primary" onClick={() => setSession((n) => n + 1)}>
            {status ? t("Neu verbinden") : t("Verbinden")}
          </button>
        )}
        {!fill && (
          <button className="btn ghost" onClick={popout} title={t("In eigenem Fenster öffnen")}>Popout ⇗</button>
        )}
        {status && <span className="muted small">{t("Status")}: {status}</span>}
      </div>
      <div
        ref={hostRef}
        className={fill ? "term-host term-host-fill" : "term-host"}
        style={{ background: "#0b0e14", borderRadius: 6, paddingLeft: 6, overflow: "hidden", ...(fill ? {} : { height: 440 }) }}
      />
      {!fill && (
        <p className="muted small" style={{ marginTop: 8 }}>
          {t("Läuft mit den Rechten des Agent-Dienstes bzw. des gewählten Benutzers – wie eine lokale Konsole.")}
        </p>
      )}
    </section>
  );
}
