import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "../api";
import type { Script } from "../types";
import { useI18n } from "../i18n";

const PLATFORMS: [string, string][] = [["windows", "Windows"], ["linux", "Linux"], ["darwin", "macOS"]];
const PLAT_LABEL: Record<string, string> = Object.fromEntries(PLATFORMS);

export function Scripts() {
  const { t } = useI18n();
  const qc = useQueryClient();
  const { data: scripts } = useQuery({ queryKey: ["scripts"], queryFn: () => api.get<Script[]>("/scripts") });
  const [editing, setEditing] = useState<Script | "new" | null>(null);
  const invalidate = () => qc.invalidateQueries({ queryKey: ["scripts"] });
  const del = useMutation({ mutationFn: (id: string) => api.del(`/scripts/${id}`), onSuccess: invalidate });

  if (editing) {
    return (
      <div className="page wide">
        <header className="page-head">
          <div className="inline-form">
            <button className="btn ghost sm" onClick={() => setEditing(null)}>← {t("Zurück")}</button>
            <h1 style={{ margin: 0 }}>{editing === "new" ? t("Neues Skript") : t("Skript bearbeiten")}</h1>
          </div>
        </header>
        <ScriptForm script={editing === "new" ? null : editing} onClose={() => setEditing(null)} onSaved={invalidate} />
      </div>
    );
  }

  return (
    <div className="page wide">
      <header className="page-head">
        <div>
          <h1>{t("Skripte")}</h1>
          <p className="muted">{t("Wiederverwendbare Skripte für Script-Checks, Tasks und Ad-hoc-Ausführung.")}</p>
        </div>
        <button className="btn primary" onClick={() => setEditing("new")}>+ {t("Neu anlegen")}</button>
      </header>

      <section className="card">
        <table className="table">
          <thead><tr><th>{t("Name")}</th><th>Shell</th><th>{t("Plattformen")}</th><th></th></tr></thead>
          <tbody>
            {(scripts ?? []).map((s) => (
              <tr key={s.id}>
                <td className="link-strong">{s.name}{s.check_only && <span className="badge badge-unknown" style={{ marginLeft: 6 }}>{t("nur Check")}</span>}</td>
                <td><span className="badge badge-unknown">{s.shell}</span></td>
                <td className="muted small">{(s.platforms && s.platforms.length > 0) ? s.platforms.map((p) => PLAT_LABEL[p] ?? p).join("/") : t("alle")}</td>
                <td>
                  <button className="btn ghost sm" onClick={() => setEditing(s)}>{t("Bearbeiten")}</button>
                  <button className="btn ghost sm" onClick={() => del.mutate(s.id)}>{t("Löschen")}</button>
                </td>
              </tr>
            ))}
            {(scripts ?? []).length === 0 && <tr><td colSpan={4} className="empty">{t("Noch keine Skripte.")}</td></tr>}
          </tbody>
        </table>
      </section>
    </div>
  );
}

