import { useState } from "react";
import { api } from "../api";
import { useI18n, gt } from "../i18n";
import type { Disk, Command } from "../types";

// Ein direkter Verzeichniseintrag mit (rekursiver) Größe – Spiegel von collect.DUEntry.
interface DUEntry {
  name: string;
  path: string;
  size: number;
  dir: boolean;
  items?: number;
  counting?: boolean;
}
interface DUResult {
  path: string;
  parent: string;
  entries: DUEntry[];
  error?: string;
}

interface NodeState {
  entries?: DUEntry[];
  loading: boolean;
  error?: string;
  expanded: boolean;
}
type NodeMap = Record<string, NodeState>;

const PALETTE = ["#4f8cff", "#34c759", "#ff9f0a", "#ff453a", "#bf5af2", "#5ac8fa", "#ffd60a", "#ff6482", "#30d158", "#a2845e"];

function fmtBytes(n: number): string {
  if (!n) return "0 B";
  const u = ["B", "KB", "MB", "GB", "TB", "PB"];
  let i = 0, v = n;
  while (v >= 1024 && i < u.length - 1) { v /= 1024; i++; }
  return `${v.toFixed(v >= 10 || i === 0 ? 0 : 1)} ${u[i]}`;
}

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

// streamScan stößt einen Scan an und pollt das (Teil-)Ergebnis. onUpdate wird bei
// jedem Zwischenstand aufgerufen, bis der Befehl „done" ist (Live-Anzeige).
async function streamScan(deviceId: string, path: string, onUpdate: (r: DUResult) => void): Promise<void> {
  const { command_id } = await api.post<{ command_id: string }>(`/devices/${deviceId}/scan-dir`, { path });
  for (let i = 0; i < 850; i++) {
    await sleep(700);
    const cmd = await api.get<Command>(`/commands/${command_id}`);
    if (cmd.output) {
      try { onUpdate(JSON.parse(cmd.output) as DUResult); } catch { /* Teil-JSON ignorieren */ }
    }
    if (cmd.status === "done") {
      if (cmd.exit_code !== 0) throw new Error(gt("Scan fehlgeschlagen"));
      return;
    }
  }
  throw new Error(gt("Zeitüberschreitung beim Scan"));
}

// TreeSizePanel zeigt links einen lazy ladenden Verzeichnisbaum, rechts ein
// Tortendiagramm der Kinder des gewählten Verzeichnisses. Gestartet wird mit der
// Laufwerksübersicht (df); Daten werden erst beim Klick gesammelt.
export function TreeSizePanel({ deviceId, disks }: { deviceId: string; disks: Disk[] }) {
  const { t } = useI18n();
  const [nodes, setNodes] = useState<NodeMap>({});
  const [selected, setSelected] = useState<string>("");

  const setNode = (path: string, patch: Partial<NodeState>) =>
    setNodes((prev) => ({ ...prev, [path]: { ...(prev[path] ?? { loading: false, expanded: false }), ...patch } }));

  const load = async (path: string) => {
    setNode(path, { loading: true, error: undefined });
    try {
      await streamScan(deviceId, path, (res) => {
        setNode(path, { entries: res.entries ?? [], error: res.error });
      });
      setNode(path, { loading: false });
    } catch (e) {
      setNode(path, { loading: false, error: (e as Error).message });
    }
  };

  // toggle öffnet/schließt einen Ordner; lädt beim ersten Öffnen, setzt Auswahl.
  const toggle = (path: string) => {
    setSelected(path);
    const cur = nodes[path];
    if (!cur || (!cur.entries && !cur.loading)) {
      setNode(path, { expanded: true });
      void load(path);
      return;
    }
    setNode(path, { expanded: !cur.expanded });
  };

  const selEntries = nodes[selected]?.entries ?? [];

  return (
    <section className="card">
      <div className="treesize">
        <div className="treesize-tree scroll-list">
          {(disks ?? []).length === 0 && <p className="muted">{t("Keine Laufwerke gemeldet.")}</p>}
          {(disks ?? []).map((d) => {
            const used = Math.max(0, d.size_bytes - d.free_bytes);
            return (
              <TreeNode
                key={d.name}
                path={d.name}
                name={d.name}
                sub={`${fmtBytes(used)} / ${fmtBytes(d.size_bytes)} · ${Math.round(d.used_percent)}%`}
                depth={0}
                nodes={nodes}
                selected={selected}
                onToggle={toggle}
              />
            );
          })}
        </div>

        <div className="treesize-chart">
          {selected === "" ? (
            <p className="muted">{t("Laufwerk oder Ordner wählen, um die Verteilung zu sehen.")}</p>
          ) : nodes[selected]?.loading ? (
            <p className="muted">{t("Wird gezählt… (kann bei großen Ordnern dauern)")}</p>
          ) : nodes[selected]?.error ? (
            <p className="muted">{t("Fehler")}: {nodes[selected]?.error}</p>
          ) : selEntries.length === 0 ? (
            <p className="muted">{t("Leerer Ordner.")}</p>
          ) : (
            <Donut entries={selEntries} title={selected} />
          )}
        </div>
      </div>
    </section>
  );
}

