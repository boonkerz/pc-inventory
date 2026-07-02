import { useEffect, useState } from "react";
import { api } from "../api";
import type { Command } from "../types";
import { useI18n, gt } from "../i18n";

interface ServiceInfo { name: string; display?: string; running: boolean; start_type?: string; }
interface ProcInfo { pid: number; name: string; user?: string; mem_bytes: number; }

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

function fmtBytes(n: number): string {
  if (!n) return "0";
  const u = ["B", "KB", "MB", "GB", "TB"];
  let i = 0, v = n;
  while (v >= 1024 && i < u.length - 1) { v /= 1024; i++; }
  return `${v.toFixed(v >= 10 || i === 0 ? 0 : 1)} ${u[i]}`;
}

// runCmd stößt einen Befehl an und pollt dessen Ergebnis (Befehl läuft on-demand).
async function runCmd(path: string, body?: unknown): Promise<{ output: string; exit: number }> {
  const { command_id } = await api.post<{ command_id: string }>(path, body);
  for (let i = 0; i < 130; i++) {
    await sleep(700);
    const cmd = await api.get<Command>(`/commands/${command_id}`);
    if (cmd.status === "done") return { output: cmd.output ?? "", exit: cmd.exit_code };
  }
  throw new Error(gt("Zeitüberschreitung – Agent offline?"));
}

// ServicesProcesses zeigt on-demand Dienste bzw. Prozesse eines Geräts mit Steuerung.
export function ServicesProcesses({ deviceId, isAdmin }: { deviceId: string; isAdmin: boolean }) {
  const { t } = useI18n();
  const [view, setView] = useState<"services" | "processes">("services");
  const [services, setServices] = useState<ServiceInfo[] | null>(null);
  const [procs, setProcs] = useState<ProcInfo[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");
  const [filter, setFilter] = useState("");

  const loadServices = async () => {
    setLoading(true); setError("");
    try {
      const { output } = await runCmd(`/devices/${deviceId}/services`);
      const data = JSON.parse(output || "{}");
      if (data.error) setError(data.error); else setServices(data.services ?? []);
    } catch (e) { setError((e as Error).message); } finally { setLoading(false); }
  };
  const loadProcs = async () => {
    setLoading(true); setError("");
    try {
      const { output } = await runCmd(`/devices/${deviceId}/processes`);
      const data = JSON.parse(output || "{}");
      if (data.error) setError(data.error); else setProcs(data.processes ?? []);
    } catch (e) { setError((e as Error).message); } finally { setLoading(false); }
  };

  // Beim Aktivieren einer Ansicht einmalig laden.
  useEffect(() => {
    if (view === "services" && services === null) void loadServices();
    if (view === "processes" && procs === null) void loadProcs();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [view]);

  const control = async (name: string, action: "start" | "stop" | "restart") => {
    setBusy(name + action);
    try {
      await runCmd(`/devices/${deviceId}/service-control`, { name, action });
      await loadServices();
    } catch (e) { setError((e as Error).message); } finally { setBusy(""); }
  };
  const kill = async (pid: number) => {
    if (!confirm(t("Prozess {pid} beenden?", { pid }))) return;
    setBusy("kill" + pid);
    try {
      await runCmd(`/devices/${deviceId}/process-kill`, { pid });
      await loadProcs();
    } catch (e) { setError((e as Error).message); } finally { setBusy(""); }
  };

  const f = filter.toLowerCase();
  const shownServices = (services ?? []).filter((s) => s.name.toLowerCase().includes(f) || (s.display ?? "").toLowerCase().includes(f));
  const shownProcs = (procs ?? []).filter((p) => p.name.toLowerCase().includes(f) || String(p.pid).includes(f));

  return (
    <section className="card">
      <div className="page-head" style={{ marginBottom: 10 }}>
        <div className="inline-form">
          <div className="tabs" style={{ margin: 0 }}>
            <button className={`tab ${view === "services" ? "tab-on" : ""}`} onClick={() => setView("services")}>{t("Dienste")}</button>
            <button className={`tab ${view === "processes" ? "tab-on" : ""}`} onClick={() => setView("processes")}>{t("Prozesse")}</button>
          </div>
        </div>
        <div className="inline-form">
          <input className="search" placeholder={t("Filtern…")} value={filter} onChange={(e) => setFilter(e.target.value)} />
          <button className="btn ghost sm" disabled={loading} onClick={() => view === "services" ? loadServices() : loadProcs()}>↻ {t("Aktualisieren")}</button>
        </div>
      </div>

      {loading ? <p className="muted">{t("Wird abgefragt… (Agent muss online sein)")}</p>
        : error ? <p className="muted">{t("Fehler")}: {error}</p>
        : view === "services" ? (
          <div className="scroll-list">
            <table className="table">
              <thead><tr><th>{t("Status")}</th><th>{t("Dienst")}</th><th>{t("Beschreibung")}</th>{isAdmin && <th>{t("Aktion")}</th>}</tr></thead>
              <tbody>
                {shownServices.map((s) => (
                  <tr key={s.name}>
                    <td><span className={`badge ${s.running ? "badge-online" : "badge-unknown"}`}><span className="dot" /> {s.running ? t("läuft") : t("gestoppt")}</span></td>
                    <td className="mono small">{s.name}</td>
                    <td className="muted small">{s.display || "—"}</td>
                    {isAdmin && (
                      <td className="inline-form">
                        {s.running
                          ? <>
                              <button className="btn ghost sm" disabled={!!busy} onClick={() => control(s.name, "restart")}>{t("Neustart")}</button>
                              <button className="btn ghost sm" disabled={!!busy} onClick={() => control(s.name, "stop")}>{t("Stop")}</button>
                            </>
                          : <button className="btn ghost sm" disabled={!!busy} onClick={() => control(s.name, "start")}>{t("Start")}</button>}
                      </td>
                    )}
                  </tr>
                ))}
                {shownServices.length === 0 && <tr><td colSpan={isAdmin ? 4 : 3} className="empty">{t("Keine Dienste.")}</td></tr>}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="scroll-list">
            <table className="table">
              <thead><tr><th>PID</th><th>{t("Prozess")}</th><th>{t("Benutzer")}</th><th>{t("Arbeitsspeicher")}</th>{isAdmin && <th></th>}</tr></thead>
              <tbody>
                {shownProcs.map((p) => (
                  <tr key={p.pid}>
                    <td className="mono small">{p.pid}</td>
                    <td>{p.name}</td>
                    <td className="muted small">{p.user || "—"}</td>
                    <td className="muted mono small">{fmtBytes(p.mem_bytes)}</td>
                    {isAdmin && <td><button className="btn ghost sm" disabled={!!busy} onClick={() => kill(p.pid)}>{t("Beenden")}</button></td>}
                  </tr>
                ))}
                {shownProcs.length === 0 && <tr><td colSpan={isAdmin ? 5 : 4} className="empty">{t("Keine Prozesse.")}</td></tr>}
              </tbody>
            </table>
          </div>
        )}
    </section>
  );
}