function ScriptForm({ script, onClose, onSaved }: { script: Script | null; onClose: () => void; onSaved: () => void }) {
  const { t } = useI18n();
  const id = script?.id ?? null;
  const [name, setName] = useState(script?.name ?? "");
  const [shell, setShell] = useState<"shell" | "powershell">(script?.shell ?? "shell");
  const [platforms, setPlatforms] = useState<string[]>(script?.platforms ?? []);
  const [content, setContent] = useState(script?.content ?? "");
  const [checkOnly, setCheckOnly] = useState(script?.check_only ?? false);
  const togglePlat = (p: string) => setPlatforms((a) => a.includes(p) ? a.filter((x) => x !== p) : [...a, p]);

  const save = useMutation({
    mutationFn: () => {
      const body = { name, shell, platforms, content, check_only: checkOnly };
      return id ? api.put(`/scripts/${id}`, body) : api.post("/scripts", body);
    },
    onSuccess: () => { onSaved(); onClose(); },
  });

  return (
    <section className="card">
      <form className="script-form" onSubmit={(e) => { e.preventDefault(); if (name) save.mutate(); }}>
        <div className="inline-form">
          <input style={{ flex: 1 }} placeholder={t("Skriptname")} value={name} onChange={(e) => setName(e.target.value)} />
          <select value={shell} onChange={(e) => setShell(e.target.value as "shell" | "powershell")}>
            <option value="shell">shell (Linux/macOS)</option>
            <option value="powershell">powershell (Windows)</option>
          </select>
        </div>
        <div className="chip-row" style={{ marginTop: 2 }}>
          <span className="muted small">{t("Plattformen (leer = alle der Shell):")}</span>
          {PLATFORMS.map(([k, v]) => (
            <label key={k} className="chip"><input type="checkbox" checked={platforms.includes(k)} onChange={() => togglePlat(k)} /> {v}</label>
          ))}
        </div>
        <div className="chip-row" style={{ marginTop: 2 }}>
          <label className="chip"><input type="checkbox" checked={checkOnly} onChange={(e) => setCheckOnly(e.target.checked)} /> {t("Nur für Checks")}</label>
          <span className="muted small">{t("(erscheint dann nicht unter „Ausführen“ oder in der Sammelaktion)")}</span>
        </div>
        <textarea
          className="code-input"
          placeholder={shell === "powershell" ? "# PowerShell …\nexit 0" : "#!/bin/sh\nexit 0"}
          value={content}
          onChange={(e) => setContent(e.target.value)}
          spellCheck={false}
        />
        <details className="help">
          <summary>{t("Hilfe: Platzhalter & Felder befüllen")}</summary>
          <p className="muted small">
            <strong>Platzhalter</strong> werden vor der Ausführung serverseitig ersetzt –
            mit den benutzerdefinierten Feldern des Geräts/Clients/Standorts:
          </p>
          <pre className="help-code">{`{{agent.anydeskId}}  {{client.vertrag}}  {{site.vlan}}`}</pre>
          <p className="muted small">
            <strong>Filter</strong> (Twig-Stil) werden mit <code>|</code> angehängt und verkettet –
            praktisch für Listen-Felder:
          </p>
          <pre className="help-code">{`{{ agent.domains | first }}        erstes Listenelement
{{ agent.domains | last }}         letztes Element
{{ agent.domains | nth(1) }}       Element 1 (0-basiert)
{{ agent.domains | count }}        Anzahl Elemente
{{ agent.domains | join(" ") }}    mit Trenner verbinden
{{ agent.name | upper }}           GROSSSCHREIBUNG (auch: lower, trim)
{{ agent.foo | default("n/a") }}   Fallback, falls leer`}</pre>
          <p className="muted small">
            Ohne Filter wird eine Liste komma-getrennt eingesetzt. Ein unbekanntes oder leeres
            Feld ergibt einen leeren String – im Skript abfangen, z.&nbsp;B.{" "}
            <code>{`[ -z "{{ agent.domains | first }}" ] && { echo "keine domain"; exit 1; }`}</code>.
          </p>
          <p className="muted small">
            <strong>Felder automatisch befüllen:</strong> Skript gibt JSON aus und der Task
            hat „Felder aus JSON" aktiviert. Unbekannte Felder werden automatisch angelegt
            (Typ wird erkannt). Wichtig: bei Shell die JSON in einfache Anführungszeichen setzen.
          </p>
          <pre className="help-code">{shell === "powershell"
            ? `'{"agent":{"anydeskId":"123456","online":true,"tags":["srv","prod"]}}'`
            : `echo '{"agent":{"anydeskId":"123456","online":true,"tags":["srv","prod"]}}'`}</pre>
          <button type="button" className="btn ghost sm" onClick={() => setContent((c) =>
            (c ? c + "\n" : "") + (shell === "powershell"
              ? `Write-Output '{"agent":{"anydeskId":"123456"}}'`
              : `echo '{"agent":{"anydeskId":"123456"}}'`))}>
            {t("Beispiel einfügen")}
          </button>
        </details>
        <div className="inline-form">
          <button className="btn primary" type="submit" disabled={save.isPending}>{id ? t("Speichern") : t("Anlegen")}</button>
          <button type="button" className="btn ghost" onClick={onClose}>{t("Abbrechen")}</button>
        </div>
      </form>
    </section>
  );
}
