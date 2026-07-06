import { useEffect, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, ApiError } from "../api";
import { useI18n } from "../i18n";
import type { ClientTree as Tree } from "../types";
import { Modal } from "./Modal";
import { CustomFieldsEditor } from "./CustomFieldsEditor";

export type OrgFilter =
  | { kind: "all" }
  | { kind: "unassigned" }
  | { kind: "client"; id: string }
  | { kind: "site"; id: string };

export function sameFilter(a: OrgFilter, b: OrgFilter): boolean {
  if (a.kind !== b.kind) return false;
  if ("id" in a && "id" in b) return a.id === b.id;
  return true;
}

interface Props {
  tree: Tree;
  total: number;
  selected: OrgFilter;
  onSelect: (f: OrgFilter) => void;
  isAdmin: boolean;
}

export function ClientTree({ tree, total, selected, onSelect, isAdmin }: Props) {
  const { t } = useI18n();
  const qc = useQueryClient();
  // Aufgeklappte Clients über Navigation hinweg merken.
  const [expanded, setExpanded] = useState<Set<string>>(() => {
    try {
      const s = sessionStorage.getItem("pcinv-tree-expanded");
      if (s) return new Set(JSON.parse(s) as string[]);
    } catch { /* ignore */ }
    return new Set();
  });
  useEffect(() => {
    sessionStorage.setItem("pcinv-tree-expanded", JSON.stringify([...expanded]));
  }, [expanded]);
  const refresh = () => {
    qc.invalidateQueries({ queryKey: ["clients"] });
    qc.invalidateQueries({ queryKey: ["devices"] });
  };

  const createClient = useMutation({
    mutationFn: () => {
      const name = prompt(t("Name des Clients (Firma)?"));
      return name ? api.post("/clients", { name }) : Promise.resolve();
    },
    onSuccess: refresh,
  });
  const createSite = useMutation({
    mutationFn: (clientID: string) => {
      const name = prompt(t("Name des Standorts (Site)?"));
      return name ? api.post("/sites", { client_id: clientID, name }) : Promise.resolve();
    },
    onSuccess: refresh,
  });
  const renameClient = useMutation({
    mutationFn: (v: { id: string; cur: string }) => {
      const name = prompt(t("Client umbenennen:"), v.cur);
      return name && name !== v.cur ? api.put(`/clients/${v.id}`, { name }) : Promise.resolve();
    },
    onSuccess: refresh,
  });
  const renameSite = useMutation({
    mutationFn: (v: { id: string; cur: string }) => {
      const name = prompt(t("Standort umbenennen:"), v.cur);
      return name && name !== v.cur ? api.put(`/sites/${v.id}`, { name }) : Promise.resolve();
    },
    onSuccess: refresh,
  });
  // Löschen mit Blockade: hängen noch Geräte dran (HTTP 409), erst nach Bestätigung
  // mit ?force=true fortfahren (Geräte bleiben, werden „nicht zugeordnet").
  const deleteOrg = async (kind: "clients" | "sites", id: string) => {
    try {
      await api.del(`/${kind}/${id}`);
    } catch (e) {
      if (e instanceof ApiError && e.status === 409 && e.data?.device_count) {
        const n = e.data.device_count as number;
        if (!confirm(t("Daran hängen noch {n} Gerät(e). Sie bleiben erhalten, werden aber nicht mehr zugeordnet. Trotzdem löschen?", { n }))) return;
        await api.del(`/${kind}/${id}?force=true`);
      } else {
        alert((e as Error).message);
        return;
      }
    }
    refresh();
  };

  const toggle = (id: string) =>
    setExpanded((s) => {
      const n = new Set(s);
      n.has(id) ? n.delete(id) : n.add(id);
      return n;
    });

  const isSel = (f: OrgFilter) => sameFilter(selected, f);
  const [fieldsFor, setFieldsFor] = useState<{ model: "client" | "site"; id: string; name: string } | null>(null);

  return (
    <div className="org-tree">
      <button className={`tree-row root ${isSel({ kind: "all" }) ? "sel" : ""}`} onClick={() => onSelect({ kind: "all" })}>
        <span className="tree-icon">▦</span>
        <span className="tree-label">{t("Alle Geräte")}</span>
        <span className="tree-count">{total}</span>
      </button>

      {tree.unassigned_count > 0 && (
        <button
          className={`tree-row ${isSel({ kind: "unassigned" }) ? "sel" : ""}`}
          onClick={() => onSelect({ kind: "unassigned" })}
        >
          <span className="tree-icon dim">○</span>
          <span className="tree-label dim">{t("Nicht zugeordnet")}</span>
          <span className="tree-count">{tree.unassigned_count}</span>
        </button>
      )}

      {(tree.clients ?? []).map((c) => {
        const open = expanded.has(c.id);
        return (
          <div key={c.id}>
            <div className={`tree-row group ${isSel({ kind: "client", id: c.id }) ? "sel" : ""}`}>
              <button className="tree-twisty" onClick={() => toggle(c.id)} aria-label={t("Aufklappen")}>
                {open ? "▾" : "▸"}
              </button>
              <button className="tree-main" onClick={() => onSelect({ kind: "client", id: c.id })}>
                <span className="tree-icon">▥</span>
                <span className="tree-label">{c.name}</span>
                <span className="tree-count">{c.device_count}</span>
              </button>
              {isAdmin && (
                <span className="tree-actions">
                  <button title={t("Standort hinzufügen")} onClick={() => createSite.mutate(c.id)}>+</button>
                  <button title={t("Felder")} onClick={() => setFieldsFor({ model: "client", id: c.id, name: c.name })}>⊞</button>
                  <button title={t("Umbenennen")} onClick={() => renameClient.mutate({ id: c.id, cur: c.name })}>✎</button>
                  <button title={t("Löschen")} onClick={() => confirm(t("Client „{name}“ löschen?", { name: c.name })) && deleteOrg("clients", c.id)}>×</button>
                </span>
              )}
            </div>
            {open &&
              (c.sites ?? []).map((s) => (
                <div key={s.id} className={`tree-row site ${isSel({ kind: "site", id: s.id }) ? "sel" : ""}`}>
                  <button className="tree-main" onClick={() => onSelect({ kind: "site", id: s.id })}>
                    <span className="tree-icon">▢</span>
                    <span className="tree-label">{s.name}</span>
                    <span className="tree-count">{s.device_count}</span>
                  </button>
                  {isAdmin && (
                    <span className="tree-actions">
                      <button title={t("Felder")} onClick={() => setFieldsFor({ model: "site", id: s.id, name: s.name })}>⊞</button>
                      <button title={t("Umbenennen")} onClick={() => renameSite.mutate({ id: s.id, cur: s.name })}>✎</button>
                      <button title={t("Löschen")} onClick={() => confirm(t("Standort „{name}“ löschen?", { name: s.name })) && deleteOrg("sites", s.id)}>×</button>
                    </span>
                  )}
                </div>
              ))}
          </div>
        );
      })}

      {isAdmin && (
        <button className="tree-add" onClick={() => createClient.mutate()}>+ {t("Neuer Client")}</button>
      )}

      {fieldsFor && (
        <Modal onClose={() => setFieldsFor(null)}>
          <div className="page">
            <h2>Felder – {fieldsFor.model === "client" ? "Client" : "Standort"} „{fieldsFor.name}"</h2>
            <CustomFieldsEditor model={fieldsFor.model} entityId={fieldsFor.id} />
          </div>
        </Modal>
      )}
    </div>
  );
}
