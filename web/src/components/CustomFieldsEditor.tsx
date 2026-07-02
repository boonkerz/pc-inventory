import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "../api";
import { useI18n } from "../i18n";
import type { CustomFieldValue } from "../types";
import { useAuth } from "../auth";

type Val = string | string[];

// parseValue wandelt den gespeicherten String je Typ in den UI-Wert.
function parseValue(cv: CustomFieldValue): Val {
  if (cv.field.type === "list" || cv.field.type === "multiselect") {
    try { const a = JSON.parse(cv.value || "[]"); return Array.isArray(a) ? a.map(String) : []; }
    catch { return cv.value ? [cv.value] : []; }
  }
  return cv.value ?? "";
}

// CustomFieldsEditor zeigt und bearbeitet benutzerdefinierte Felder einer Entität.
export function CustomFieldsEditor({ model, entityId }: { model: "client" | "site" | "device"; entityId: string }) {
  const { t } = useI18n();
  const qc = useQueryClient();
  const { user } = useAuth();
  const canEdit = user?.role === "admin";
  const key = ["custom-field-values", model, entityId];
  const { data } = useQuery({ queryKey: key, queryFn: () => api.get<CustomFieldValue[]>(`/custom-field-values?model=${model}&entity_id=${entityId}`) });
  const [vals, setVals] = useState<Record<string, Val>>({});

  useEffect(() => {
    if (data) {
      const m: Record<string, Val> = {};
      for (const cv of data) m[cv.field.id] = parseValue(cv);
      setVals(m);
    }
  }, [data]);

  const save = useMutation({
    mutationFn: () => api.put("/custom-field-values", { model, entity_id: entityId, values: vals }),
    onSuccess: () => qc.invalidateQueries({ queryKey: key }),
  });

  if (!data) return <div className="muted small">{t("Lädt…")}</div>;
  if (data.length === 0) return <p className="muted">{t("Keine Felder für diese Ebene definiert (unter „Einstellungen → Benutzerdefinierte Felder“).")}</p>;

  const set = (id: string, v: Val) => setVals((s) => ({ ...s, [id]: v }));

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
      {data.map((cv) => {
        const f = cv.field;
        const v = vals[f.id] ?? (f.type === "list" || f.type === "multiselect" ? [] : "");
        return (
          <label key={f.id} className="field">
            <span className="muted small">{f.name}{f.required ? " *" : ""}</span>
            {f.type === "checkbox" ? (
              <input type="checkbox" disabled={!canEdit} checked={v === "true"} onChange={(e) => set(f.id, e.target.checked ? "true" : "false")} />
            ) : f.type === "number" ? (
              <input type="number" disabled={!canEdit} value={v as string} onChange={(e) => set(f.id, e.target.value)} />
            ) : f.type === "datetime" ? (
              <input type="datetime-local" disabled={!canEdit} value={v as string} onChange={(e) => set(f.id, e.target.value)} />
            ) : f.type === "select" ? (
              <select disabled={!canEdit} value={v as string} onChange={(e) => set(f.id, e.target.value)}>
                <option value="">—</option>
                {f.options.map((o) => <option key={o} value={o}>{o}</option>)}
              </select>
            ) : f.type === "multiselect" ? (
              <div className="chip-row">
                {f.options.map((o) => {
                  const arr = v as string[];
                  return (
                    <label key={o} className="chip">
                      <input type="checkbox" disabled={!canEdit} checked={arr.includes(o)}
                        onChange={(e) => set(f.id, e.target.checked ? [...arr, o] : arr.filter((x) => x !== o))} /> {o}
                    </label>
                  );
                })}
              </div>
            ) : f.type === "list" ? (
              <ListInput value={v as string[]} disabled={!canEdit} onChange={(a) => set(f.id, a)} />
            ) : (
              <input type="text" disabled={!canEdit} value={v as string} onChange={(e) => set(f.id, e.target.value)} />
            )}
          </label>
        );
      })}
      {canEdit && (
        <div>
          <button className="btn primary" onClick={() => save.mutate()} disabled={save.isPending}>{t("Speichern")}</button>
          {save.isSuccess && <span className="muted small" style={{ marginLeft: 10 }}>{t("gespeichert ✓")}</span>}
        </div>
      )}
    </div>
  );
}

// ListInput: frei eingebbare Einträge (Chips zum Hinzufügen/Entfernen).
function ListInput({ value, onChange, disabled }: { value: string[]; onChange: (a: string[]) => void; disabled?: boolean }) {
  const [text, setText] = useState("");
  const add = () => { const t = text.trim(); if (t && !value.includes(t)) { onChange([...value, t]); setText(""); } };
  return (
    <div>
      <div className="chip-row">
        {value.map((x) => (
          <span key={x} className="chip">{x}{!disabled && <button className="chip-x" onClick={() => onChange(value.filter((y) => y !== x))}>×</button>}</span>
        ))}
        {value.length === 0 && <span className="muted small">—</span>}
      </div>
      {!disabled && (
        <div className="inline-form" style={{ marginTop: 6 }}>
          <input value={text} placeholder="Eintrag…" onChange={(e) => setText(e.target.value)}
            onKeyDown={(e) => { if (e.key === "Enter") { e.preventDefault(); add(); } }} />
          <button className="btn" onClick={add} disabled={!text.trim()}>+ Hinzufügen</button>
        </div>
      )}
    </div>
  );
}
