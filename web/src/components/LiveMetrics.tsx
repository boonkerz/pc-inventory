import { useEffect, useRef, useState } from "react";
import { api } from "../api";
import type { Command } from "../types";
import { useI18n } from "../i18n";

interface NetIO { name: string; bytes_sent: number; bytes_recv: number; }
interface Metrics {
  timestamp: number;
  cpu_percent: number;
  cpu_per_core?: number[];
  load_avg?: number[];
  mem_used_percent: number;
  mem_used: number;
  mem_total: number;
  swap_used_percent: number;
  disks?: { name: string; used_percent: number; used: number; total: number }[];
  net?: NetIO[];
}

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

function fmtBytes(n: number): string {
  if (!n) return "0 B";
  const u = ["B", "KB", "MB", "GB", "TB"];
  let i = 0, v = n;
  while (v >= 1024 && i < u.length - 1) { v /= 1024; i++; }
  return `${v.toFixed(v >= 10 || i === 0 ? 0 : 1)} ${u[i]}`;
}
const fmtRate = (bytesPerSec: number) => `${fmtBytes(bytesPerSec)}/s`;

function Bar({ label, percent, detail }: { label: string; percent: number; detail?: string }) {
  const p = Math.max(0, Math.min(100, percent));
  const cls = p > 90 ? "warn" : p > 75 ? "mid" : "";
  return (
    <div className="metric-row">
      <div className="metric-head"><span>{label}</span><span className="muted small">{detail ?? `${p.toFixed(0)}%`}</span></div>
      <div className="metric-track"><span className={`metric-fill ${cls}`} style={{ width: `${p}%` }} /></div>
    </div>
  );
}

// LiveMetrics fragt die Auslastung fortlaufend on-demand ab (CPU/RAM/Disk/Netzwerk).
export function LiveMetrics({ deviceId }: { deviceId: string }) {
  const { t } = useI18n();
  const [m, setM] = useState<Metrics | null>(null);
  const [netRate, setNetRate] = useState<{ up: number; down: number }>({ up: 0, down: 0 });
  const [error, setError] = useState("");
  const [running, setRunning] = useState(true);
  const prev = useRef<Metrics | null>(null);
  const alive = useRef(true);

  useEffect(() => {
    alive.current = true;
    const poll = async (id: string): Promise<Command | null> => {
      for (let i = 0; i < 40 && alive.current; i++) {
        await sleep(500);
        const cmd = await api.get<Command>(`/commands/${id}`);
        if (cmd.status === "done") return cmd;
      }
      return null;
    };
    const loop = async () => {
      while (alive.current && running) {
        try {
          const { command_id } = await api.post<{ command_id: string }>(`/devices/${deviceId}/metrics`);
          const cmd = await poll(command_id);
          if (!alive.current) return;
          if (!cmd) { setError(t("Zeitüberschreitung – Agent offline?")); await sleep(3000); continue; }
          const data = JSON.parse(cmd.output || "{}") as Metrics;
          setError("");
          // Netzwerk-Rate aus dem Delta der kumulativen Zähler.
          if (prev.current && data.timestamp > prev.current.timestamp) {
            const dt = (data.timestamp - prev.current.timestamp) / 1000;
            const sum = (list?: NetIO[]) => (list ?? []).reduce((a, n) => ({ s: a.s + n.bytes_sent, r: a.r + n.bytes_recv }), { s: 0, r: 0 });
            const now = sum(data.net), before = sum(prev.current.net);
            setNetRate({ up: Math.max(0, (now.s - before.s) / dt), down: Math.max(0, (now.r - before.r) / dt) });
          }
          prev.current = data;
          setM(data);
        } catch (e) { setError((e as Error).message); await sleep(2000); }
        await sleep(1500);
      }
    };
    void loop();
    return () => { alive.current = false; };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [deviceId, running]);

  return (
    <section className="card">
      <div className="page-head" style={{ marginBottom: 12 }}>
        <h2 style={{ margin: 0 }}>{t("Live-Auslastung")} {running && <span className="live-dot" />}</h2>
        <button className="btn ghost sm" onClick={() => setRunning((r) => !r)}>{running ? t("⏸ Pause") : t("▶ Fortsetzen")}</button>
      </div>

      {error && <p className="muted">{error}</p>}
      {!m ? <p className="muted">{t("Wird abgefragt… (Agent muss online sein)")}</p> : (
        <div className="grid-2">
          <div>
            <Bar label="CPU" percent={m.cpu_percent}
              detail={`${m.cpu_percent.toFixed(0)}%${m.load_avg && m.load_avg[0] ? ` · Load ${m.load_avg.map((x) => x.toFixed(2)).join(" / ")}` : ""}`} />
            {(m.cpu_per_core ?? []).length > 1 && (
              <div className="core-grid">
                {m.cpu_per_core!.map((c, i) => (
                  <div key={i} className="core-cell" title={`Kern ${i}: ${c.toFixed(0)}%`}>
                    <span className={`core-fill ${c > 90 ? "warn" : c > 75 ? "mid" : ""}`} style={{ height: `${Math.max(3, c)}%` }} />
                  </div>
                ))}
              </div>
            )}
            <Bar label={t("Arbeitsspeicher")} percent={m.mem_used_percent}
              detail={`${fmtBytes(m.mem_used)} / ${fmtBytes(m.mem_total)} (${m.mem_used_percent.toFixed(0)}%)`} />
            {m.swap_used_percent > 0 && <Bar label="Swap" percent={m.swap_used_percent} />}
          </div>

          <div>
            <div className="metric-row">
              <div className="metric-head"><span>{t("Netzwerk")}</span></div>
              <div className="net-rates">
                <span className="net-down">↓ {fmtRate(netRate.down)}</span>
                <span className="net-up">↑ {fmtRate(netRate.up)}</span>
              </div>
            </div>
            <div style={{ marginTop: 10 }}>
              <div className="metric-head"><span>{t("Datenträger")}</span></div>
              {(m.disks ?? []).map((d) => (
                <Bar key={d.name} label={d.name} percent={d.used_percent}
                  detail={`${fmtBytes(d.used)} / ${fmtBytes(d.total)} (${d.used_percent.toFixed(0)}%)`} />
              ))}
            </div>
          </div>
        </div>
      )}
    </section>
  );
}
