import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "../api";
import { useI18n } from "../i18n";

interface Point { ts: number; cpu: number; mem: number; disk: number; }

const W = 600, H = 150, PAD = 4;

// MetricsHistory zeigt CPU/RAM/Disk als Verlaufschart (24h/7d/30d).
export function MetricsHistory({ deviceId }: { deviceId: string }) {
  const { t } = useI18n();
  const [range, setRange] = useState<"24h" | "7d" | "30d">("24h");
  const { data } = useQuery({
    queryKey: ["metrics-history", deviceId, range],
    queryFn: () => api.get<Point[]>(`/devices/${deviceId}/metrics-history?range=${range}`),
    refetchInterval: 60000,
  });
  const points = data ?? [];

  const path = (key: "cpu" | "mem" | "disk") => {
    if (points.length < 2) return "";
    return points.map((p, i) => {
      const x = PAD + (i / (points.length - 1)) * (W - 2 * PAD);
      const y = PAD + (1 - Math.min(100, p[key]) / 100) * (H - 2 * PAD);
      return `${i === 0 ? "M" : "L"}${x.toFixed(1)},${y.toFixed(1)}`;
    }).join(" ");
  };

  const fmt = (ms: number) => new Date(ms).toLocaleString([], range === "24h"
    ? { hour: "2-digit", minute: "2-digit" }
    : { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" });
  const yPct = (v: number) => ((PAD + (1 - v / 100) * (H - 2 * PAD)) / H) * 100;
  const last = points[points.length - 1];
  const val = (v?: number) => (v == null ? "–" : `${Math.round(v)} %`);

  return (
    <div className="metrics-history">
      <div className="inline-form" style={{ justifyContent: "space-between", alignItems: "center" }}>
        <div className="chart-legend">
          <span className="lg cpu">CPU {val(last?.cpu)}</span>
          <span className="lg mem">RAM {val(last?.mem)}</span>
          <span className="lg disk">Disk {val(last?.disk)}</span>
        </div>
        <select value={range} onChange={(e) => setRange(e.target.value as "24h" | "7d" | "30d")}>
          <option value="24h">{t("24 Stunden")}</option>
          <option value="7d">{t("7 Tage")}</option>
          <option value="30d">{t("30 Tage")}</option>
        </select>
      </div>
      {points.length < 2 ? (
        <p className="muted small">{t("Noch keine Verlaufsdaten (werden je Checkin gesammelt).")}</p>
      ) : (
        <>
          <div className="chart-area">
            <div className="chart-yaxis">
              {[100, 75, 50, 25, 0].map((v) => (
                <span key={v} style={{ top: `${yPct(v)}%` }}>{v} %</span>
              ))}
            </div>
            <svg viewBox={`0 0 ${W} ${H}`} className="chart" preserveAspectRatio="none">
              {[25, 50, 75].map((g) => {
                const y = PAD + (1 - g / 100) * (H - 2 * PAD);
                return <line key={g} x1={PAD} y1={y} x2={W - PAD} y2={y} className="grid" />;
              })}
              <path d={path("disk")} className="line disk" />
              <path d={path("mem")} className="line mem" />
              <path d={path("cpu")} className="line cpu" />
            </svg>
          </div>
          <div className="chart-x muted small">
            <span>{fmt(points[0].ts)}</span>
            <span>{fmt(points[points.length - 1].ts)}</span>
          </div>
        </>
      )}
    </div>
  );
}