// TreeNode rendert eine Ordnerzeile und – wenn aufgeklappt und geladen – rekursiv
// ihre Ordner-Kinder. Den eigenen Zustand liest jeder Knoten per Pfad aus der Map.
// Dateien erscheinen nur im Tortendiagramm, nicht als Baumknoten.
function TreeNode({ path, name, sub, counting, depth, nodes, selected, onToggle }: {
  path: string; name: string; sub?: string; counting?: boolean; depth: number;
  nodes: NodeMap; selected: string; onToggle: (p: string) => void;
}) {
  const state = nodes[path];
  const indicator = state?.loading ? "⏳" : counting ? "•" : state?.expanded ? "▾" : "▸";
  return (
    <div>
      <div
        className={`treesize-row treesize-row-click${selected === path ? " treesize-row-on" : ""}${counting ? " treesize-counting" : ""}`}
        style={{ paddingLeft: 8 + depth * 16 }}
        onClick={() => onToggle(path)}
      >
        <span className="treesize-ind">{indicator}</span>
        <span className="treesize-name">{name}</span>
        {sub && <span className="muted small treesize-sub">{sub}</span>}
      </div>
      {state?.expanded && state.error && <div className="treesize-row" style={{ paddingLeft: 8 + (depth + 1) * 16 }}><span className="muted small">{state.error}</span></div>}
      {state?.expanded && (state.entries ?? []).filter((e) => e.dir).map((e) => (
        <TreeNode
          key={e.path}
          path={e.path}
          name={e.name}
          sub={e.counting ? gt("zählt…") + ` ${fmtBytes(e.size)}` : `${fmtBytes(e.size)}${e.items ? ` · ${e.items} ` + gt("Dateien(n)") : ""}`}
          counting={e.counting}
          depth={depth + 1}
          nodes={nodes}
          selected={selected}
          onToggle={onToggle}
        />
      ))}
    </div>
  );
}

// Donut zeichnet ein Tortendiagramm (CSS conic-gradient) der größten Einträge.
function Donut({ entries, title }: { entries: DUEntry[]; title: string }) {
  const total = entries.reduce((a, e) => a + e.size, 0) || 1;
  const top = entries.slice(0, 9);
  const restSize = entries.slice(9).reduce((a, e) => a + e.size, 0);
  const slices = top.map((e, i) => ({ name: e.name, size: e.size, color: PALETTE[i % PALETTE.length] }));
  if (restSize > 0) slices.push({ name: "Rest", size: restSize, color: "#6b7280" });

  let acc = 0;
  const stops = slices.map((s) => {
    const start = (acc / total) * 100;
    acc += s.size;
    const end = (acc / total) * 100;
    return `${s.color} ${start}% ${end}%`;
  }).join(", ");

  return (
    <div>
      <div className="muted small" style={{ marginBottom: 8, wordBreak: "break-all" }}>{title} · {fmtBytes(total)}</div>
      <div className="donut-wrap">
        <div className="donut" style={{ background: `conic-gradient(${stops})` }}><div className="donut-hole" /></div>
        <ul className="donut-legend">
          {slices.map((s, i) => (
            <li key={i}>
              <span className="donut-swatch" style={{ background: s.color }} />
              <span className="donut-leg-name">{s.name}</span>
              <span className="muted small">{fmtBytes(s.size)} · {Math.round((s.size / total) * 100)}%</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
